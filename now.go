package bark

import (
	"os"
	"time"
)

func Now() time.Time {
	env := os.Getenv("NOW")
	// fmt.Println("env", env)
	if env != "" {
		now, err := time.Parse(time.RFC3339, env)
		if err == nil {
			return now
		}
	}
	return time.Now()
}
