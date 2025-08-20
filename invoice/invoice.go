package invoice

import (
	"context"
	"database/sql"
	"errors"
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

func (i *InvoiceMaker) createInvoice(ctx context.Context, subscription *dto.Subscription) (*dto.Invoice, error) {
	// todo IF
	row := i.dbConn.QueryRowContext(ctx, "SELECT SUM(`usage`) FROM daily_api_usage WHERE account_id = ? AND date >= ? AND date <= ?",
		subscription.AccountID,
		subscription.From.Format("20060102"),
		subscription.EstimatedTo.Format("20060102"),
	)
	var total int64
	if err := row.Scan(&total); err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	invoiceDto := &dto.Invoice{
		AccountID:          subscription.AccountID,
		SubscriptionID:     subscription.ID,
		TotalUsage:         uint(total),
		FreeCreditDiscount: 0, //todo
		Subtotal:           0, //todo
		TotalPrice:         0, //todo
	}
	if err := invoiceDto.Insert(ctx, i.dbConn); err != nil {
		return nil, err
	}

	return invoiceDto, nil
}

func (i *InvoiceMaker) publishNotifyQueue(ctx context.Context, invoice *dto.Invoice) { /* todo */ }

func (i *InvoiceMaker) listSubscriptions(ctx context.Context, t time.Time) ([]*dto.Subscription, error) {
	cutoff := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)

	query := "SELECT s.id, s.account_id, s.from, s.estimated_to " +
		"FROM account a JOIN subscription s ON a.id = s.account_id " +
		"WHERE s.estimated_to = ?"
	rows, err := i.dbConn.QueryContext(ctx, query, cutoff)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subscriptions []*dto.Subscription
	for rows.Next() {
		var dst dto.Subscription
		if err := rows.Scan(
			&dst.ID,
			&dst.AccountID,
			&dst.From,
			&dst.EstimatedTo,
		); err != nil {
			return nil, err
		}
		subscriptions = append(subscriptions, &dst)
	}

	return subscriptions, nil
}
