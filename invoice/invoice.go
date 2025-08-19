package invoice

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	"github.com/szks-repo/usage-based-billing-sample/pkg/db/dto"
	"github.com/szks-repo/usage-based-billing-sample/pkg/now"
	"github.com/szks-repo/usage-based-billing-sample/pkg/pipeline"
)

type InvoiceMaker struct {
	dbConn     *sql.DB
	reconciler UsageReconciler
}

func NewInvoiceMaker(
	dbConn *sql.DB,
	reconciler UsageReconciler,
) *InvoiceMaker {
	return &InvoiceMaker{
		dbConn:     dbConn,
		reconciler: reconciler,
	}
}

func (i *InvoiceMaker) CreateInvoiceDaily(ctx context.Context) {
	baseDate := now.FromContext(ctx).AddDate(0, 0, -1)

	accountIds, err := i.listAccountIds(ctx, baseDate)
	if err != nil {
		slog.Error("Failed to listAccountIds", "error", err)
	}

	slog.Info("target accountIds", "accountIds", accountIds)
	if len(accountIds) == 0 {
		return
	}

	pipeline.Pipeline3(
		ctx,
		pipeline.From(accountIds),
		pipeline.ForEach[int64](func(accountId int64) {
			i.reconciler.Do(ctx, baseDate, accountId)
		}),
		pipeline.Map[int64, *dto.Invoice](func(accountId int64) (*dto.Invoice, error) {
			return i.createInvoice(ctx, accountId)
		}),
		pipeline.ForEach[*dto.Invoice](func(invoice *dto.Invoice) {
			i.publishNotifyQueue(ctx, invoice)
		}),
	)
}

func (i *InvoiceMaker) createInvoice(ctx context.Context, accountId int64) (*dto.Invoice, error) {
	// todo
	return nil, nil
}

func (i *InvoiceMaker) publishNotifyQueue(ctx context.Context, invoice *dto.Invoice) { /* todo */ }

func (i *InvoiceMaker) listAccountIds(ctx context.Context, t time.Time) ([]int64, error) {
	cutoff := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)

	rows, err := i.dbConn.QueryContext(ctx, `SELECT a.id FROM account a JOIN subscription s ON a.id = s.account_id WHERE s.estimated_to = ?`, cutoff)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accountIds []int64
	for rows.Next() {
		var accountId int64
		if err := rows.Scan(&accountId); err != nil {
			return nil, err
		}
		accountIds = append(accountIds, accountId)
	}

	return accountIds, nil
}
