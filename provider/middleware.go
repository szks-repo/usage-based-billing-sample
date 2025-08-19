package provider

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/streadway/amqp"

	httplib "github.com/szks-repo/usage-based-billing-sample/pkg/http"
	"github.com/szks-repo/usage-based-billing-sample/pkg/now"
	"github.com/szks-repo/usage-based-billing-sample/pkg/rabbitmq"
	"github.com/szks-repo/usage-based-billing-sample/pkg/types"
	"github.com/szks-repo/usage-based-billing-sample/pkg/types/ctxkey"
)

type Middleware interface {
	Wrap(next http.Handler) http.Handler
}

type middleware struct {
	apiKeyChecker ApiKeyChecker
	mqConn        *rabbitmq.Conn
	queue         amqp.Queue
}

func NewMiddleware(
	apiKeyChecker ApiKeyChecker,
	mqConn *rabbitmq.Conn,
	queue amqp.Queue,
) Middleware {
	return &middleware{
		apiKeyChecker: apiKeyChecker,
		mqConn:        mqConn,
		queue:         queue,
	}
}

func (mw *middleware) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := now.FromContext(r.Context())

		apiKey := r.Header.Get("x-api-key")

		accountId, err := mw.apiKeyChecker.Check(r.Context(), apiKey)
		if err != nil {
			slog.Info("Invalid api key", "apiKey", apiKey, "error", err)
			http.Error(w, "Unauthorized: missing or invalid api key", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), ctxkey.ApiKey{}, apiKey)
		ctx = context.WithValue(ctx, ctxkey.AccountId{}, accountId)

		w2 := httplib.NewResponseWriterWrapper(w)

		next.ServeHTTP(w2, r.WithContext(ctx))
		slog.Info("End main handler", "path", r.URL.Path, "satusCode", w2.StatusCode())
		if w2.StatusCode() < 200 || w2.StatusCode() >= 300 {
			return
		}

		ts := now.FromContext(ctx)
		payload, err := json.Marshal(&types.ApiAccessLog{
			AccountId:  accountId,
			Timestamp:  ts,
			ClientIP:   r.RemoteAddr,
			Path:       r.URL.Path,
			Method:     r.Method,
			StatusCode: w2.StatusCode(),
			Latency:    int64(time.Since(start)),
			UserAgent:  r.UserAgent(),
		})
		if err != nil {
			slog.Error("Failed to json.Marshal", "error", err)
			return
		}

		if err := backoff.Retry(func() error {
			return mw.mqConn.Channel.Publish(
				"",            // exchange
				mw.queue.Name, // routing key
				false,         // mandatory
				false,         // immediate
				amqp.Publishing{
					ContentType:  "application/json",
					Body:         payload,
					DeliveryMode: amqp.Persistent,
					Headers: map[string]any{
						"timestamp": ts,
					},
					Timestamp: ts,
				},
			)
		}, backoff.WithMaxRetries(backoff.NewExponentialBackOff(), 5)); err != nil {
			slog.Error("Failed to publish message to RabbitMQ", "error", err)
			return
		}
	})
}
