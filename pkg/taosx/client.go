package taosx

import (
	"context"
	"errors"
	"sync"
	"time"
)

type Client interface {
	Exec(context.Context, Statement) (ExecResult, error)
	Query(context.Context, Query) (Rows, error)
	WriteBatch(context.Context, Batch) (WriteResult, error)
	SchemalessWrite(context.Context, SchemalessPayload) (WriteResult, error)
	Health(context.Context) HealthStatus
	Close(context.Context) error
}

type Driver interface {
	Exec(context.Context, Statement) (ExecResult, error)
	Query(context.Context, Query) (Rows, error)
	WriteBatch(context.Context, Batch) (WriteResult, error)
	SchemalessWrite(context.Context, SchemalessPayload) (WriteResult, error)
	Health(context.Context) error
	Close(context.Context) error
}

type client struct {
	cfg     Config
	driver  Driver
	metrics Metrics
	clock   func() time.Time
	mu      sync.Mutex
	closed  bool
}

func New(ctx context.Context, cfg Config, opts ...Option) (Client, error) {
	const op = "taosx.New"
	options := defaultOptions()
	for _, opt := range opts {
		opt(&options)
	}
	if ctx == nil {
		err := validationError(op, "context is required", nil)
		recordErrorMetric(options.metrics, "new", err)
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		wrapped := contextError(op, err)
		recordErrorMetric(options.metrics, "new", wrapped)
		return nil, wrapped
	}
	cfg = cfg.Normalize()
	if err := cfg.Validate(); err != nil {
		recordErrorMetric(options.metrics, "new", err)
		return nil, err
	}
	options.metrics.IncCounter(MetricClientCreatedTotal, map[string]string{"name": cfg.Name, "driver_mode": string(cfg.DriverMode)})
	return &client{cfg: cfg, driver: options.driver, metrics: options.metrics, clock: options.clock}, nil
}

func (c *client) Exec(ctx context.Context, stmt Statement) (ExecResult, error) {
	const op = "taosx.Exec"
	if err := c.checkReady(ctx, op); err != nil {
		return ExecResult{}, err
	}
	if err := validateSQL(op, stmt.SQL); err != nil {
		recordErrorMetric(c.metrics, "exec", err)
		return ExecResult{}, err
	}
	c.metrics.IncCounter(MetricClientRequestsTotal, map[string]string{"op": "exec", "driver_mode": string(c.cfg.DriverMode)})
	result, err := c.driver.Exec(ctx, stmt)
	if err != nil {
		wrapped := normalizeDriverError(ErrorKindSQL, op, err)
		recordErrorMetric(c.metrics, "exec", wrapped)
		return ExecResult{}, wrapped
	}
	return result, nil
}

func (c *client) Query(ctx context.Context, query Query) (Rows, error) {
	const op = "taosx.Query"
	if err := c.checkReady(ctx, op); err != nil {
		return nil, err
	}
	if err := validateSQL(op, query.SQL); err != nil {
		recordErrorMetric(c.metrics, "query", err)
		return nil, err
	}
	c.metrics.IncCounter(MetricClientRequestsTotal, map[string]string{"op": "query", "driver_mode": string(c.cfg.DriverMode)})
	rows, err := c.driver.Query(ctx, query)
	if err != nil {
		wrapped := normalizeDriverError(ErrorKindSQL, op, err)
		recordErrorMetric(c.metrics, "query", wrapped)
		return nil, wrapped
	}
	return rows, nil
}

func (c *client) WriteBatch(ctx context.Context, batch Batch) (WriteResult, error) {
	const op = "taosx.WriteBatch"
	if err := c.checkReady(ctx, op); err != nil {
		return WriteResult{}, err
	}
	if err := batch.Validate(); err != nil {
		recordErrorMetric(c.metrics, "write_batch", err)
		return WriteResult{}, err
	}
	c.metrics.IncCounter(MetricClientRequestsTotal, map[string]string{"op": "write_batch", "driver_mode": string(c.cfg.DriverMode)})
	for range batch.Points {
		c.metrics.IncCounter(MetricClientBatchRowsTotal, map[string]string{"database": batch.Database})
	}
	result, err := c.driver.WriteBatch(ctx, batch)
	if err != nil {
		wrapped := normalizeDriverError(ErrorKindWrite, op, err)
		recordErrorMetric(c.metrics, "write_batch", wrapped)
		result = markPartialWrite(result, int64(len(batch.Points)))
		return result, wrapped
	}
	return result, nil
}

func (c *client) SchemalessWrite(ctx context.Context, payload SchemalessPayload) (WriteResult, error) {
	const op = "taosx.SchemalessWrite"
	if err := c.checkReady(ctx, op); err != nil {
		return WriteResult{}, err
	}
	if err := payload.Validate(); err != nil {
		recordErrorMetric(c.metrics, "schemaless_write", err)
		return WriteResult{}, err
	}
	c.metrics.IncCounter(MetricClientRequestsTotal, map[string]string{"op": "schemaless_write", "driver_mode": string(c.cfg.DriverMode)})
	for range payload.Lines {
		c.metrics.IncCounter(MetricClientSchemalessLinesTotal, map[string]string{"protocol": string(payload.Protocol)})
	}
	result, err := c.driver.SchemalessWrite(ctx, payload)
	if err != nil {
		wrapped := normalizeDriverError(ErrorKindWrite, op, err)
		recordErrorMetric(c.metrics, "schemaless_write", wrapped)
		result = markPartialWrite(result, int64(len(payload.Lines)))
		return result, wrapped
	}
	return result, nil
}

func markPartialWrite(result WriteResult, attempted int64) WriteResult {
	if result.RowsAttempted == 0 {
		result.RowsAttempted = attempted
	}
	result.Partial = true
	return result
}

func (c *client) Close(ctx context.Context) error {
	const op = "taosx.Close"
	if c == nil {
		return validationError(op, "client is nil", nil)
	}
	if ctx == nil {
		err := validationError(op, "context is required", nil)
		recordErrorMetric(c.metrics, "close", err)
		return err
	}
	if err := ctx.Err(); err != nil {
		wrapped := contextError(op, err)
		recordErrorMetric(c.metrics, "close", wrapped)
		return wrapped
	}
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return nil
	}
	c.closed = true
	driver := c.driver
	name := c.cfg.Name
	c.mu.Unlock()
	if err := driver.Close(ctx); err != nil {
		wrapped := normalizeDriverError(ErrorKindConnection, op, err)
		recordErrorMetric(c.metrics, "close", wrapped)
		return wrapped
	}
	c.metrics.IncCounter(MetricClientClosedTotal, map[string]string{"name": name})
	return nil
}

func (c *client) checkReady(ctx context.Context, op string) error {
	if c == nil {
		return validationError(op, "client is nil", nil)
	}
	if ctx == nil {
		err := validationError(op, "context is required", nil)
		recordErrorMetric(c.metrics, op, err)
		return err
	}
	if err := ctx.Err(); err != nil {
		wrapped := contextError(op, err)
		recordErrorMetric(c.metrics, op, wrapped)
		return wrapped
	}
	if c.isClosed() {
		err := driverError(ErrorKindConnection, op, "client is closed", false, nil)
		recordErrorMetric(c.metrics, op, err)
		return err
	}
	return nil
}

func (c *client) isClosed() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.closed
}

func recordErrorMetric(metrics Metrics, op string, err error) {
	if metrics == nil {
		return
	}
	metrics.IncCounter(MetricClientErrorsTotal, map[string]string{
		"op":   op,
		"kind": string(errorKind(err)),
	})
}

func normalizeDriverError(kind ErrorKind, op string, err error) error {
	var target *Error
	if errors.As(err, &target) {
		return target
	}
	return driverError(kind, op, err.Error(), kind == ErrorKindConnection || kind == ErrorKindUnavailable || kind == ErrorKindTimeout, err)
}

type unavailableDriver struct{}

func (unavailableDriver) Exec(context.Context, Statement) (ExecResult, error) {
	return ExecResult{}, driverError(ErrorKindUnavailable, "driver.Exec", "TDengine driver is not configured", true, nil)
}

func (unavailableDriver) Query(context.Context, Query) (Rows, error) {
	return nil, driverError(ErrorKindUnavailable, "driver.Query", "TDengine driver is not configured", true, nil)
}

func (unavailableDriver) WriteBatch(context.Context, Batch) (WriteResult, error) {
	return WriteResult{}, driverError(ErrorKindUnavailable, "driver.WriteBatch", "TDengine driver is not configured", true, nil)
}

func (unavailableDriver) SchemalessWrite(context.Context, SchemalessPayload) (WriteResult, error) {
	return WriteResult{}, driverError(ErrorKindUnavailable, "driver.SchemalessWrite", "TDengine driver is not configured", true, nil)
}

func (unavailableDriver) Health(context.Context) error {
	return driverError(ErrorKindUnavailable, "driver.Health", "TDengine driver is not configured", true, nil)
}

func (unavailableDriver) Close(context.Context) error {
	return nil
}
