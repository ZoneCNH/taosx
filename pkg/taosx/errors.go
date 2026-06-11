package taosx

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

type ErrorKind string

const (
	ErrorKindConfig      ErrorKind = "config"
	ErrorKindValidation  ErrorKind = "validation"
	ErrorKindConnection  ErrorKind = "connection"
	ErrorKindUnavailable ErrorKind = "unavailable"
	ErrorKindTimeout     ErrorKind = "timeout"
	ErrorKindAuth        ErrorKind = "auth"
	ErrorKindConflict    ErrorKind = "conflict"
	ErrorKindRateLimit   ErrorKind = "rate_limit"
	ErrorKindSQL         ErrorKind = "sql"
	ErrorKindSchema      ErrorKind = "schema"
	ErrorKindWrite       ErrorKind = "write"
	ErrorKindInternal    ErrorKind = "internal"
)

type Error struct {
	Kind      ErrorKind
	Op        string
	Message   string
	Retryable bool
	Cause     error
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	if e.Op == "" {
		return fmt.Sprintf("%s: %s", e.Kind, e.Message)
	}
	return fmt.Sprintf("%s: %s: %s", e.Kind, e.Op, e.Message)
}

func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

func IsRetryable(err error) bool {
	var target *Error
	if errors.As(err, &target) {
		return target.Retryable
	}
	return false
}

func IsKind(err error, kind ErrorKind) bool {
	var target *Error
	if errors.As(err, &target) {
		return target.Kind == kind
	}
	return false
}

// NewError creates a new Error with the given parameters.
func NewError(kind ErrorKind, op string, message string, retryable bool) *Error {
	return &Error{
		Kind:      kind,
		Op:        op,
		Message:   message,
		Retryable: retryable,
	}
}

// WrapError creates a new Error that wraps an existing error.
func WrapError(kind ErrorKind, op string, message string, retryable bool, cause error) *Error {
	return &Error{
		Kind:      kind,
		Op:        op,
		Message:   message,
		Retryable: retryable,
		Cause:     cause,
	}
}

func errorKind(err error) ErrorKind {
	var target *Error
	if errors.As(err, &target) {
		return target.Kind
	}
	return ErrorKindInternal
}

func validationError(op string, message string, cause error) *Error {
	return &Error{Kind: ErrorKindValidation, Op: op, Message: redact(message), Cause: cause}
}

func configError(op string, message string, cause error) *Error {
	return &Error{Kind: ErrorKindConfig, Op: op, Message: redact(message), Cause: cause}
}

func contextError(op string, cause error) *Error {
	if cause == nil {
		return &Error{Kind: ErrorKindValidation, Op: op, Message: "context is required"}
	}
	if errors.Is(cause, context.DeadlineExceeded) {
		return &Error{Kind: ErrorKindTimeout, Op: op, Message: redact(cause.Error()), Retryable: true, Cause: cause}
	}
	return &Error{Kind: ErrorKindUnavailable, Op: op, Message: redact(cause.Error()), Cause: cause}
}

func driverError(kind ErrorKind, op string, message string, retryable bool, cause error) *Error {
	return &Error{Kind: kind, Op: op, Message: redact(message), Retryable: retryable, Cause: cause}
}

func redact(s string) string {
	for _, marker := range []string{"password=", "passwd=", "token=", "secret="} {
		lower := strings.ToLower(s)
		if idx := strings.Index(lower, marker); idx >= 0 {
			end := strings.IndexAny(s[idx+len(marker):], " \t\n\r&")
			if end < 0 {
				return s[:idx+len(marker)] + "***"
			}
			end += idx + len(marker)
			s = s[:idx+len(marker)] + "***" + s[end:]
		}
	}
	return s
}
