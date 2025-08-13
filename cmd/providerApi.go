package cmd

import (
	"github.com/spf13/cobra"

	"github.com/szks-repo/usage-based-billing-sample/provider"
)

// providerApiCmd represents the providerApi command
var providerApiCmd = &cobra.Command{
	Use:   "providerApi",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		srv := provider.NewApiServer(
			"todo",
			":8080",
		)
		srv.ListenAndServe()
	},
}

func init() {
	rootCmd.AddCommand(providerApiCmd)
}
