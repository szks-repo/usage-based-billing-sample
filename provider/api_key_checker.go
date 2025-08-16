package provider

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	"github.com/hashicorp/golang-lru/v2/expirable"
	"github.com/szks-repo/usage-based-billing-sample/pkg/now"
)

// todo add layer
type ApiKeyChecker interface {
	Check(ctx context.Context, apiKey string) (int64, error)
}

type apiKeyChecker struct {
	dbConn       *sql.DB
	lruCache     *expirable.LRU[string, int64]
	cacheExpires time.Duration
}

func NewApiKeyChecker(
	dbConn *sql.DB,
	lruCache *expirable.LRU[string, int64],
	cacheExpires time.Duration,
) ApiKeyChecker {
	return &apiKeyChecker{
		dbConn:       dbConn,
		lruCache:     lruCache,
		cacheExpires: cacheExpires,
	}
}

func (c *apiKeyChecker) Check(ctx context.Context, apiKey string) (int64, error) {
	slog.Info("apiKeyChecker.Check", "apiKey", apiKey)

	if fromCache, ok := c.lruCache.Get(apiKey); ok {
		return fromCache, nil
	}

	now := now.FromContext(ctx)
	var accountId int64
	var expiredAt time.Time
	row := c.dbConn.QueryRowContext(ctx, `SELECT a.account_id, a.expired_at FROM active_api_key a WHERE a.api_key = ? AND a.expired_at > ?`, apiKey, now)
	if err := row.Scan(
		&accountId,
		&expiredAt,
	); err != nil {
		return 0, err
	}

	slog.Debug("query end", "accountId", accountId, "expiredAt", expiredAt)
	if c.shouldCache(now, expiredAt, c.cacheExpires) {
		c.lruCache.Add(apiKey, accountId)
	}

	return accountId, nil
}

func (c *apiKeyChecker) shouldCache(now, expriedAt time.Time, cacheExpires time.Duration) bool {
	return now.Add(cacheExpires).Before(expriedAt)
}
