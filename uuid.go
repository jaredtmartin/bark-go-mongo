package bark

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/gofiber/fiber/v2"
)

const tokenLength = 16

var last_test_uuid = 0

func Uuid(c *fiber.Ctx) string {
	env := c.Locals("env")

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
