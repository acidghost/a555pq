//nolint:revive
package shared

type OutputFormatType int

const (
	Table OutputFormatType = iota
	JSON
)

var OutputFormat OutputFormatType

func (of OutputFormatType) String() string {
	switch of {
	case Table:
		return "table"
	case JSON:
		return "json"
	default:
		return "table"
	}
}

func ParseOutputFormat(s string) OutputFormatType {
	switch s {
	case "table":
		return Table
	case "json":
		return JSON
	default:
		return Table
	}
}

func (of *OutputFormatType) Set(s string) error {
	*of = ParseOutputFormat(s)
	return nil
}

func (of OutputFormatType) Type() string {
	return "string"
}
