package bark

import (
	"os"
	"time"
)

func Now() time.Time {
	env := os.Getenv("ENV")
	if env == "test" {
		now, err := time.Parse(time.RFC3339, os.Getenv("NOW"))
		if err == nil {
			return now
		}
	}
	return time.Now()
}
