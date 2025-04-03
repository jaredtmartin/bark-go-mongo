package bark

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
)

const tokenLength = 16

var last_test_uuid = 0

// Uuid generates a random UUID string.
// In test mode, it generates a sequential UUID for testing purposes.
func Uuid() string {
	env := os.Getenv("ENV")

	if env == "test" {
		last_test_uuid += 1
		uuid := fmt.Sprintf("%016d", last_test_uuid)
		return uuid
	}
	b := make([]byte, tokenLength)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}
