package templatex

import (
	"context"
	"testing"
	"time"
)

func TestNewRejectsInvalidConfig(t *testing.T) {
	_, err := New(context.Background(), Config{Timeout: time.Second})
	if err == nil {
		t.Fatal("expected invalid config to fail")
	}
}

func TestNewRejectsCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := New(ctx, Config{Name: "templatex"})
	if err == nil {
		t.Fatal("expected canceled context to fail")
	}
}

func TestCloseIsIdempotent(t *testing.T) {
	client, err := New(context.Background(), Config{Name: "templatex"})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	if err := client.Close(context.Background()); err != nil {
		t.Fatalf("first close: %v", err)
	}
	if err := client.Close(context.Background()); err != nil {
		t.Fatalf("second close: %v", err)
	}
}
