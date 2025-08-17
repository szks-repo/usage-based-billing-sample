package cmd

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/spf13/cobra"
	"github.com/szks-repo/usage-based-billing-sample/pkg/db"
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

		var (
			awsRegion = "ap-northeast-1"
			s3Url     = "http://localhost:9000"
			s3Bucket  = "api-access-log"
			queueUrl  = "amqp://localhost:5672"
		)

		ctx := cmd.Context()
		nctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
		defer stop()

		db.MustInit()
		defer db.Close()

		mqConn, err := rabbitmq.NewConn(queueUrl)
		if err != nil {
			slog.Error("Failed to connect to RabbitMQ", "error", err)
			return
		}
		defer mqConn.Close()

		cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(awsRegion))
		if err != nil {
			panic(err)
		}
		s3Client := s3.NewFromConfig(cfg, func(o *s3.Options) {
			o.BaseEndpoint = &s3Url
			if strings.Contains(s3Url, "localhost") {
				o.UsePathStyle = true
			}
		})

		worker := worker.NewWorker(
			mqConn,
			worker.NewAccessLogRecorder(
				s3Client,
				s3Bucket,
				1024<<10*5,
				time.Second*30,
				db.Get(),
			),
		)
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
