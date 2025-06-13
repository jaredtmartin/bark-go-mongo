package bark

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"os"
)

const tokenLength = 16

var last_test_uuid = 0

func resetTestUuid() {
	last_test_uuid = 0
}

// Uuid generates a random UUID string.
// In test mode, it generates a sequential UUID for testing purposes.
// length is optional and will be rounded down to the nearest even number.
func Uuid(length ...int) string {
	env := os.Getenv("ENV")
	l := tokenLength
	if len(length) > 0 {
		l = length[0]
	}
	if env == "test" {
		last_test_uuid += 1
		uuid := fmt.Sprintf("%0*d", l, last_test_uuid)
		return uuid
	}
	log.Println("Generating UUID in production mode of length", l)
	b := make([]byte, l/2) // Divide by 2 because hex encoding doubles the length
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}
