package taosx

import (
	"context"
	"strings"
	"testing"
)

func TestHealthReportsHealthyWhenDriverHealthy(t *testing.T) {
	client, err := New(context.Background(), validConfig(), WithDriver(&recordingDriver{}))
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	status := client.Health(context.Background())
	if status.Status != HealthHealthy {
		t.Fatalf("unexpected status: %#v", status)
	}
	if status.DriverMode != DriverModeWebSocket || status.Database != "metrics" {
		t.Fatalf("health status missing config context: %#v", status)
	}
}

func TestHealthRedactsDriverErrors(t *testing.T) {
	client, err := New(context.Background(), validConfig(), WithDriver(&recordingDriver{healthErr: errDriverHealth}))
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	status := client.Health(context.Background())
	if status.Status != HealthDegraded {
		t.Fatalf("unexpected status: %#v", status)
	}
	if strings.Contains(status.Message, "taosdata") {
		t.Fatalf("health message leaked secret: %q", status.Message)
	}
}

func TestHealthReportsUnhealthyForClosedClient(t *testing.T) {
	client, err := New(context.Background(), validConfig(), WithDriver(&recordingDriver{}))
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	if err := client.Close(context.Background()); err != nil {
		t.Fatalf("close: %v", err)
	}

	status := client.Health(context.Background())
	if status.Status != HealthUnhealthy {
		t.Fatalf("unexpected status: %#v", status)
	}
}
