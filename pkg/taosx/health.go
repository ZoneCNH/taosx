package taosx

import (
	"context"
	"time"
)

type HealthState string

const (
	HealthHealthy   HealthState = "healthy"
	HealthDegraded  HealthState = "degraded"
	HealthUnhealthy HealthState = "unhealthy"
)

type HealthStatus struct {
	Name       string
	Status     HealthState
	CheckedAt  time.Time
	LatencyMs  int64
	DriverMode DriverMode
	Database   string
	Message    string
	Metadata   map[string]string
	Details    map[string]string
}

func (c *client) Health(ctx context.Context) HealthStatus {
	start := c.clock()
	status := HealthStatus{
		Name:       c.cfg.Name,
		Status:     HealthHealthy,
		CheckedAt:  start,
		DriverMode: c.cfg.DriverMode,
		Database:   c.cfg.Database,
		Details: map[string]string{
			"endpoint": c.cfg.Endpoint,
		},
	}
	if ctx == nil {
		status.Status = HealthUnhealthy
		status.Message = "context is required"
		status.LatencyMs = c.clock().Sub(start).Milliseconds()
		return status
	}
	if err := ctx.Err(); err != nil {
		status.Status = HealthUnhealthy
		status.Message = err.Error()
		status.LatencyMs = c.clock().Sub(start).Milliseconds()
		return status
	}
	if c.isClosed() {
		status.Status = HealthUnhealthy
		status.Message = "client is closed"
		status.LatencyMs = c.clock().Sub(start).Milliseconds()
		return status
	}
	if err := c.driver.Health(ctx); err != nil {
		status.Status = HealthDegraded
		status.Message = redact(err.Error())
	}
	status.LatencyMs = c.clock().Sub(start).Milliseconds()
	c.metrics.SetGauge(MetricClientHealthStatus, healthValue(status.Status), map[string]string{"name": c.cfg.Name, "driver_mode": string(c.cfg.DriverMode)})
	c.metrics.ObserveHistogram(MetricClientHealthLatencyMS, float64(status.LatencyMs), map[string]string{"name": c.cfg.Name})
	return status
}

func healthValue(state HealthState) float64 {
	switch state {
	case HealthHealthy:
		return 1
	case HealthDegraded:
		return 0.5
	default:
		return 0
	}
}
