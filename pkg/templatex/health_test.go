package templatex

import (
	"context"
	"testing"
)

func TestHealthCheckHealthy(t *testing.T) {
	client, err := New(context.Background(), Config{Name: "templatex"})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	status := client.HealthCheck(context.Background())
	if status.Status != HealthHealthy {
		t.Fatalf("expected healthy status, got %q", status.Status)
	}
	if status.Name != "templatex" {
		t.Fatalf("expected templatex health name, got %q", status.Name)
	}
}

func TestHealthCheckClosedClientUnhealthy(t *testing.T) {
	client, err := New(context.Background(), Config{Name: "templatex"})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	if err := client.Close(context.Background()); err != nil {
		t.Fatalf("close client: %v", err)
	}

	status := client.HealthCheck(context.Background())
	if status.Status != HealthUnhealthy {
		t.Fatalf("expected unhealthy status, got %q", status.Status)
	}
}
