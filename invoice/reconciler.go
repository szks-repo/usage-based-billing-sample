package invoice

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/szks-repo/usage-based-billing-sample/pkg/db/dto"
)

type UsageReconciler interface {
	Do(ctx context.Context, baseDate time.Time, subscription *dto.Subscription) error
}

func NewUsageReconciler() UsageReconciler {
	return &usageReconciler{}
}

type usageReconciler struct {
	s3Client *s3.Client
}

func (rr *usageReconciler) Do(ctx context.Context, baseDate time.Time, subscription *dto.Subscription) error {
	// todo
	return nil
}
