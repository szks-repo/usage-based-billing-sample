package cmd

import (
	"log/slog"

	"github.com/spf13/cobra"

	"github.com/szks-repo/usage-based-billing-sample/pkg/rabbitmq"
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
		slog.Info("Starting provider API server")

		mqConn, err := rabbitmq.NewConn("amqp://localhost:5672")
		defer mqConn.Close()
		if err != nil {
			slog.Error("Failed to connect to RabbitMQ", "error", err)
			return
		}

		srv := provider.NewApiServer(mqConn, ":8080")
		if err := srv.ListenAndServe(); err != nil {
			slog.Error("Failed to start provider API server", "error", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(providerApiCmd)
}
