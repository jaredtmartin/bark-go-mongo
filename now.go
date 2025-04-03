package bark

import (
	"context"
	"time"
)

// Returns the current time in RFC3339 format
// Returns time set in the context if it is set
// Otherwise, returns the current time
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
