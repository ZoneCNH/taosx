package taosx

import (
	"context"
	"errors"
	"reflect"
	"sort"
	"testing"
	"time"
)

func TestClientAPISnapshot(t *testing.T) {
	clientType := reflect.TypeOf((*Client)(nil)).Elem()
	got := make([]string, 0, clientType.NumMethod())
	for i := 0; i < clientType.NumMethod(); i++ {
		method := clientType.Method(i)
		got = append(got, method.Name+" "+method.Type.String())
	}
	sort.Strings(got)

	want := []string{
		"Close func(context.Context) error",
		"DeleteRange func(context.Context, string, time.Time) (taosx.ExecResult, error)",
		"Exec func(context.Context, taosx.Statement) (taosx.ExecResult, error)",
		"Health func(context.Context) taosx.HealthStatus",
		"Query func(context.Context, taosx.Query) (taosx.Rows, error)",
		"SchemalessWrite func(context.Context, taosx.SchemalessPayload) (taosx.WriteResult, error)",
		"WriteBatch func(context.Context, taosx.Batch) (taosx.WriteResult, error)",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("client api drift\nwant: %#v\n got: %#v", want, got)
	}
}

func TestFakeClientImplementsClient(t *testing.T) {
	fake := NewFakeClient()
	var client Client = fake
	fake.ExecResult = ExecResult{RowsAffected: 2}

	result, err := client.Exec(context.Background(), NewStatement("CREATE DATABASE IF NOT EXISTS metrics"))
	if err != nil {
		t.Fatalf("exec: %v", err)
	}
	if result.RowsAffected != 2 || fake.ExecCalls() != 1 {
		t.Fatalf("unexpected exec result/calls: %#v calls=%d", result, fake.ExecCalls())
	}

	fake.DeleteResult = ExecResult{RowsAffected: 4}
	deleteResult, err := client.DeleteRange(context.Background(), "cpu", time.Unix(2, 0))
	if err != nil {
		t.Fatalf("delete range: %v", err)
	}
	if deleteResult.RowsAffected != 4 || fake.DeleteCalls() != 1 {
		t.Fatalf("unexpected delete result/calls: %#v calls=%d", deleteResult, fake.DeleteCalls())
	}

	fake.WriteResult = WriteResult{RowsWritten: 1, RowsAttempted: 1}
	writeResult, err := client.WriteBatch(context.Background(), Batch{
		Database: "metrics",
		Points: []Point{{
			Table:     "cpu",
			Timestamp: time.Unix(1, 0),
			Fields:    map[string]any{"usage": 0.8},
		}},
	})
	if err != nil {
		t.Fatalf("write batch: %v", err)
	}
	if writeResult.RowsWritten != 1 || fake.WriteCalls() != 1 {
		t.Fatalf("unexpected write result/calls: %#v calls=%d", writeResult, fake.WriteCalls())
	}

	if status := client.Health(context.Background()); status.Status != HealthHealthy {
		t.Fatalf("health status = %q, want %q", status.Status, HealthHealthy)
	}
	if err := client.Close(context.Background()); err != nil {
		t.Fatalf("close: %v", err)
	}
	if !fake.Closed() || fake.CloseCalls() != 1 {
		t.Fatalf("closed=%t close calls=%d, want closed once", fake.Closed(), fake.CloseCalls())
	}
}

func TestClientExecDelegatesAndCloseIsIdempotent(t *testing.T) {
	driver := &recordingDriver{execResult: ExecResult{RowsAffected: 3}}
	metrics := &recordingMetrics{}
	client, err := New(context.Background(), validConfig(), WithDriver(driver), WithMetrics(metrics))
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	result, err := client.Exec(context.Background(), NewStatement("CREATE DATABASE IF NOT EXISTS metrics"))
	if err != nil {
		t.Fatalf("exec: %v", err)
	}
	if result.RowsAffected != 3 {
		t.Fatalf("unexpected exec result: %#v", result)
	}
	if driver.execCalls != 1 {
		t.Fatalf("exec calls = %d, want 1", driver.execCalls)
	}
	if err := client.Close(context.Background()); err != nil {
		t.Fatalf("close: %v", err)
	}
	if err := client.Close(context.Background()); err != nil {
		t.Fatalf("second close: %v", err)
	}
	if driver.closeCalls != 1 {
		t.Fatalf("close calls = %d, want 1", driver.closeCalls)
	}
	if !metrics.hasCounter(MetricClientCreatedTotal) || !metrics.hasCounter(MetricClientClosedTotal) {
		t.Fatalf("expected lifecycle metrics, got %#v", metrics.counters)
	}
}

func TestDefaultDriverReturnsRetryableUnavailable(t *testing.T) {
	client, err := New(context.Background(), validConfig())
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	_, err = client.Exec(context.Background(), NewStatement("CREATE DATABASE IF NOT EXISTS metrics"))
	if !IsKind(err, ErrorKindUnavailable) || !IsRetryable(err) {
		t.Fatalf("expected retryable unavailable error, got %v", err)
	}
}

func TestClientContextErrors(t *testing.T) {
	client, err := New(context.Background(), validConfig(), WithDriver(&recordingDriver{}))
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
	defer cancel()
	<-ctx.Done()
	_, err = client.Exec(ctx, NewStatement("CREATE DATABASE IF NOT EXISTS metrics"))
	if !IsKind(err, ErrorKindTimeout) || !IsRetryable(err) {
		t.Fatalf("expected retryable timeout, got %v", err)
	}
}

func validConfig() Config {
	return Config{
		Endpoint: "localhost:6041",
		Database: "metrics",
		Username: "root",
		Password: "taosdata",
	}
}

type recordingDriver struct {
	execResult      ExecResult
	execErr         error
	queryRows       Rows
	queryErr        error
	writeResult     WriteResult
	deleteResult    ExecResult
	writeErr        error
	deleteErr       error
	schemalessErr   error
	healthErr       error
	closeErr        error
	execCalls       int
	queryCalls      int
	writeCalls      int
	deleteCalls     int
	schemalessCalls int
	closeCalls      int
}

func (d *recordingDriver) Exec(context.Context, Statement) (ExecResult, error) {
	d.execCalls++
	return d.execResult, d.execErr
}

func (d *recordingDriver) Query(context.Context, Query) (Rows, error) {
	d.queryCalls++
	return d.queryRows, d.queryErr
}

func (d *recordingDriver) WriteBatch(context.Context, Batch) (WriteResult, error) {
	d.writeCalls++
	return d.writeResult, d.writeErr
}

func (d *recordingDriver) SchemalessWrite(context.Context, SchemalessPayload) (WriteResult, error) {
	d.schemalessCalls++
	return d.writeResult, d.schemalessErr
}

func (d *recordingDriver) DeleteRange(context.Context, string, time.Time) (ExecResult, error) {
	d.deleteCalls++
	return d.deleteResult, d.deleteErr
}

func (d *recordingDriver) Health(context.Context) error {
	return d.healthErr
}

func (d *recordingDriver) Close(context.Context) error {
	d.closeCalls++
	return d.closeErr
}

type recordingMetrics struct {
	counters []string
	gauges   []string
}

func (m *recordingMetrics) IncCounter(name string, _ map[string]string) {
	m.counters = append(m.counters, name)
}

func (m *recordingMetrics) ObserveHistogram(string, float64, map[string]string) {}

func (m *recordingMetrics) SetGauge(name string, _ float64, _ map[string]string) {
	m.gauges = append(m.gauges, name)
}

func (m *recordingMetrics) hasCounter(name string) bool {
	for _, candidate := range m.counters {
		if candidate == name {
			return true
		}
	}
	return false
}

func (m *recordingMetrics) counterCount(name string) int {
	count := 0
	for _, candidate := range m.counters {
		if candidate == name {
			count++
		}
	}
	return count
}

var errDriverHealth = errors.New("connect failed password=taosdata")
