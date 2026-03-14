package logutil

import "strings"

// SanitizeLogValue replaces control characters (ASCII < 32 and DEL 127) in
// image-controlled strings with spaces before they are written to log output.
// This prevents log injection via crafted image metadata (CWE-117): a malicious
// image can embed newlines or ANSI escape sequences in env var names or file
// paths; without sanitization those bytes would appear as fabricated log lines.
func SanitizeLogValue(s string) string {
	return strings.Map(func(r rune) rune {
		if r < 32 || r == 127 {
			return ' '
		}
		return r
	}, s)
}
