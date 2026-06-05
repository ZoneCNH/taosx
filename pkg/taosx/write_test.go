package taosx

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestWriteBatchValidatesAndDelegates(t *testing.T) {
	driver := &recordingDriver{writeResult: WriteResult{RowsWritten: 1}}
	client, err := New(context.Background(), validConfig(), WithDriver(driver))
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	result, err := client.WriteBatch(context.Background(), Batch{
		Database: "metrics",
		Points: []Point{{
			Table:     "meters",
			Timestamp: time.Unix(1700000000, 0),
			Fields:    map[string]any{"value": 1.2},
		}},
	})
	if err != nil {
		t.Fatalf("write batch: %v", err)
	}
	if result.RowsWritten != 1 || driver.writeCalls != 1 {
		t.Fatalf("batch was not delegated: result=%#v calls=%d", result, driver.writeCalls)
	}
}

func TestWriteBatchRejectsUnsafeTable(t *testing.T) {
	client, err := New(context.Background(), validConfig(), WithDriver(&recordingDriver{}))
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	_, err = client.WriteBatch(context.Background(), Batch{
		Database: "metrics",
		Points: []Point{{
			Table:     "meters;DROP",
			Timestamp: time.Unix(1700000000, 0),
			Fields:    map[string]any{"value": 1.2},
		}},
	})
	if !IsKind(err, ErrorKindValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestSchemalessWriteValidatesAndDelegates(t *testing.T) {
	driver := &recordingDriver{writeResult: WriteResult{RowsWritten: 1}}
	client, err := New(context.Background(), validConfig(), WithDriver(driver))
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	result, err := client.SchemalessWrite(context.Background(), SchemalessPayload{
		Protocol:  SchemalessLineProtocol,
		Precision: SchemalessPrecisionNanosecond,
		Lines:     []string{"meters,location=office value=1.2 1700000000000000000"},
	})
	if err != nil {
		t.Fatalf("schemaless write: %v", err)
	}
	if result.RowsWritten != 1 || driver.schemalessCalls != 1 {
		t.Fatalf("schemaless write was not delegated: result=%#v calls=%d", result, driver.schemalessCalls)
	}
}

func TestSchemalessWriteRejectsInvalidPrecision(t *testing.T) {
	client, err := New(context.Background(), validConfig(), WithDriver(&recordingDriver{}))
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	_, err = client.SchemalessWrite(context.Background(), SchemalessPayload{
		Protocol:  SchemalessLineProtocol,
		Precision: SchemalessPrecision("minute"),
		Lines:     []string{"meters,location=office value=1.2 1700000000000000000"},
	})
	if !IsKind(err, ErrorKindValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestWriteBatchPreservesPartialResultOnDriverError(t *testing.T) {
	driver := NewFakeDriver()
	driver.WriteResult = WriteResult{RowsWritten: 1}
	driver.WriteError = errors.New("tdengine partial write failed password=taosdata")
	client, err := New(context.Background(), validConfig(), WithDriver(driver))
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	result, err := client.WriteBatch(context.Background(), Batch{
		Database: "metrics",
		Points: []Point{{
			Table:     "meters",
			Timestamp: time.Unix(1700000000, 0),
			Fields:    map[string]any{"value": 1.2},
		}},
	})
	if !IsKind(err, ErrorKindWrite) {
		t.Fatalf("expected write error, got %v", err)
	}
	if !result.Partial || result.RowsWritten != 1 || result.RowsAttempted != 1 {
		t.Fatalf("partial result not preserved: %#v", result)
	}
	if strings.Contains(err.Error(), "taosdata") {
		t.Fatalf("write error leaked secret: %v", err)
	}
}
