package model

import (
	"errors"
	"fmt"
	"math"
	"math/big"
	"time"

	"github.com/samber/lo"
	parser "github.com/szks-repo/rat-expr-parser"

	"github.com/szks-repo/usage-based-billing-sample/pkg/take"
	"github.com/szks-repo/usage-based-billing-sample/pkg/tax"
)

type (
	PriceTable struct {
		item                          *PriceTableItem
		additionalRangePricesPerUsage RangePrices
	}

	PriceTableItem struct {
		applyStartedAt    time.Time
		basePricePerUsage *big.Rat
	}

	RangePrice struct {
		minUsage int
		maxUsage int
		price    *big.Rat
	}

	RangePrices []*RangePrice
)

func NewPriceTable(rangePrices RangePrices) *PriceTable {
	return &PriceTable{
		item: &PriceTableItem{
			applyStartedAt:    time.Time{},
			basePricePerUsage: take.Left(new(big.Rat).SetString("0.001")),
		},
		additionalRangePricesPerUsage: rangePrices,
	}
}

type RangePriceBuilder struct {
	items []*RangePrice
	errs  []error
}

func (b *RangePriceBuilder) Set(minUsage, maxUsage int, pricePerUsage string) {
	rat, err := parser.NewRatFromString(pricePerUsage)
	if err != nil {
		b.errs = append(b.errs, err)
		return
	}
	b.items = append(b.items, &RangePrice{
		minUsage: minUsage,
		maxUsage: maxUsage,
		price:    rat,
	})
}

func (b *RangePriceBuilder) Build() (RangePrices, error) {
	if len(b.errs) > 0 {
		return nil, errors.Join(b.errs...)
	}
	return b.items, nil
}

type CalculateResult struct {
	Subtotal        *big.Rat
	TotalPrice      *big.Rat
	TotalUsage      uint64
	FreeCreditUsage uint64
}

func (pt *PriceTable) MustCalculate(dailyUsages []*DailyApiUsage, freeCredit uint64) *CalculateResult {
	totalUsage := lo.SumBy(dailyUsages, func(du *DailyApiUsage) uint64 {
		return du.Usage()
	})

	freeCreditUsage := int64(freeCredit)
	totalUsageAfterCerditApplied := int64(totalUsage) - int64(freeCredit)
	if totalUsageAfterCerditApplied < 0 {
		totalUsageAfterCerditApplied = 0
		freeCreditUsage += totalUsageAfterCerditApplied
	}

	subtotal := new(big.Rat).SetInt64(0)
	if totalUsageAfterCerditApplied > 0 {
		var err error
		subtotal, err = parser.NewRatFromString(fmt.Sprintf(
			"%d * (%s)",
			totalUsageAfterCerditApplied,
			pt.item.basePricePerUsage.RatString(),
		))
		if err != nil {
			panic(err)
		}

		// for _, additional := range pt.additionalRangePricesPerUsage {
		// }
	}

	return &CalculateResult{
		Subtotal:        subtotal,
		TotalPrice:      subtotal,
		TotalUsage:      totalUsage,
		FreeCreditUsage: uint64(freeCreditUsage),
	}
}

type DailyApiUsage struct {
	date  time.Time
	usage uint64
}

func NewDailyApiUsage(date time.Time, usage uint64) *DailyApiUsage {
	return &DailyApiUsage{
		date:  date,
		usage: usage,
	}
}

func (du *DailyApiUsage) Date() time.Time {
	return du.date
}

func (du *DailyApiUsage) Usage() uint64 {
	return du.usage
}

type Invoice struct {
	totalUsage            uint64
	freeCreditUsage       uint64
	subtotal              *big.Rat
	totalPrice            *big.Rat
	taxIncludedTotalPrice uint64
	taxRate               tax.TaxRate
	taxAmount             *big.Rat
}

func NewInvoice(
	accountId uint64,
	subscriptionId uint64,
	freeCreditBalance uint64,
	dailyUsages []*DailyApiUsage,
	taxRate tax.TaxRate,
	priceTable *PriceTable,
) *Invoice {
	result := priceTable.MustCalculate(dailyUsages, freeCreditBalance)

	taxIncludedPriceRat := take.Left(parser.NewRatFromString(fmt.Sprintf(
		"(%s) * ((%s+100)/100)",
		result.TotalPrice.RatString(),
		taxRate,
	)))

	taxIncludedPrice := uint64(math.Floor(take.Left(taxIncludedPriceRat.Float64())))
	taxAmount := take.Left(parser.NewRatFromString(fmt.Sprintf(
		"%d - (%s)",
		taxIncludedPrice,
		result.TotalPrice.RatString(),
	)))

	return &Invoice{
		totalUsage:            result.TotalUsage,
		freeCreditUsage:       result.FreeCreditUsage,
		subtotal:              result.Subtotal,
		totalPrice:            result.TotalPrice,
		taxRate:               taxRate,
		taxIncludedTotalPrice: taxIncludedPrice,
		taxAmount:             taxAmount,
	}
}

func (i *Invoice) TotalUsage() uint64 {
	return i.totalUsage
}

func (i *Invoice) FreeCreditUsage() uint64 {
	return i.freeCreditUsage
}

func (i *Invoice) TotalPriceString() string {
	return i.totalPrice.FloatString(5)
}

func (i *Invoice) TaxIncludedTotalPrice() uint64 {
	return i.taxIncludedTotalPrice
}

func (i *Invoice) TaxRate() tax.TaxRate {
	return i.taxRate
}

func (i *Invoice) TaxAmountString() string {
	return i.taxAmount.FloatString(5)
}

func (i *Invoice) SubtotalString() string {
	return i.subtotal.FloatString(5)
}
