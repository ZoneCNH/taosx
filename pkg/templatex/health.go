package templatex

import (
	"context"
	"time"
)

type HealthStatusValue string

const (
	HealthHealthy   HealthStatusValue = "healthy"
	HealthDegraded  HealthStatusValue = "degraded"
	HealthUnhealthy HealthStatusValue = "unhealthy"
)

type HealthStatus struct {
	Name      string
	Status    HealthStatusValue
	Message   string
	CheckedAt time.Time
	LatencyMs int64
	Metadata  map[string]string
}

func (c *Client) HealthCheck(ctx context.Context) HealthStatus {
	start := time.Now()

	if err := ctx.Err(); err != nil {
		return HealthStatus{
			Name:      "templatex",
			Status:    HealthUnhealthy,
			Message:   err.Error(),
			CheckedAt: time.Now(),
			LatencyMs: time.Since(start).Milliseconds(),
		}
	}

	c.mu.Lock()
	closed := c.closed
	c.mu.Unlock()

	if closed {
		return HealthStatus{
			Name:      "templatex",
			Status:    HealthUnhealthy,
			Message:   "client is closed",
			CheckedAt: time.Now(),
			LatencyMs: time.Since(start).Milliseconds(),
		}
	}

	return HealthStatus{
		Name:      "templatex",
		Status:    HealthHealthy,
		Message:   "ok",
		CheckedAt: time.Now(),
		LatencyMs: time.Since(start).Milliseconds(),
	}
}
