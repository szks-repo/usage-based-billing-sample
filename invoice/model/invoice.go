package model

import (
	"fmt"
	"math"
	"math/big"
	"time"

	parser "github.com/szks-repo/rat-expr-parser"
	"github.com/szks-repo/usage-based-billing-sample/pkg/take"
	"github.com/szks-repo/usage-based-billing-sample/pkg/tax"
)

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
	result, err := parser.NewRatFromString(fmt.Sprintf(
		"%d + (%s)",
		dailyUsage,
		pi.basePricePerUsage.RatString(),
	))
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
	taxRate               tax.TaxRate
	taxAmount             *big.Rat
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

func NewInvoice(
	accountId uint64,
	subscriptionId uint64,
	freeCreditBalance uint64,
	dailyUsages []*DailyApiUsage,
	taxRate tax.TaxRate,
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

	taxIncludedPriceRat := parser.NewRatFromString(fmt.Sprintf(
		"(%s) * ((%s+100)/100)",
		totalPrice.RatString(),
		taxRate,
	))

	taxIncludedPrice := uint64(math.Floor(take.Left(taxIncludedPriceRat.Float64())))
	taxAmount := parser.NewRatFromString(fmt.Sprintf("%d - (%s)", taxIncludedPrice, subtotal.RatString()))

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

func (i *Invoice) FreeCreditDiscount() uint64 {
	return i.freeCreditDiscount
}
