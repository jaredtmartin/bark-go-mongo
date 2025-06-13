package bark

import (
	"log"
	"os"
	"testing"
)

func TestUuidInTestEnv(t *testing.T) {
	// Set the environment to "test"
	resetTestUuid()
	os.Setenv("ENV", "test")
	defer os.Unsetenv("ENV")

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
	log.Println("Generated UUID:", uuid)
	// Check that the UUID is not empty
	if uuid == "" {
		t.Error("Expected a non-empty UUID, got an empty string")
	}

	// Check that the UUID is of the correct length
	if len(uuid) != tokenLength {
		t.Errorf("Expected UUID length to be %d, got %d", tokenLength, len(uuid))
	}
}
func TestUuidWithLength(t *testing.T) {
	// Set the environment to "test"
	os.Setenv("ENV", "test")
	defer os.Unsetenv("ENV")
	resetTestUuid()
	testLen := 8
	uuid := Uuid(testLen)
	if uuid != "00000001" {
		t.Errorf("Expected uuid to be '00000001', got '%s'", uuid)
	}
	os.Setenv("ENV", "production")
	uuid = Uuid(testLen)
	if len(uuid) != testLen {
		t.Errorf("Expected UUID length to be %d, got %d", testLen, len(uuid))
	}
	// Odd lengths will be rounded down
	testLen = 5
	uuid = Uuid(testLen)
	if len(uuid) != testLen-1 {
		t.Errorf("Expected UUID length to be %d, got %d", testLen-1, len(uuid))
	}
}
