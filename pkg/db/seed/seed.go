package seed

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"time"

	"github.com/szks-repo/usage-based-billing-sample/pkg/db/dto"
)

func Exec(ctx context.Context, dbConn *sql.DB) {
	slog.Info("Seeding satrt")

	accounts := []*dto.Account{
		{
			ID:          1,
			AccountName: "test-account-1",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          2,
			AccountName: "test-account-2",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}
	for _, item := range accounts {
		if _, err := dto.AccountByID(ctx, dbConn, item.ID); err != nil && !errors.Is(err, sql.ErrNoRows) {
			continue
		}
		if err := item.Insert(ctx, dbConn); err != nil {
			slog.Warn("failed to insert account", "error", err)
		}
	}

	for _, item := range []*dto.ActiveAPIKey{
		{
			AccountID: accounts[0].ID,
			APIKey:    "test-api-key-1",
			ExpiredAt: time.Now().AddDate(1, 0, 0),
			CreatedAt: time.Now(),
		},
		{
			AccountID: accounts[1].ID,
			APIKey:    "test-api-key-2",
			ExpiredAt: time.Now().AddDate(1, 0, 0),
			CreatedAt: time.Now(),
		},
	} {
		if _, err := dto.ActiveAPIKeyByAPIKey(ctx, dbConn, item.APIKey); err != nil && !errors.Is(err, sql.ErrNoRows) {
			slog.Warn("failed to ActiveAPIKeyByAPIKey active_api_key", "error", err)
			continue
		}
		if err := item.Insert(ctx, dbConn); err != nil {
			slog.Warn("failed to insert active_api_key", "error", err)
		}
	}

	slog.Info("Seeding complete")
}
