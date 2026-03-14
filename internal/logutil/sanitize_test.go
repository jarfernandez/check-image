package logutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSanitizeLogValue(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "Normal string unchanged",
			input: "NORMAL_VAR",
			want:  "NORMAL_VAR",
		},
		{
			name:  "Newline replaced with space",
			input: "EVIL_KEY\nfake log line",
			want:  "EVIL_KEY fake log line",
		},
		{
			name:  "Carriage return replaced with space",
			input: "KEY\rinjected",
			want:  "KEY injected",
		},
		{
			name:  "ANSI ESC byte replaced with space",
			input: "path\x1b[2Jhidden",
			want:  "path [2Jhidden",
		},
		{
			name:  "DEL character replaced with space",
			input: "path\x7fvalue",
			want:  "path value",
		},
		{
			name:  "Multiple control characters",
			input: "a\nb\rc\x1bd",
			want:  "a b c d",
		},
		{
			name:  "Empty string",
			input: "",
			want:  "",
		},
		{
			name:  "Unicode characters preserved",
			input: "ruta/configuración/clave",
			want:  "ruta/configuración/clave",
		},
		{
			name:  "String with no control characters",
			input: "/etc/ssl/private.key",
			want:  "/etc/ssl/private.key",
		},
		{
			name:  "Null byte replaced",
			input: "key\x00name",
			want:  "key name",
		},
		{
			name:  "Tab replaced with space",
			input: "key\tname",
			want:  "key name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizeLogValue(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}
