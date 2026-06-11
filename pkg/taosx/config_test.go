package taosx

import (
	"strings"
	"testing"
	"time"
)

func TestConfigValidateDefaultsDriverMode(t *testing.T) {
	cfg := Config{
		Endpoint: "localhost:6041",
		Database: "metrics",
	}

	normalized := cfg.Normalize()
	if normalized.Name != PackageName {
		t.Fatalf("unexpected default name: %q", normalized.Name)
	}
	if normalized.DriverMode != DriverModeWebSocket {
		t.Fatalf("unexpected default driver mode: %q", normalized.DriverMode)
	}
	if normalized.Timeout != 5*time.Second {
		t.Fatalf("unexpected default timeout: %s", normalized.Timeout)
	}
	if err := normalized.Validate(); err != nil {
		t.Fatalf("validate normalized config: %v", err)
	}
}

func TestConfigValidateRequiresEndpointAndDatabase(t *testing.T) {
	err := (Config{}).Validate()
	if !IsKind(err, ErrorKindValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}

	err = (Config{Endpoint: "localhost:6041"}).Validate()
	if !IsKind(err, ErrorKindValidation) || !strings.Contains(err.Error(), "database is required") {
		t.Fatalf("expected database validation error, got %v", err)
	}
}

func TestConfigValidateRejectsInvalidDriverMode(t *testing.T) {
	err := (Config{
		DriverMode: "odbc",
		Endpoint:   "localhost:6041",
		Database:   "metrics",
	}).Validate()
	if !IsKind(err, ErrorKindValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestConfigSanitizedRedactsPasswordAndDSN(t *testing.T) {
	cfg := Config{
		Endpoint: "http://localhost:6041",
		Database: "metrics",
		Username: "root",
		Password: "taosdata",
		TLS:      true,
	}

	sanitized := cfg.Sanitized()
	if sanitized.Password != "***" {
		t.Fatalf("password was not redacted: %q", sanitized.Password)
	}
	if strings.Contains(sanitized.DSN, "taosdata") {
		t.Fatalf("dsn leaked password: %q", sanitized.DSN)
	}
	if !strings.Contains(sanitized.DSN, "driver_mode=websocket") {
		t.Fatalf("dsn missing driver mode: %q", sanitized.DSN)
	}
}
