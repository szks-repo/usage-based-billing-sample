package invoice

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/szks-repo/gopipeline"
	parser "github.com/szks-repo/rat-expr-parser"

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

	var priceTable PriceTable

	gopipeline.New3(
		ctx,
		gopipeline.From(subscriptions),
		gopipeline.ForEach(func(subscription *dto.Subscription) {
			i.reconciler.Do(ctx, baseDate, subscription)
		}),
		gopipeline.Map(func(subscription *dto.Subscription) (*Invoice, error) {
			return i.createInvoice(ctx, subscription, priceTable)
		}),
		gopipeline.ForEach(func(invoice *Invoice) {
			i.publishNotifyQueue(ctx, invoice)
		}),
	)
}

func (i *InvoiceMaker) listSubscriptionDailyApiUsages(ctx context.Context, subscription *dto.Subscription) ([]*DailyApiUsage, error) {
	rows, err := i.dbConn.QueryContext(ctx, "SELECT `date`, `usage` FROM daily_api_usage WHERE account_id = ? AND date >= ? AND date <= ?",
		subscription.AccountID,
		subscription.From.Format("20060102"),
		subscription.EstimatedTo.Format("20060102"),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*DailyApiUsage
	for rows.Next() {
		var dst DailyApiUsage
		if err := rows.Scan(
			&dst.date,
			&dst.usage,
		); err != nil {
			return nil, err
		}
		result = append(result, &dst)
	}

	return result, nil
}

func (i *InvoiceMaker) getFreeCreditBalanceByAccountId(ctx context.Context, accountId uint64) (uint64, error) {
	row := i.dbConn.QueryRowContext(
		ctx,
		"SELECT balance FROM free_credit_balance_snapshot WHERE account_id = ? ORDER BY created_at DESC LIMIT 1",
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
	priceTable PriceTable,
) (*Invoice, error) {
	dailyUsages, err := i.listSubscriptionDailyApiUsages(ctx, subscription)
	if err != nil {
		return nil, err
	}

	freeCredit, err := i.getFreeCreditBalanceByAccountId(ctx, subscription.AccountID)
	if err != nil {
		return nil, err
	}

	invoice := NewInvoice(
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
		invoice.TaxRate(),
		invoice.TaxAmountString(),
		invoice.TotalPriceString(),
		uint(invoice.TaxIncludedTotalPrice()),
	); err != nil {
		return nil, err
	}

	return invoice, nil
}

func (i *InvoiceMaker) publishNotifyQueue(ctx context.Context, invoice *Invoice) { /* todo */ }

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

type (
	PriceTableItem struct {
		applyStartedAt                time.Time
		basePricePerUsage             *big.Rat
		additionalRangePricesPerUsage RangePrices
	}
	PriceTable []*PriceTableItem

	RangePrice struct {
		minUsage int
		maxUsage int
		price    *big.Rat
	}

	RangePrices []*RangePrice
)

func (pt PriceTable) GetIsShouldApplyDate(date time.Time) *PriceTableItem {
	for _, item := range pt {
		if item.applyStartedAt.Equal(date) || item.applyStartedAt.After(date) {
			return item
		}
	}
	return pt[len(pt)-1]
}

func (pi *PriceTableItem) MustCalculate(dailyUsage uint64) *big.Rat {
	result, err := parser.NewFromString(strings.Join([]string{
		strconv.FormatUint(dailyUsage, 64),
		"*",
		"(" + pi.basePricePerUsage.RatString() + ")",
	}, ""))
	if err != nil {
		panic(err)
	}

	return result
}

type Invoice struct {
	totalUsage            uint64
	subtotal              *big.Rat
	freeCreditDiscount    uint64
	totalPrice            *big.Rat
	taxIncludedTotalPrice uint64
	taxRate               uint8
	taxAmount             *big.Rat
}

type DailyApiUsage struct {
	date  time.Time
	usage uint64
}

func (du *DailyApiUsage) Date() time.Time {
	return du.date
}

func (du *DailyApiUsage) Usage() uint64 {
	return du.usage
}

func NewInvoice(
	accountId uint64,
	subscriptionId uint64,
	freeCreditBalance uint64,
	dailyUsages []*DailyApiUsage,
	taxRate uint8,
	priceTable PriceTable,
) *Invoice {

	subtotal := new(big.Rat)
	var totalUsage uint64
	for _, du := range dailyUsages {
		totalUsage += du.Usage()
		item := priceTable.GetIsShouldApplyDate(du.Date())
		subtotal = subtotal.Add(subtotal, item.MustCalculate(du.Usage()))
	}

	// if freeCreditBalance > 0 {
	// 	f64, _ := rat.Float64()
	// 	freeCreditBalance = min(freeCreditBalance, uint64(math.Floor(f64)))
	// }

	totalPrice := subtotal

	taxIncludedPriceRat := parser.NewFromString(strings.Join([]string{
		totalPrice.RatString(),
		"*",
		fmt.Sprintf("((%d+100)/100)", taxRate),
	}, ""))
	f64, _ := taxIncludedPriceRat.Float64()
	taxIncludedPrice := uint64(math.Floor(f64))
	taxAmount := parser.NewFromString(fmt.Sprintf("%d - (%s)", taxIncludedPrice, subtotal.RatString()))

	return &Invoice{
		totalUsage:            totalUsage,
		subtotal:              subtotal,
		totalPrice:            totalPrice,
		taxRate:               taxRate,
		taxIncludedTotalPrice: taxIncludedPrice,
		taxAmount:             taxAmount,
	}
}

func (i *Invoice) TotalUsage() uint64 {
	return i.totalUsage
}

func (i *Invoice) TotalPriceString() string {

}

func (i *Invoice) TaxIncludedTotalPrice() uint64 {
	return i.taxIncludedTotalPrice
}

func (i *Invoice) TaxRate() uint8 {
	return i.taxRate
}

func (i *Invoice) TaxAmountString() string {
	return i.taxAmount.FloatString(5)
}

func (i *Invoice) SubtotalString() string {
	return i.subtotal.FloatString(5)
}

func (i *Invoice) FreeCreditDiscount() uint64 {
	return i.freeCreditDiscount
}
