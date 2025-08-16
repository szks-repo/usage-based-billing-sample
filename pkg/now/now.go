package now

import (
	"context"
	"time"
)

type nowKey struct{}

func FromContext(ctx context.Context) time.Time {
	if t, ok := ctx.Value(nowKey{}).(time.Time); ok {
		return t
	}
	return time.Now()
}

func WithContext(ctx context.Context, t time.Time) context.Context {
	return context.WithValue(ctx, nowKey{}, t)
}
