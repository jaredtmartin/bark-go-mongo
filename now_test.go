package bark_test

import (
	"context"
	"testing"

	"github.com/jaredtmartin/bark-go-mongo"
)

func TestNowMode(t *testing.T) {
	ctx := context.Background()
	now := "2022-03-27T19:48:27.43Z"
	expected := "2022-03-27 19:48:27.43 +0000 UTC"

	// should return current time if now is not set
	result := bark.Now(ctx).String()
	if result == now {
		t.Errorf("Expected Now() to return current time, got %s", result)
	}
	// should return test time if now is set
	ctx = context.WithValue(ctx, bark.NowKey, now)
	result = bark.Now(ctx).String()
	if result != expected {
		t.Errorf("Expected Now() to return %s, got %s", expected, result)
	}
}
