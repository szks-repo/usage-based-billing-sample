package cmd

import (
	"log/slog"

	"github.com/spf13/cobra"
	"github.com/szks-repo/usage-based-billing-sample/invoice"
	"github.com/szks-repo/usage-based-billing-sample/pkg/db"
)

// providerApiCmd represents the providerApi command
var createDailyInvoiceCmd = &cobra.Command{
	Use:   "createDailyInvoice",
	Short: "create daily invoice",
	Run: func(cmd *cobra.Command, args []string) {
		slog.Info("Starting daily invoice maker")

		ctx := cmd.Context()

		db.MustInit()
		defer db.Close()

		maker := invoice.NewInvoiceMaker(
			db.Get(),
			invoice.NewUsageReconciler(),
		)
		maker.CreateInvoiceDaily(ctx)
	},
}

func init() {
	rootCmd.AddCommand(createDailyInvoiceCmd)
}
