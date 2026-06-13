//go:build integration

package taosx

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	_ "github.com/taosdata/driver-go/v3/taosWS"
)

func TestTDengineWebSocketIntegration(t *testing.T) {
	if os.Getenv("TAOSX_INTEGRATION") != "1" {
		t.Skip("set TAOSX_INTEGRATION=1 to run TDengine WebSocket integration")
	}

	settings, err := tdengineIntegrationSettingsFromEnv()
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	db, err := sql.Open("taosWS", settings.dsn)
	if err != nil {
		t.Fatalf("open TDengine WebSocket %s: %s", settings.redactedDSN, sanitizeTDengineError(err, settings))
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(time.Minute)

	client, err := New(ctx, settings.config, WithDriver(&sqlTDengineDriver{db: db}))
	if err != nil {
		t.Fatalf("new taosx client for %s: %s", settings.redactedDSN, sanitizeTDengineError(err, settings))
	}
	defer func() {
		if err := client.Close(context.Background()); err != nil {
			t.Errorf("close TDengine client: %s", sanitizeTDengineError(err, settings))
		}
	}()

	status := client.Health(ctx)
	if status.Status != HealthHealthy {
		t.Fatalf("health status = %s via %s: %s", status.Status, settings.redactedDSN, sanitizeTDengineText(status.Message, settings))
	}

	rows, err := client.Query(ctx, NewQuery("SHOW TABLES"))
	if err != nil {
		t.Fatalf("query SHOW TABLES via %s: %s", settings.redactedDSN, sanitizeTDengineError(err, settings))
	}
	defer rows.Close()

	if columns := rows.Columns(); len(columns) == 0 {
		t.Fatalf("SHOW TABLES returned no columns via %s", settings.redactedDSN)
	}

	for rows.Next() {
		// Iterating at least once through sql.Rows validates the taosx Rows adapter.
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("read SHOW TABLES rows via %s: %s", settings.redactedDSN, sanitizeTDengineError(err, settings))
	}
}

type tdengineIntegrationSettings struct {
	config      Config
	dsn         string
	redactedDSN string
}

func tdengineIntegrationSettingsFromEnv() (tdengineIntegrationSettings, error) {
	endpoint := os.Getenv("TAOSX_TDENGINE_ENDPOINT")
	username := os.Getenv("TAOSX_TDENGINE_USER")
	password := os.Getenv("TAOSX_TDENGINE_PASSWORD")
	database := os.Getenv("TAOSX_TDENGINE_DATABASE")
	tls := boolEnv("TAOSX_TDENGINE_TLS")
	dsn := os.Getenv("TAOSX_TDENGINE_DSN")

	if dsn != "" {
		parsed, err := parseTDengineWSDSN(dsn)
		if err != nil {
			return tdengineIntegrationSettings{}, fmt.Errorf("parse TAOSX_TDENGINE_DSN: %w", err)
		}
		if endpoint == "" {
			endpoint = parsed.endpoint
		}
		if username == "" {
			username = parsed.username
		}
		if password == "" {
			password = parsed.password
		}
		if database == "" {
			database = parsed.database
		}
		if parsed.tls {
			tls = true
		}
	}

	missing := missingTDengineEnv(endpoint, username, password, database)
	if len(missing) > 0 {
		return tdengineIntegrationSettings{}, fmt.Errorf("missing TDengine integration env: %s", strings.Join(missing, ", "))
	}

	endpoint = normalizeTDengineEndpoint(endpoint)
	cfg := Config{
		Endpoint:   endpoint,
		Database:   database,
		Username:   username,
		Password:   password,
		DriverMode: DriverModeWebSocket,
		TLS:        tls,
		Timeout:    10 * time.Second,
	}
	if err := cfg.Normalize().Validate(); err != nil {
		return tdengineIntegrationSettings{}, fmt.Errorf("validate TDengine integration config: %w", err)
	}

	if dsn == "" {
		dsn = tdengineWSDSN(username, password, endpoint, database, tls)
	}

	return tdengineIntegrationSettings{
		config:      cfg,
		dsn:         dsn,
		redactedDSN: redactTDengineDSN(dsn),
	}, nil
}

func missingTDengineEnv(endpoint, username, password, database string) []string {
	var missing []string
	if endpoint == "" {
		missing = append(missing, "TAOSX_TDENGINE_ENDPOINT")
	}
	if username == "" {
		missing = append(missing, "TAOSX_TDENGINE_USER")
	}
	if password == "" {
		missing = append(missing, "TAOSX_TDENGINE_PASSWORD")
	}
	if database == "" {
		missing = append(missing, "TAOSX_TDENGINE_DATABASE")
	}
	return missing
}

func tdengineWSDSN(username, password, endpoint, database string, tls bool) string {
	scheme := "ws"
	if tls {
		scheme = "wss"
	}
	return fmt.Sprintf("%s:%s@%s(%s)/%s", url.QueryEscape(username), url.QueryEscape(password), scheme, endpoint, url.PathEscape(database))
}

type parsedTDengineDSN struct {
	endpoint string
	username string
	password string
	database string
	tls      bool
}

func parseTDengineWSDSN(dsn string) (parsedTDengineDSN, error) {
	credentials, rest, found := strings.Cut(dsn, "@")
	if !found {
		return parsedTDengineDSN{}, fmt.Errorf("missing credentials separator")
	}
	username, password, found := strings.Cut(credentials, ":")
	if !found {
		return parsedTDengineDSN{}, fmt.Errorf("missing password separator")
	}

	scheme, rest, found := strings.Cut(rest, "(")
	if !found {
		return parsedTDengineDSN{}, fmt.Errorf("missing endpoint start")
	}
	endpoint, databasePath, found := strings.Cut(rest, ")")
	if !found {
		return parsedTDengineDSN{}, fmt.Errorf("missing endpoint end")
	}
	databasePath = strings.TrimPrefix(databasePath, "/")
	databasePath, _, _ = strings.Cut(databasePath, "?")
	databasePath, _, _ = strings.Cut(databasePath, "#")

	decodedUser, err := url.QueryUnescape(username)
	if err != nil {
		return parsedTDengineDSN{}, fmt.Errorf("decode username: %w", err)
	}
	decodedPassword, err := url.QueryUnescape(password)
	if err != nil {
		return parsedTDengineDSN{}, fmt.Errorf("decode password: %w", err)
	}
	decodedDatabase, err := url.PathUnescape(databasePath)
	if err != nil {
		return parsedTDengineDSN{}, fmt.Errorf("decode database: %w", err)
	}

	return parsedTDengineDSN{
		endpoint: endpoint,
		username: decodedUser,
		password: decodedPassword,
		database: decodedDatabase,
		tls:      scheme == "wss",
	}, nil
}

func normalizeTDengineEndpoint(endpoint string) string {
	endpoint = strings.TrimSpace(endpoint)
	endpoint = strings.TrimPrefix(endpoint, "http://")
	endpoint = strings.TrimPrefix(endpoint, "https://")
	endpoint = strings.TrimPrefix(endpoint, "ws://")
	endpoint = strings.TrimPrefix(endpoint, "wss://")
	endpoint, _, _ = strings.Cut(endpoint, "/")
	if !strings.Contains(endpoint, ":") {
		endpoint += ":6041"
	}
	return endpoint
}

func boolEnv(key string) bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv(key))) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func redactTDengineDSN(dsn string) string {
	credentials, rest, found := strings.Cut(dsn, "@")
	if !found {
		return "<redacted>"
	}
	_, _, found = strings.Cut(credentials, ":")
	if !found {
		return "<redacted>"
	}
	rest, _, _ = strings.Cut(rest, "?")
	return "<user>:***@" + rest
}

func sanitizeTDengineError(err error, settings tdengineIntegrationSettings) string {
	if err == nil {
		return ""
	}
	return sanitizeTDengineText(err.Error(), settings)
}

func sanitizeTDengineText(text string, settings tdengineIntegrationSettings) string {
	replacements := []struct {
		old string
		new string
	}{
		{settings.dsn, "<redacted-dsn>"},
		{settings.config.Password, "<password>"},
		{url.QueryEscape(settings.config.Password), "<password>"},
		{settings.config.Username, "<user>"},
		{url.QueryEscape(settings.config.Username), "<user>"},
	}
	for _, replacement := range replacements {
		if replacement.old != "" {
			text = strings.ReplaceAll(text, replacement.old, replacement.new)
		}
	}
	return text
}

type sqlTDengineDriver struct {
	db *sql.DB
}

func (d *sqlTDengineDriver) Exec(ctx context.Context, stmt Statement) (ExecResult, error) {
	result, err := d.db.ExecContext(ctx, stmt.SQL, stmt.Args...)
	if err != nil {
		return ExecResult{}, err
	}
	rowsAffected, _ := result.RowsAffected()
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
	return WriteResult{}, driverError(ErrorKindUnavailable, "tdengine_sql_driver.WriteBatch", "batch write is not covered by sql integration adapter", false, nil)
}

func (d *sqlTDengineDriver) SchemalessWrite(context.Context, SchemalessPayload) (WriteResult, error) {
	return WriteResult{}, driverError(ErrorKindUnavailable, "tdengine_sql_driver.SchemalessWrite", "schemaless write is not covered by sql integration adapter", false, nil)
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
