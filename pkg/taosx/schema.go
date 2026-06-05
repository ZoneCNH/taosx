package taosx

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var identifierPattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]{0,63}$`)

type ColumnSpec struct {
	Name string
	Type string
}

type StableSpec struct {
	Name    string
	Columns []ColumnSpec
	Tags    []ColumnSpec
}

func RenderCreateStable(spec StableSpec) (string, error) {
	const op = "RenderCreateStable"
	if err := validateIdentifier(spec.Name); err != nil {
		return "", driverError(ErrorKindSchema, op, err.Error(), false, err)
	}
	if len(spec.Columns) == 0 {
		err := errors.New("columns are required")
		return "", driverError(ErrorKindSchema, op, err.Error(), false, err)
	}
	columns, err := renderColumns(spec.Columns)
	if err != nil {
		return "", driverError(ErrorKindSchema, op, err.Error(), false, err)
	}
	tags, err := renderColumns(spec.Tags)
	if err != nil {
		return "", driverError(ErrorKindSchema, op, err.Error(), false, err)
	}
	if tags == "" {
		return fmt.Sprintf("CREATE STABLE IF NOT EXISTS %s (%s)", quoteIdentifier(spec.Name), columns), nil
	}
	return fmt.Sprintf("CREATE STABLE IF NOT EXISTS %s (%s) TAGS (%s)", quoteIdentifier(spec.Name), columns, tags), nil
}

func renderColumns(columns []ColumnSpec) (string, error) {
	parts := make([]string, 0, len(columns))
	for _, column := range columns {
		if err := validateIdentifier(column.Name); err != nil {
			return "", err
		}
		if strings.TrimSpace(column.Type) == "" {
			return "", errors.New("column type is required")
		}
		parts = append(parts, quoteIdentifier(column.Name)+" "+strings.ToUpper(strings.TrimSpace(column.Type)))
	}
	return strings.Join(parts, ", "), nil
}

func validateIdentifier(identifier string) error {
	if !identifierPattern.MatchString(identifier) {
		return fmt.Errorf("identifier %q must match %s", identifier, identifierPattern.String())
	}
	return nil
}

func quoteIdentifier(identifier string) string {
	return "`" + identifier + "`"
}

func itoa(i int) string {
	return fmt.Sprintf("%d", i)
}
