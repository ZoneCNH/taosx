package templatex

import (
	"errors"
	"time"
)

type Config struct {
	Name    string
	Timeout time.Duration
	Secret  string
}

type SanitizedConfig struct {
	Name    string
	Timeout time.Duration
	Secret  string
}

func (c Config) Validate() error {
	if c.Name == "" {
		return errors.New("name is required")
	}
	if c.Timeout < 0 {
		return errors.New("timeout must not be negative")
	}
	return nil
}

func (c Config) Sanitize() SanitizedConfig {
	secret := ""
	if c.Secret != "" {
		secret = "***"
	}
	return SanitizedConfig{
		Name:    c.Name,
		Timeout: c.Timeout,
		Secret:  secret,
	}
}
