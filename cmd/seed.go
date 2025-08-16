package cmd

import (
	"log/slog"

	"github.com/spf13/cobra"
	"github.com/szks-repo/usage-based-billing-sample/pkg/db"
	"github.com/szks-repo/usage-based-billing-sample/pkg/db/seed"
)

// userClientCmd represents the userClient command
var seedDbCmd = &cobra.Command{
	Use:   "seedDb",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		slog.Info("Starting seed db")

		db.MustInit()
		defer db.Close()

		seed.Exec(cmd.Context(), db.Get())
	},
}

func init() {
	rootCmd.AddCommand(seedDbCmd)
}
