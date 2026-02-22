package commands

import (
	"io"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// Style variables are zero-value lipgloss.Style by default (no colors).
// initRenderer updates them with stream-aware, color-capable styles.
// Zero-value styles are nil-renderer-safe: lipgloss falls back to its global
// default renderer on Render(), which auto-detects the terminal.
//
// PassStyle and FailStyle are exported for use by main.go final status messages.
var (
	PassStyle lipgloss.Style
	FailStyle lipgloss.Style

	headerStyle lipgloss.Style
	keyStyle    lipgloss.Style
	valueStyle  lipgloss.Style
	dimStyle    lipgloss.Style
)

// initRenderer configures Lip Gloss styles for the given output stream and color mode.
// colorMode must be one of "auto", "always", or "never".
// Call from PersistentPreRunE after the --color flag is parsed.
func initRenderer(colorMode string, out io.Writer) {
	r := lipgloss.NewRenderer(out)

	switch colorMode {
	case "never":
		r.SetColorProfile(termenv.Ascii)
	case "always":
		// NO_COLOR is respected even with --color=always (no-color.org compliance).
		if os.Getenv("NO_COLOR") == "" {
			r.SetColorProfile(termenv.TrueColor)
		}
		// "auto": r auto-detects TTY, NO_COLOR, and CLICOLOR_FORCE via termenv.
	}

	PassStyle = r.NewStyle().Foreground(lipgloss.Color("2")) // green
	FailStyle = r.NewStyle().Foreground(lipgloss.Color("1")) // red
	headerStyle = r.NewStyle().Bold(true)
	keyStyle = r.NewStyle().Bold(true)
	valueStyle = r.NewStyle().Foreground(lipgloss.Color("6")) // cyan
	dimStyle = r.NewStyle().Faint(true)
}

// statusPrefix returns a colored ✓ or ✗ symbol followed by a space.
func statusPrefix(passed bool) string {
	if passed {
		return PassStyle.Render("✓") + " "
	}
	return FailStyle.Render("✗") + " "
}
