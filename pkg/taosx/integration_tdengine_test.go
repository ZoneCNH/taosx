//go:build integration

package taosx

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"testing"
	"time"

	_ "github.com/taosdata/driver-go/v3/taosWS"
)

func TestTDengineWebSocketIntegration(t *testing.T) {
	env := tdengineIntegrationEnv(t)
	db, err := sql.Open(env.driver, env.dsn)
	if err != nil {
		t.Fatalf("open TDengine connection failed")
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cfg := Config{
		Name:       "tdengine-integration",
		DriverMode: DriverModeWebSocket,
		Endpoint:   env.endpoint,
		Database:   env.database,
		Username:   env.user,
		Password:   env.password,
		Timeout:    5 * time.Second,
	}
	client, err := New(ctx, cfg, WithDriver(&sqlTDengineDriver{db: db}))
	if err != nil {
		t.Fatalf("new taosx client failed: kind=%s", errorKind(err))
	}
	t.Cleanup(func() {
		if err := client.Close(context.Background()); err != nil {
			t.Fatalf("close taosx client failed: kind=%s", errorKind(err))
		}
	})

	health := client.Health(ctx)
	if health.Status != HealthHealthy {
		t.Fatalf("TDengine health status = %s", health.Status)
	}

	rows, err := client.Query(ctx, NewQuery("SHOW DATABASES"))
	if err != nil {
		t.Fatalf("query TDengine metadata failed: kind=%s", errorKind(err))
	}
	defer rows.Close()

	columns := rows.Columns()
	if len(columns) == 0 {
		t.Fatalf("expected metadata query columns")
	}
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			t.Fatalf("iterate metadata rows: %v", err)
		}
		t.Fatalf("expected at least one TDengine database row")
	}
	values := make([]any, len(columns))
	raw := make([]sql.RawBytes, len(columns))
	for i := range raw {
		values[i] = &raw[i]
	}
	if err := rows.Scan(values...); err != nil {
		t.Fatalf("scan metadata row: %v", err)
	}
	hasValue := false
	for _, value := range raw {
		if len(value) > 0 {
			hasValue = true
			break
		}
	}
	if !hasValue {
		t.Fatalf("expected non-empty metadata row")
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("metadata rows error: %v", err)
	}
}

type tdengineIntegrationConfig struct {
	driver   string
	dsn      string
	endpoint string
	database string
	user     string
	password string
}

func tdengineIntegrationEnv(t *testing.T) tdengineIntegrationConfig {
	t.Helper()
	if os.Getenv("TAOSX_INTEGRATION") != "1" {
		t.Skip("set TAOSX_INTEGRATION=1 to run real TDengine integration tests")
	}
	driver := envDefault("TAOSX_TDENGINE_DRIVER", "taosWS")
	if driver != "taosWS" {
		t.Skipf("TDengine integration test currently supports taosWS, got %q", driver)
	}

	dsn := os.Getenv("TAOSX_TDENGINE_DSN")
	endpoint := os.Getenv("TAOSX_TDENGINE_ENDPOINT")
	user := os.Getenv("TAOSX_TDENGINE_USER")
	password := os.Getenv("TAOSX_TDENGINE_PASSWORD")
	database := os.Getenv("TAOSX_TDENGINE_DATABASE")
	if dsn == "" {
		requireEnv(t, "TAOSX_TDENGINE_ENDPOINT", endpoint)
		requireEnv(t, "TAOSX_TDENGINE_USER", user)
		requireEnv(t, "TAOSX_TDENGINE_PASSWORD", password)
		requireEnv(t, "TAOSX_TDENGINE_DATABASE", database)
		dsn = fmt.Sprintf("%s@ws(%s)/%s", url.UserPassword(user, password).String(), endpoint, database)
	}
	if endpoint == "" {
		endpoint = "configured"
	}
	if database == "" {
		database = "configured"
	}

	return tdengineIntegrationConfig{
		driver:   driver,
		dsn:      dsn,
		endpoint: endpoint,
		database: database,
		user:     user,
		password: password,
	}
}

func requireEnv(t *testing.T, name string, value string) {
	t.Helper()
	if value == "" {
		t.Skipf("%s is required for TDengine integration tests", name)
	}
}

func envDefault(name string, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}

type sqlTDengineDriver struct {
	db *sql.DB
}

func (d *sqlTDengineDriver) Exec(ctx context.Context, stmt Statement) (ExecResult, error) {
	result, err := d.db.ExecContext(ctx, stmt.SQL, stmt.Args...)
	if err != nil {
		return ExecResult{}, err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return ExecResult{}, nil
	}
	return ExecResult{RowsAffected: rowsAffected}, nil
}

func (d *sqlTDengineDriver) Query(ctx context.Context, query Query) (Rows, error) {
	rows, err := d.db.QueryContext(ctx, query.SQL, query.Args...)
	if err != nil {
		return nil, err
	}
	return &sqlTDengineRows{rows: rows}, nil
}

func (d *sqlTDengineDriver) WriteBatch(context.Context, Batch) (WriteResult, error) {
	return WriteResult{}, NewError(ErrorKindUnavailable, "sqlTDengineDriver.WriteBatch", "write batch is outside this integration adapter", true)
}

func (d *sqlTDengineDriver) SchemalessWrite(context.Context, SchemalessPayload) (WriteResult, error) {
	return WriteResult{}, NewError(ErrorKindUnavailable, "sqlTDengineDriver.SchemalessWrite", "schemaless write is outside this integration adapter", true)
}

func (d *sqlTDengineDriver) Health(ctx context.Context) error {
	return d.db.PingContext(ctx)
}

func (d *sqlTDengineDriver) Close(context.Context) error {
	return d.db.Close()
}

type sqlTDengineRows struct {
	rows *sql.Rows
}

func (r *sqlTDengineRows) Columns() []string {
	columns, err := r.rows.Columns()
	if err != nil {
		return nil
	}
	return columns
}

func (r *sqlTDengineRows) Next() bool {
	return r.rows.Next()
}

func (r *sqlTDengineRows) Scan(dest ...any) error {
	return r.rows.Scan(dest...)
}

func (r *sqlTDengineRows) Err() error {
	return r.rows.Err()
}

func (r *sqlTDengineRows) Close() error {
	return r.rows.Close()
}
