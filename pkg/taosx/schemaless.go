package taosx

type SchemalessProtocol string
type SchemalessPrecision string

const (
	SchemalessLineProtocol   SchemalessProtocol = "line"
	SchemalessTelnetProtocol SchemalessProtocol = "telnet"
	SchemalessJSONProtocol   SchemalessProtocol = "json"
)

const (
	SchemalessPrecisionHour        SchemalessPrecision = "hour"
	SchemalessPrecisionMinute      SchemalessPrecision = "min"
	SchemalessPrecisionSecond      SchemalessPrecision = "sec"
	SchemalessPrecisionMillisecond SchemalessPrecision = "ms"
	SchemalessPrecisionMicrosecond SchemalessPrecision = "us"
	SchemalessPrecisionNanosecond  SchemalessPrecision = "ns"
)

type SchemalessPayload struct {
	Protocol  SchemalessProtocol
	Precision SchemalessPrecision
	Lines     []string
}

func (p SchemalessPayload) Validate() error {
	const op = "SchemalessPayload.Validate"
	switch p.Protocol {
	case SchemalessLineProtocol, SchemalessTelnetProtocol, SchemalessJSONProtocol:
	default:
		return validationError(op, "protocol must be line, telnet, or json", nil)
	}
	switch p.Precision {
	case "", SchemalessPrecisionHour, SchemalessPrecisionMinute, SchemalessPrecisionSecond, SchemalessPrecisionMillisecond, SchemalessPrecisionMicrosecond, SchemalessPrecisionNanosecond:
	default:
		return validationError(op, "precision must be hour, min, sec, ms, us, or ns", nil)
	}
	if len(p.Lines) == 0 {
		return validationError(op, "lines are required", nil)
	}
	for i, line := range p.Lines {
		if line == "" {
			return validationError(op, "lines["+itoa(i)+"] is required", nil)
		}
	}
	return nil
}
