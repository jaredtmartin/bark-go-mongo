package bark

import (
	"context"
	"time"
)

func Now(ctx context.Context) time.Time {
	nowStr, ok := ctx.Value(NowKey).(string)
	if !ok {
		return time.Now()
	}
	now, err := time.Parse(time.RFC3339, nowStr)
	if err != nil {
		return time.Now()
	}
	return now
}
