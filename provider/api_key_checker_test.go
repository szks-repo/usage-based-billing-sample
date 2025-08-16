package provider

import (
	"testing"
	"time"

	"github.com/zeebo/assert"
)

func Test_apiKeyChecker_shouldCache(t *testing.T) {
	t.Parallel()

	type args struct {
		now          time.Time
		expiredAt    time.Time
		cacheExpires time.Duration
	}

	tests := []struct {
		args args
		want bool
	}{
		{
			args: args{
				now:          time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
				expiredAt:    time.Date(2025, 1, 1, 0, 20, 0, 0, time.UTC),
				cacheExpires: time.Minute * 20,
			},
			want: true,
		},
		{
			args: args{
				now:          time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
				expiredAt:    time.Date(2025, 1, 1, 0, 20, 0, 1, time.UTC),
				cacheExpires: time.Minute * 20,
			},
			want: true,
		},
		{
			args: args{
				now:          time.Date(2025, 1, 1, 0, 0, 0, 1, time.UTC),
				expiredAt:    time.Date(2025, 1, 1, 0, 20, 0, 0, time.UTC),
				cacheExpires: time.Minute * 20,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			assert.Equal(t, tt.want, new(apiKeyChecker).shouldCache(tt.args.now, tt.args.expiredAt, tt.args.cacheExpires))
		})
	}
}
