package output

import (
	"encoding/json"
	"fmt"
	"io"
)

// Format represents the output format for CLI results.
type Format string

const (
	FormatText Format = "text"
	FormatJSON Format = "json"
)

// ParseFormat parses a string into a Format, returning an error for unsupported values.
func ParseFormat(s string) (Format, error) {
	switch s {
	case string(FormatText):
		return FormatText, nil
	case string(FormatJSON):
		return FormatJSON, nil
	default:
		return "", fmt.Errorf("unsupported output format %q, valid values are: text, json", s)
	}
}

// RenderJSON writes v as indented JSON to w.
func RenderJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	return enc.Encode(v)
}
