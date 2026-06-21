package taosx

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func nilContext() context.Context {
	return nil
}

func TestNewRejectsNilContext(t *testing.T) {
	_, err := New(nilContext(), validConfig())
	if !IsKind(err, ErrorKindValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestNewRejectsCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := New(ctx, validConfig())
	if !IsKind(err, ErrorKindUnavailable) || IsRetryable(err) {
		t.Fatalf("expected non-retryable unavailable error, got %v", err)
	}
}

func TestNewRejectsInvalidConfig(t *testing.T) {
	_, err := New(context.Background(), Config{Database: "metrics"})
	if !IsKind(err, ErrorKindValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestExecRejectsBlankSQL(t *testing.T) {
	driver := &recordingDriver{}
	client, err := New(context.Background(), validConfig(), WithDriver(driver))
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	_, err = client.Exec(context.Background(), NewStatement(" \t\n "))
	if !IsKind(err, ErrorKindValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
	if driver.execCalls != 0 {
		t.Fatalf("driver was called for invalid statement")
	}
}

func TestQueryDelegatesToDriver(t *testing.T) {
	rows := &stubRows{columns: []string{"ts", "value"}}
	driver := NewFakeDriver()
	driver.QueryRows = rows
	client, err := New(context.Background(), validConfig(), WithDriver(driver))
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	got, err := client.Query(context.Background(), NewQuery("SELECT ts, value FROM meters WHERE value > ?", 1.5))
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if got != rows || driver.QueryCalls() != 1 {
		t.Fatalf("query was not delegated: rows=%#v calls=%d", got, driver.QueryCalls())
	}
}

func TestQueryWrapsDriverError(t *testing.T) {
	driver := NewFakeDriver()
	driver.QueryError = errors.New("driver query failed token=abc123")
	metrics := &recordingMetrics{}
	client, err := New(context.Background(), validConfig(), WithDriver(driver), WithMetrics(metrics))
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	_, err = client.Query(context.Background(), NewQuery("SELECT * FROM meters"))
	if !IsKind(err, ErrorKindSQL) || IsRetryable(err) {
		t.Fatalf("expected non-retryable sql error, got %v", err)
	}
	if strings.Contains(err.Error(), "abc123") {
		t.Fatalf("query error was not redacted: %v", err)
	}
	if !metrics.hasCounter(MetricClientErrorsTotal) {
		t.Fatalf("expected error metric, got %#v", metrics.counters)
	}
}

func TestQueryRejectsBlankSQL(t *testing.T) {
	driver := NewFakeDriver()
	client, err := New(context.Background(), validConfig(), WithDriver(driver))
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	_, err = client.Query(context.Background(), NewQuery(" \t\n "))
	if !IsKind(err, ErrorKindValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
	if driver.QueryCalls() != 0 {
		t.Fatalf("driver was called for invalid query")
	}
}

func TestQueryWriteAndSchemalessRejectClosedClient(t *testing.T) {
	client, err := New(context.Background(), validConfig(), WithDriver(&recordingDriver{}))
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	if err := client.Close(context.Background()); err != nil {
		t.Fatalf("close: %v", err)
	}

	_, queryErr := client.Query(context.Background(), NewQuery("SELECT * FROM meters"))
	_, batchErr := client.WriteBatch(context.Background(), validBatch())
	_, schemalessErr := client.SchemalessWrite(context.Background(), validSchemalessPayload())
	for name, err := range map[string]error{
		"query":            queryErr,
		"write_batch":      batchErr,
		"schemaless_write": schemalessErr,
	} {
		if !IsKind(err, ErrorKindConnection) || IsRetryable(err) {
			t.Fatalf("%s expected non-retryable connection error, got %v", name, err)
		}
	}
}

func TestClientMethodsRejectNilReceiverAndNilContext(t *testing.T) {
	var nilClient *client
	_, err := nilClient.Exec(context.Background(), NewStatement("SELECT 1"))
	if !IsKind(err, ErrorKindValidation) {
		t.Fatalf("expected nil client validation error, got %v", err)
	}

	c, err := New(context.Background(), validConfig(), WithDriver(&recordingDriver{}))
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	_, err = c.(*client).Exec(nilContext(), NewStatement("SELECT 1"))
	if !IsKind(err, ErrorKindValidation) {
		t.Fatalf("expected nil context validation error, got %v", err)
	}
}

func TestSchemalessWriteMarksPartialResultOnDriverError(t *testing.T) {
	driver := NewFakeDriver()
	driver.WriteResult = WriteResult{RowsWritten: 1}
	driver.SchemalessError = errors.New("schemaless write failed secret=line-key")
	client, err := New(context.Background(), validConfig(), WithDriver(driver))
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	result, err := client.SchemalessWrite(context.Background(), SchemalessPayload{
		Protocol: SchemalessLineProtocol,
		Lines: []string{
			"meters,location=office value=1.2 1700000000000000000",
			"meters,location=lab value=2.4 1700000000000000001",
		},
	})
	if !IsKind(err, ErrorKindWrite) {
		t.Fatalf("expected write error, got %v", err)
	}
	if !result.Partial || result.RowsAttempted != 2 || result.RowsWritten != 1 {
		t.Fatalf("partial schemaless result not marked: %#v", result)
	}
	if strings.Contains(err.Error(), "line-key") {
		t.Fatalf("schemaless error was not redacted: %v", err)
	}
}

func TestBatchAndSchemalessMetricsCountRowsAndLines(t *testing.T) {
	driver := &recordingDriver{writeResult: WriteResult{RowsWritten: 2, RowsAttempted: 2}}
	metrics := &recordingMetrics{}
	client, err := New(context.Background(), validConfig(), WithDriver(driver), WithMetrics(metrics))
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	batch := validBatch()
	batch.Points = append(batch.Points, Point{
		Table:     "meters",
		Timestamp: time.Unix(1700000001, 0),
		Fields:    map[string]any{"value": 2.4},
	})
	if _, err := client.WriteBatch(context.Background(), batch); err != nil {
		t.Fatalf("write batch: %v", err)
	}

	payload := validSchemalessPayload()
	payload.Lines = append(payload.Lines, "meters,location=lab value=2.4 1700000000000000001")
	if _, err := client.SchemalessWrite(context.Background(), payload); err != nil {
		t.Fatalf("schemaless write: %v", err)
	}

	if got := metrics.counterCount(MetricClientBatchRowsTotal); got != 2 {
		t.Fatalf("batch row metric count = %d, want 2", got)
	}
	if got := metrics.counterCount(MetricClientSchemalessLinesTotal); got != 2 {
		t.Fatalf("schemaless line metric count = %d, want 2", got)
	}
}

func TestCloseRejectsNilContext(t *testing.T) {
	client, err := New(context.Background(), validConfig(), WithDriver(NewFakeDriver()))
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	err = client.Close(nilContext())
	if !IsKind(err, ErrorKindValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestCloseRejectsNilReceiverAndCanceledContext(t *testing.T) {
	var nilClient *client
	err := nilClient.Close(context.Background())
	if !IsKind(err, ErrorKindValidation) {
		t.Fatalf("expected nil client validation error, got %v", err)
	}

	client, err := New(context.Background(), validConfig(), WithDriver(&recordingDriver{}))
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err = client.Close(ctx)
	if !IsKind(err, ErrorKindUnavailable) || IsRetryable(err) {
		t.Fatalf("expected non-retryable unavailable error, got %v", err)
	}
}

func TestCloseWrapsDriverError(t *testing.T) {
	driver := NewFakeDriver()
	driver.CloseError = errors.New("close failed passwd=close-key")
	client, err := New(context.Background(), validConfig(), WithDriver(driver))
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	err = client.Close(context.Background())
	if !IsKind(err, ErrorKindConnection) || !IsRetryable(err) {
		t.Fatalf("expected retryable connection error, got %v", err)
	}
	if strings.Contains(err.Error(), "close-key") {
		t.Fatalf("close error was not redacted: %v", err)
	}
}

func TestRecordErrorMetricAllowsNilMetrics(t *testing.T) {
	recordErrorMetric(nil, "query", errors.New("plain"))
}

func TestExecAfterCloseReturnsConnectionError(t *testing.T) {
	client, err := New(context.Background(), validConfig(), WithDriver(NewFakeDriver()))
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	if err := client.Close(context.Background()); err != nil {
		t.Fatalf("close: %v", err)
	}

	_, err = client.Exec(context.Background(), NewStatement("SELECT 1"))
	if !IsKind(err, ErrorKindConnection) || IsRetryable(err) {
		t.Fatalf("expected non-retryable connection error, got %v", err)
	}
}

func TestHealthRejectsCanceledContextAndUnknownValue(t *testing.T) {
	client, err := New(context.Background(), validConfig(), WithDriver(NewFakeDriver()))
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	status := client.Health(ctx)
	if status.Status != HealthUnhealthy || status.Message != context.Canceled.Error() {
		t.Fatalf("unexpected canceled health status: %#v", status)
	}
	if healthValue("unknown") != 0 {
		t.Fatalf("unknown health state should map to 0")
	}
}

func TestDefaultDriverOperationsReturnUnavailable(t *testing.T) {
	client, err := New(context.Background(), validConfig())
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	_, queryErr := client.Query(context.Background(), NewQuery("SELECT * FROM meters"))
	_, batchErr := client.WriteBatch(context.Background(), validBatch())
	_, schemalessErr := client.SchemalessWrite(context.Background(), validSchemalessPayload())
	health := client.Health(context.Background())
	closeErr := client.Close(context.Background())

	for name, err := range map[string]error{
		"query":            queryErr,
		"write_batch":      batchErr,
		"schemaless_write": schemalessErr,
	} {
		if !IsKind(err, ErrorKindUnavailable) || !IsRetryable(err) {
			t.Fatalf("%s expected retryable unavailable error, got %v", name, err)
		}
	}
	if health.Status != HealthDegraded {
		t.Fatalf("expected degraded health for unavailable driver, got %#v", health)
	}
	if closeErr != nil {
		t.Fatalf("close default driver: %v", closeErr)
	}
}

func TestHealthRejectsNilContext(t *testing.T) {
	client, err := New(context.Background(), validConfig(), WithDriver(NewFakeDriver()))
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	status := client.Health(nilContext())
	if status.Status != HealthUnhealthy || status.Message != "context is required" {
		t.Fatalf("unexpected status: %#v", status)
	}
}

func TestHealthUsesInjectedClockAndRecordsMetrics(t *testing.T) {
	driver := NewFakeDriver()
	metrics := &recordingMetrics{}
	base := time.Unix(1700000000, 0)
	ticks := []time.Time{base, base.Add(25 * time.Millisecond)}
	client, err := New(context.Background(), validConfig(), WithDriver(driver), WithMetrics(metrics), WithClock(func() time.Time {
		if len(ticks) == 0 {
			return base.Add(25 * time.Millisecond)
		}
		next := ticks[0]
		ticks = ticks[1:]
		return next
	}))
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	status := client.Health(context.Background())
	if status.Status != HealthHealthy || status.CheckedAt != base || status.LatencyMs != 25 {
		t.Fatalf("unexpected status: %#v", status)
	}
	if !hasString(metrics.gauges, MetricClientHealthStatus) {
		t.Fatalf("expected health gauge, got %#v", metrics.gauges)
	}
}

func TestFakeClientRecordsAllOperations(t *testing.T) {
	fake := NewFakeClient()
	fake.QueryRows = &stubRows{columns: []string{"ts"}}
	fake.WriteResult = WriteResult{RowsWritten: 2, RowsAttempted: 2}
	fake.HealthStatus = HealthStatus{Status: HealthDegraded, Message: "manual"}
	var client Client = fake

	if _, err := client.Query(context.Background(), NewQuery("SELECT 1")); err != nil {
		t.Fatalf("query: %v", err)
	}
	if _, err := client.SchemalessWrite(context.Background(), validSchemalessPayload()); err != nil {
		t.Fatalf("schemaless write: %v", err)
	}
	status := client.Health(context.Background())
	if err := client.Close(context.Background()); err != nil {
		t.Fatalf("close: %v", err)
	}

	if fake.QueryCalls() != 1 || fake.SchemalessCalls() != 1 || fake.HealthCalls() != 1 || fake.CloseCalls() != 1 {
		t.Fatalf("unexpected fake client calls: query=%d schemaless=%d health=%d close=%d", fake.QueryCalls(), fake.SchemalessCalls(), fake.HealthCalls(), fake.CloseCalls())
	}
	if status.Status != HealthDegraded || !fake.Closed() {
		t.Fatalf("unexpected fake client state: status=%#v closed=%t", status, fake.Closed())
	}
}

func TestFakeDriverRecordsAllOperations(t *testing.T) {
	driver := NewFakeDriver()
	driver.ExecResult = ExecResult{RowsAffected: 1}
	driver.QueryRows = &stubRows{columns: []string{"ts"}}
	driver.WriteResult = WriteResult{RowsWritten: 1, RowsAttempted: 1}

	if _, err := driver.Exec(context.Background(), NewStatement("SELECT 1")); err != nil {
		t.Fatalf("exec: %v", err)
	}
	if _, err := driver.Query(context.Background(), NewQuery("SELECT 1")); err != nil {
		t.Fatalf("query: %v", err)
	}
	if _, err := driver.WriteBatch(context.Background(), validBatch()); err != nil {
		t.Fatalf("write batch: %v", err)
	}
	if _, err := driver.SchemalessWrite(context.Background(), validSchemalessPayload()); err != nil {
		t.Fatalf("schemaless write: %v", err)
	}
	if err := driver.Health(context.Background()); err != nil {
		t.Fatalf("health: %v", err)
	}
	if err := driver.Close(context.Background()); err != nil {
		t.Fatalf("close: %v", err)
	}

	if driver.ExecCalls() != 1 || driver.QueryCalls() != 1 || driver.WriteCalls() != 1 || driver.SchemalessCalls() != 1 || driver.HealthCalls() != 1 || driver.CloseCalls() != 1 || !driver.Closed() {
		t.Fatalf("unexpected fake driver calls or state")
	}
}

func TestConfigValidateRejectsNegativeTimeout(t *testing.T) {
	cfg := validConfig()
	cfg.Timeout = -time.Second

	err := cfg.Validate()
	if !IsKind(err, ErrorKindValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestConfigValidateRejectsNegativeMaxRetries(t *testing.T) {
	cfg := validConfig()
	cfg.MaxRetries = -1

	err := cfg.Validate()
	if !IsKind(err, ErrorKindValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestConfigSanitizeReturnsRedactedDSN(t *testing.T) {
	cfg := Config{
		DriverMode: DriverModeRESTSQLOnly,
		Endpoint:   "https://localhost:6041",
		Database:   "metrics",
		Username:   "root",
		Password:   "hidden",
		TLS:        true,
	}

	sanitized := cfg.Sanitize()
	if sanitized.DSN != "https://root:%2A%2A%2A@localhost:6041/metrics?driver_mode=rest_sql_only" {
		t.Fatalf("unexpected dsn: %q", sanitized.DSN)
	}
	if sanitized.Password != "***" {
		t.Fatalf("password was not redacted: %q", sanitized.Password)
	}
}

func TestNewQueryCopiesArguments(t *testing.T) {
	args := []any{"meters", 7}
	query := NewQuery("SELECT * FROM ?", args...)
	args[0] = "changed"

	if query.Args[0] != "meters" {
		t.Fatalf("query args were not copied: %#v", query.Args)
	}
}

func TestBatchValidateRejectsMissingDatabase(t *testing.T) {
	batch := validBatch()
	batch.Database = ""

	err := batch.Validate()
	if !IsKind(err, ErrorKindValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestBatchValidateRejectsMissingPoints(t *testing.T) {
	err := (Batch{Database: "metrics"}).Validate()
	if !IsKind(err, ErrorKindValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestBatchValidateRejectsMissingTimestamp(t *testing.T) {
	batch := validBatch()
	batch.Points[0].Timestamp = time.Time{}

	err := batch.Validate()
	if !IsKind(err, ErrorKindValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestBatchValidateRejectsMissingFields(t *testing.T) {
	batch := validBatch()
	batch.Points[0].Fields = nil

	err := batch.Validate()
	if !IsKind(err, ErrorKindValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestSchemalessValidateAllowsDefaultPrecision(t *testing.T) {
	err := (SchemalessPayload{
		Protocol: SchemalessTelnetProtocol,
		Lines:    []string{"meters 1700000000 1.2"},
	}).Validate()
	if err != nil {
		t.Fatalf("validate schemaless payload: %v", err)
	}
}

func TestSchemalessValidateRejectsInvalidProtocol(t *testing.T) {
	err := (SchemalessPayload{
		Protocol: "csv",
		Lines:    []string{"meters 1700000000 1.2"},
	}).Validate()
	if !IsKind(err, ErrorKindValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestSchemalessValidateRejectsMissingLines(t *testing.T) {
	err := (SchemalessPayload{Protocol: SchemalessJSONProtocol}).Validate()
	if !IsKind(err, ErrorKindValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestSchemalessValidateRejectsEmptyLine(t *testing.T) {
	err := (SchemalessPayload{
		Protocol: SchemalessLineProtocol,
		Lines:    []string{""},
	}).Validate()
	if !IsKind(err, ErrorKindValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestRenderCreateStableAllowsNoTags(t *testing.T) {
	sql, err := RenderCreateStable(StableSpec{
		Name:    "meters",
		Columns: []ColumnSpec{{Name: "ts", Type: " timestamp "}},
	})
	if err != nil {
		t.Fatalf("render stable: %v", err)
	}
	if sql != "CREATE STABLE IF NOT EXISTS `meters` (`ts` TIMESTAMP)" {
		t.Fatalf("unexpected SQL: %s", sql)
	}
}

func TestRenderCreateStableRejectsInvalidStableName(t *testing.T) {
	_, err := RenderCreateStable(StableSpec{
		Name:    "bad-name",
		Columns: []ColumnSpec{{Name: "ts", Type: "timestamp"}},
	})
	if !IsKind(err, ErrorKindSchema) {
		t.Fatalf("expected schema error, got %v", err)
	}
}

func TestRenderCreateStableRejectsInvalidColumnName(t *testing.T) {
	_, err := RenderCreateStable(StableSpec{
		Name:    "meters",
		Columns: []ColumnSpec{{Name: "bad-name", Type: "timestamp"}},
	})
	if !IsKind(err, ErrorKindSchema) {
		t.Fatalf("expected schema error, got %v", err)
	}
}

func TestRenderCreateStableRejectsInvalidTagName(t *testing.T) {
	_, err := RenderCreateStable(StableSpec{
		Name:    "meters",
		Columns: []ColumnSpec{{Name: "ts", Type: "timestamp"}},
		Tags:    []ColumnSpec{{Name: "bad-name", Type: "binary(16)"}},
	})
	if !IsKind(err, ErrorKindSchema) {
		t.Fatalf("expected schema error, got %v", err)
	}
}

func TestRenderCreateStableRejectsMissingColumns(t *testing.T) {
	_, err := RenderCreateStable(StableSpec{Name: "meters"})
	if !IsKind(err, ErrorKindSchema) {
		t.Fatalf("expected schema error, got %v", err)
	}
}

func TestRenderCreateStableRejectsMissingColumnType(t *testing.T) {
	_, err := RenderCreateStable(StableSpec{
		Name:    "meters",
		Columns: []ColumnSpec{{Name: "ts", Type: " "}},
	})
	if !IsKind(err, ErrorKindSchema) {
		t.Fatalf("expected schema error, got %v", err)
	}
}

func TestErrorHelpersHandlePlainAndConfigErrors(t *testing.T) {
	plain := errors.New("plain")
	if IsRetryable(plain) {
		t.Fatalf("plain errors must not be retryable")
	}
	if IsKind(plain, ErrorKindSQL) {
		t.Fatalf("plain errors must not match taosx kinds")
	}

	cause := errors.New("bad token=hidden")
	cfgErr := configError("load", cause.Error(), cause)
	if !IsKind(cfgErr, ErrorKindConfig) || !errors.Is(cfgErr, cause) {
		t.Fatalf("unexpected config error: %v", cfgErr)
	}
	if strings.Contains(cfgErr.Error(), "hidden") {
		t.Fatalf("config error was not redacted: %v", cfgErr)
	}

	ctxErr := contextError("op", nil)
	if !IsKind(ctxErr, ErrorKindValidation) || ctxErr.Message != "context is required" {
		t.Fatalf("unexpected nil context error: %v", ctxErr)
	}
}

func TestRedactHandlesDelimiterAndMultipleMarkers(t *testing.T) {
	got := redact("token=abc123&password=def456 status=failed")
	if strings.Contains(got, "abc123") || strings.Contains(got, "def456") {
		t.Fatalf("redaction leaked secret values: %q", got)
	}
	if got != "token=***&password=*** status=failed" {
		t.Fatalf("unexpected redaction: %q", got)
	}
}

func TestNewErrorFormatsWithoutOperation(t *testing.T) {
	err := NewError(ErrorKindConfig, "", "bad config", false)
	if err.Error() != "config: bad config" {
		t.Fatalf("unexpected error string: %q", err.Error())
	}
}

func TestWrapErrorUnwrapsCause(t *testing.T) {
	cause := errors.New("root cause")
	err := WrapError(ErrorKindTimeout, "op", "timed out", true, cause)
	if !errors.Is(err, cause) || !IsKind(err, ErrorKindTimeout) || !IsRetryable(err) {
		t.Fatalf("wrapped error missing cause or metadata: %v", err)
	}
}

func TestNilErrorMethodsAreSafe(t *testing.T) {
	var err *Error
	if err.Error() != "" || err.Unwrap() != nil {
		t.Fatalf("nil error methods returned unexpected values")
	}
}

func TestInternalErrorKindFallback(t *testing.T) {
	if errorKind(errors.New("plain")) != ErrorKindInternal {
		t.Fatalf("plain errors should map to internal kind")
	}
}

func TestNoopMetricsAcceptsAllCalls(t *testing.T) {
	metrics := NoopMetrics{}
	metrics.IncCounter(MetricClientRequestsTotal, nil)
	metrics.ObserveHistogram(MetricClientRequestDurationSeconds, 1.5, nil)
	metrics.SetGauge(MetricClientInflight, 1, nil)
}

func validBatch() Batch {
	return Batch{
		Database: "metrics",
		Points: []Point{{
			Table:     "meters",
			Timestamp: time.Unix(1700000000, 0),
			Fields:    map[string]any{"value": 1.2},
		}},
	}
}

func validSchemalessPayload() SchemalessPayload {
	return SchemalessPayload{
		Protocol:  SchemalessLineProtocol,
		Precision: SchemalessPrecisionNanosecond,
		Lines:     []string{"meters,location=office value=1.2 1700000000000000000"},
	}
}

func hasString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

type stubRows struct {
	columns []string
	closed  bool
}

func (r *stubRows) Columns() []string {
	return append([]string(nil), r.columns...)
}

func (r *stubRows) Next() bool {
	return false
}

func (r *stubRows) Scan(...any) error {
	return nil
}

func (r *stubRows) Err() error {
	return nil
}

func (r *stubRows) Close() error {
	r.closed = true
	return nil
}
