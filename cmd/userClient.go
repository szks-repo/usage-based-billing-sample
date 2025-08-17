package cmd

import (
	"io"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"sync"
	"time"

	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"github.com/szks-repo/usage-based-billing-sample/pkg/db"
	"github.com/szks-repo/usage-based-billing-sample/pkg/now"
)

// userClientCmd represents the userClient command
var userClientCmd = &cobra.Command{
	Use:   "userClient",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		slog.Info("Starting user client")

		ctx := cmd.Context()

		db.MustInit()
		defer db.Close()

		dbConn := db.Get()

		rows, err := dbConn.QueryContext(ctx, `SELECT api_key FROM active_api_key WHERE expired_at > ?`, now.FromContext(ctx))
		if err != nil {
			panic(err)
		}
		defer rows.Close()

		var apiKeys []string
		for rows.Next() {
			var s string
			if err := rows.Scan(&s); err != nil {
				panic(err)
			}
			apiKeys = append(apiKeys, s)
		}

		if len(apiKeys) == 0 {
			slog.Info("Active api keys not found")
			return
		}

		var (
			apiUrl     = "http://localhost:8080/api/v1/one"
			httpClient = &http.Client{
				Transport: &http.Transport{
					MaxIdleConnsPerHost: 30,
					IdleConnTimeout:     time.Minute,
				},
			}
		)

		var wg sync.WaitGroup
		for _, key := range apiKeys {
			wg.Go(func() {
				slog.Info("Start api user", "apiKey", key)
				for {
					select {
					case <-ctx.Done():
						return
					default:
					}

					req := lo.Must(http.NewRequestWithContext(ctx, http.MethodGet, apiUrl, nil))
					req.Header.Set("x-api-key", key)
					res, err := httpClient.Do(req)
					if err != nil {
						slog.Info("Failed to *http.Client.Do", "error", err)
						return
					}
					defer res.Body.Close()

					if res.StatusCode != 200 {
						slog.Warn("Http request status not 200", "statusCode", res.StatusCode)
					} else {
						io.Copy(io.Discard, res.Body)
					}

					time.Sleep(time.Microsecond * time.Duration(rand.IntN(500)))
				}
			})
		}
		wg.Wait()
	},
}

func init() {
	rootCmd.AddCommand(userClientCmd)
}
