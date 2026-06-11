package taosx

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/ZoneCNH/taosx/internal/sanitize"
	"github.com/ZoneCNH/taosx/internal/validation"
)

type DriverMode string

const (
	DriverModeWebSocket    DriverMode = "websocket"
	DriverModeNativeLegacy DriverMode = "native_legacy"
	DriverModeRESTSQLOnly  DriverMode = "rest_sql_only"
)

type Config struct {
	Name       string
	DriverMode DriverMode
	Endpoint   string
	Database   string
	Username   string
	Password   string
	Timeout    time.Duration
	MaxRetries int
	TLS        bool
}

type SanitizedConfig struct {
	Name       string
	DriverMode DriverMode
	Endpoint   string
	Database   string
	Username   string
	Password   string
	Timeout    time.Duration
	MaxRetries int
	TLS        bool
	DSN        string
}

func (c Config) Normalize() Config {
	if c.Name == "" {
		c.Name = PackageName
	}
	if c.DriverMode == "" {
		c.DriverMode = DriverModeWebSocket
	}
	if c.Timeout == 0 {
		c.Timeout = 5 * time.Second
	}
	return c
}

func (c Config) Validate() error {
	const op = "Config.Validate"
	c = c.Normalize()
	if err := validation.RequireNonEmpty("name", c.Name); err != nil {
		return validationError(op, err.Error(), err)
	}
	if err := validation.RequireNonEmpty("endpoint", c.Endpoint); err != nil {
		return validationError(op, err.Error(), err)
	}
	if err := validation.RequireNonEmpty("database", c.Database); err != nil {
		return validationError(op, err.Error(), err)
	}
	switch c.DriverMode {
	case DriverModeWebSocket, DriverModeNativeLegacy, DriverModeRESTSQLOnly:
	default:
		err := fmt.Errorf("driver_mode must be one of %s, %s, %s", DriverModeWebSocket, DriverModeNativeLegacy, DriverModeRESTSQLOnly)
		return validationError(op, err.Error(), err)
	}
	if c.Timeout < 0 {
		err := errors.New("timeout must not be negative")
		return validationError(op, err.Error(), err)
	}
	if c.MaxRetries < 0 {
		err := errors.New("max_retries must not be negative")
		return validationError(op, err.Error(), err)
	}
	return nil
}

func (c Config) Sanitized() SanitizedConfig {
	c = c.Normalize()
	return SanitizedConfig{
		Name:       c.Name,
		DriverMode: c.DriverMode,
		Endpoint:   c.Endpoint,
		Database:   c.Database,
		Username:   c.Username,
		Password:   sanitize.Secret(c.Password),
		Timeout:    c.Timeout,
		MaxRetries: c.MaxRetries,
		TLS:        c.TLS,
		DSN:        c.RedactedDSN(),
	}
}

func (c Config) Sanitize() SanitizedConfig {
	return c.Sanitized()
}

func (c Config) RedactedDSN() string {
	c = c.Normalize()
	scheme := "taosws"
	if c.DriverMode == DriverModeRESTSQLOnly {
		scheme = "http"
	}
	if c.TLS {
		if c.DriverMode == DriverModeRESTSQLOnly {
			scheme = "https"
		} else {
			scheme = "taoswss"
		}
	}
	u := url.URL{Scheme: scheme, Host: strings.TrimPrefix(strings.TrimPrefix(c.Endpoint, "http://"), "https://"), Path: "/" + c.Database}
	if c.Username != "" {
		u.User = url.UserPassword(c.Username, "***")
	}
	q := url.Values{}
	q.Set("driver_mode", string(c.DriverMode))
	u.RawQuery = q.Encode()
	return u.String()
}
