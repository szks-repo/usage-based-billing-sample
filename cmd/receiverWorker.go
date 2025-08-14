package cmd

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

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

		ctx := cmd.Context()
		nctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
		defer stop()

		mqConn, err := rabbitmq.NewConn("amqp://localhost:5672")
		if err != nil {
			slog.Error("Failed to connect to RabbitMQ", "error", err)
			return
		}
		defer mqConn.Close()

		worker := worker.NewWorker(mqConn)
		go worker.Run(ctx)

		<-nctx.Done()
		slog.Info("Received shutdown signal, stopping worker")

		ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
		defer cancel()
		slog.Info("Worker stopped gracefully")
	},
}

func init() {
	rootCmd.AddCommand(receiverWorkerCmd)
}
