package cmd

import (
	"log/slog"

	"github.com/spf13/cobra"
	"github.com/szks-repo/usage-based-billing-sample/pkg/rabbitmq"
	"github.com/szks-repo/usage-based-billing-sample/worker"
)

// receiverWorkerCmd represents the receiverWorker command
var receiverWorkerCmd = &cobra.Command{
	Use:   "receiverWorker",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		slog.Info("Starting receiver worker")

		mqConn, err := rabbitmq.NewConn("amqp://localhost:5672")
		if err != nil {
			slog.Error("Failed to connect to RabbitMQ", "error", err)
			return
		}
		defer mqConn.Close()

		worker := worker.NewWorker(mqConn)
		worker.Run(cmd.Context())
	},
}

func init() {
	rootCmd.AddCommand(receiverWorkerCmd)
}
