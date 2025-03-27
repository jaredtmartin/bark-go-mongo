package bark

import (
	"os"
	"testing"
)

func TestUuidInTestEnv(t *testing.T) {
	// Set the environment to "test"
	os.Setenv("ENV", "test")
	defer os.Unsetenv("ENV")

	// Reset the global variable for testing
	last_test_uuid = 0

	// Generate UUIDs and check their format
	uuid1 := Uuid()
	uuid2 := Uuid()

	if uuid1 != "0000000000000001" {
		t.Errorf("Expected uuid1 to be '0000000000000001', got '%s'", uuid1)
	}

	if uuid2 != "0000000000000002" {
		t.Errorf("Expected uuid2 to be '0000000000000002', got '%s'", uuid2)
	}
}

func TestUuidInNonTestEnv(t *testing.T) {
	// Set the environment to something other than "test"
	os.Setenv("ENV", "production")
	defer os.Unsetenv("ENV")

	uuid := Uuid()

	// Check that the UUID is not empty
	if uuid == "" {
		t.Error("Expected a non-empty UUID, got an empty string")
	}

	// Check that the UUID is of the correct length
	if len(uuid) != tokenLength*2 {
		t.Errorf("Expected UUID length to be %d, got %d", tokenLength*2, len(uuid))
	}
}
