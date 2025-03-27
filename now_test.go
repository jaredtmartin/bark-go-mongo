package bark

import (
	"testing"
)

func TestNowInTestMode(t *testing.T) {
	if Now().String() == "2025-03-27 19:48:27.43 +0000 UTC" {
		t.Errorf("Expected Now() to return current time, got %s", Now().String())
	}
	t.Setenv("NOW", "2025-03-27T19:48:27.43Z")
	if Now().String() != "2025-03-27 19:48:27.43 +0000 UTC" {
		t.Errorf("Expected Now() to return 2025-03-27 19:48:27.43 +0000 UTC, got %s", Now().String())
	}

}
