package invoice

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"time"

	"github.com/szks-repo/gopipeline"

	"github.com/szks-repo/usage-based-billing-sample/invoice/model"
	"github.com/szks-repo/usage-based-billing-sample/pkg/db/dto"
	"github.com/szks-repo/usage-based-billing-sample/pkg/now"
	"github.com/szks-repo/usage-based-billing-sample/pkg/tax"
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

	subscriptions, err := i.listSubscriptions(ctx, baseDate)
	if err != nil {
		slog.Error("Failed to listAccountIds", "error", err)
	}

	slog.Info("target subscriptions", "len", len(subscriptions))
	if len(subscriptions) == 0 {
		return
	}

	gopipeline.New3(
		ctx,
		gopipeline.From(subscriptions),
		gopipeline.ForEach(func(subscription *dto.Subscription) {
			i.reconciler.Do(ctx, baseDate, subscription)
		}),
		gopipeline.Map(func(subscription *dto.Subscription) (*model.Invoice, error) {
			return i.createInvoice(ctx, subscription)
		}),
		gopipeline.ForEach(func(invoice *model.Invoice) {
			i.publishNotifyQueue(ctx, invoice)
		}),
	)
}

func (i *InvoiceMaker) getPriceTable(ctx context.Context, accountId uint64) (*model.PriceTable, error) {
	query := "SELECT `min_usage`, `max_usage`, `price_per_usage` FROM account_price_table " +
		"WHERE account_id = ? " +
		"ORDER BY min_usage ASC"
	rows, err := i.dbConn.QueryContext(ctx, query, accountId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var builder model.RangePriceBuilder
	for rows.Next() {
		var minUsage int
		var maxUsage int
		var pricePerUsage string
		if err := rows.Scan(
			&minUsage,
			&maxUsage,
			&pricePerUsage,
		); err != nil {
			return nil, err
		}

		builder.Set(minUsage, maxUsage, pricePerUsage)
	}

	rangePrices, err := builder.Build()
	if err != nil {
		return nil, err
	}

	return model.NewPriceTable(rangePrices), nil
}

func (i *InvoiceMaker) listSubscriptionDailyApiUsages(ctx context.Context, subscription *dto.Subscription) ([]*model.DailyApiUsage, error) {
	rows, err := i.dbConn.QueryContext(ctx, "SELECT `date`, `usage` FROM daily_api_usage WHERE account_id = ? AND date >= ? AND date <= ?",
		subscription.AccountID,
		subscription.From.Format("20060102"),
		subscription.EstimatedTo.Format("20060102"),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*model.DailyApiUsage
	for rows.Next() {
		var date time.Time
		var usage uint64
		if err := rows.Scan(
			&date,
			&usage,
		); err != nil {
			return nil, err
		}
		result = append(result, model.NewDailyApiUsage(date, usage))
	}

	return result, nil
}

func (i *InvoiceMaker) getFreeCreditBalanceByAccountId(ctx context.Context, accountId uint64) (uint64, error) {
	row := i.dbConn.QueryRowContext(
		ctx,
		"SELECT balance FROM account_free_credit_balance_snapshot WHERE account_id = ? ORDER BY created_at DESC LIMIT 1",
		accountId,
	)
	var balance uint64
	if err := row.Scan(&balance); err != nil && errors.Is(err, sql.ErrNoRows) {
		return 0, err
	}

	return balance, nil
}

func (i *InvoiceMaker) createInvoice(
	ctx context.Context,
	subscription *dto.Subscription,
) (*model.Invoice, error) {
	dailyUsages, err := i.listSubscriptionDailyApiUsages(ctx, subscription)
	if err != nil {
		return nil, err
	}

	freeCredit, err := i.getFreeCreditBalanceByAccountId(ctx, subscription.AccountID)
	if err != nil {
		return nil, err
	}

	priceTable, err := i.getPriceTable(ctx, subscription.AccountID)
	if err != nil {
		return nil, err
	}

	invoice := model.NewInvoice(
		subscription.AccountID,
		subscription.ID,
		freeCredit,
		dailyUsages,
		tax.DefaultTaxRate,
		priceTable,
	)

	query := "INSERT INTO invoice " +
		"(account_id, subscription_id, total_usage, free_credit_discount, subtotal, tax_rate, tax_amount, total_price, tax_included_total_price) " +
		"VALUES (?,?,?,?,?,?,?,?,?)"

	if _, err := i.dbConn.ExecContext(
		ctx,
		query,
		subscription.AccountID,
		subscription.ID,
		uint(invoice.TotalUsage()),
		uint(invoice.FreeCreditDiscount()),
		invoice.SubtotalString(),
		invoice.TaxRate().Uint8(),
		invoice.TaxAmountString(),
		invoice.TotalPriceString(),
		uint(invoice.TaxIncludedTotalPrice()),
	); err != nil {
		return nil, err
	}

	return invoice, nil
}

func (i *InvoiceMaker) publishNotifyQueue(ctx context.Context, invoice *model.Invoice) { /* todo */ }

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
