package invoice

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type UsageReconciler interface {
	Do(ctx context.Context, baseDate time.Time, accountId int64) error
}

func NewUsageReconciler() UsageReconciler {
	return &usageReconciler{}
}

type usageReconciler struct {
	s3Client *s3.Client
}

func (rr *usageReconciler) Do(ctx context.Context, baseDate time.Time, accountId int64) error {
	// todo
	return nil
}
