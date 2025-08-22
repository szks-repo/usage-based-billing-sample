package model

import (
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/szks-repo/usage-based-billing-sample/pkg/take"
	"github.com/szks-repo/usage-based-billing-sample/pkg/tax"
)

func TestNewInvoice(t *testing.T) {
	t.Parallel()

	type args struct {
		dailyUsages       []*DailyApiUsage
		freeCreditBalance uint64
		priceTable        *PriceTable
	}

	tests := []struct {
		args args
		want *Invoice
	}{
		{
			args: args{
				dailyUsages: []*DailyApiUsage{
					{
						date:  time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
						usage: 10000,
					},
					{
						date:  time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC),
						usage: 10000,
					},
				},
				freeCreditBalance: 0,
				priceTable:        NewPriceTable(nil),
			},
			want: &Invoice{
				totalUsage:            20000,
				freeCreditUsage:       0,
				subtotal:              take.Left(new(big.Rat).SetString("20.00000")),
				totalPrice:            take.Left(new(big.Rat).SetString("20.00000")),
				taxIncludedTotalPrice: 22,
				taxRate:               tax.DefaultTaxRate,
				taxAmount:             take.Left(new(big.Rat).SetString("2.00000")),
			},
		},
		{
			args: args{
				dailyUsages: []*DailyApiUsage{
					{
						date:  time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
						usage: 100000,
					},
					{
						date:  time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC),
						usage: 100000,
					},
				},
				freeCreditBalance: 0,
				priceTable:        NewPriceTable(nil),
			},
			want: &Invoice{
				totalUsage:            200000,
				freeCreditUsage:       0,
				subtotal:              take.Left(new(big.Rat).SetString("200.00000")),
				totalPrice:            take.Left(new(big.Rat).SetString("200.00000")),
				taxIncludedTotalPrice: 220,
				taxRate:               tax.DefaultTaxRate,
				taxAmount:             take.Left(new(big.Rat).SetString("20.00000")),
			},
		},
		{
			args: args{
				dailyUsages: []*DailyApiUsage{
					{
						date:  time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
						usage: 100000,
					},
					{
						date:  time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC),
						usage: 100000,
					},
					{
						date:  time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC),
						usage: 100000,
					},
				},
				freeCreditBalance: 100000,
				priceTable:        NewPriceTable(nil),
			},
			want: &Invoice{
				totalUsage:            300000,
				freeCreditUsage:       100000,
				subtotal:              take.Left(new(big.Rat).SetString("200.00000")),
				totalPrice:            take.Left(new(big.Rat).SetString("200.00000")),
				taxIncludedTotalPrice: 220,
				taxRate:               tax.DefaultTaxRate,
				taxAmount:             take.Left(new(big.Rat).SetString("20.00000")),
			},
		},
		{
			args: args{
				dailyUsages: []*DailyApiUsage{
					{
						date:  time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
						usage: 123450,
					},
					{
						date:  time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC),
						usage: 133333,
					},
				},
				freeCreditBalance: 0,
				priceTable:        NewPriceTable(nil),
			},
			want: &Invoice{
				totalUsage:            256783,
				freeCreditUsage:       0,
				subtotal:              take.Left(new(big.Rat).SetString("256.78300")),
				totalPrice:            take.Left(new(big.Rat).SetString("256.78300")),
				taxIncludedTotalPrice: 282,
				taxRate:               tax.DefaultTaxRate,
				taxAmount:             take.Left(new(big.Rat).SetString("25.21700")),
			},
		},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := NewInvoice(
				1,
				1,
				tt.args.freeCreditBalance,
				tt.args.dailyUsages,
				tax.DefaultTaxRate,
				tt.args.priceTable,
			)
			if !assert.Equal(t, tt.want, got) {
				t.Log(got.TotalUsage())
				t.Log(got.SubtotalString())
				t.Log(got.TotalPriceString())
				t.Log(got.TaxAmountString())
				t.Log(got.TaxIncludedTotalPrice())
			}
		})
	}
}
