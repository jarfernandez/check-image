package commands

import (
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	cterm "github.com/charmbracelet/x/term"
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

// termOut stores the output writer supplied to initRenderer for terminal width detection.
var termOut io.Writer

// initRenderer configures Lip Gloss styles for the given output stream and color mode.
// colorMode must be one of "auto", "always", or "never".
// Call from PersistentPreRunE after the --color flag is parsed.
func initRenderer(colorMode string, out io.Writer) {
	termOut = out
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

// terminalWidth returns the width of the terminal associated with termOut,
// falling back to defaultTermWidth when the output is not a TTY or the size
// cannot be determined.
const defaultTermWidth = 60

func terminalWidth() int {
	if f, ok := termOut.(*os.File); ok {
		if w, _, err := cterm.GetSize(f.Fd()); err == nil && w > 0 {
			return w
		}
	}
	return defaultTermWidth
}

// sectionHeader renders a horizontal rule with the check name for use as a
// section separator in the all command. Works in both color and no-color modes:
//
//	── name ──────────────────────────────────────────────────
func sectionHeader(name string) string {
	width := terminalWidth()

	const leftPrefix = "── "
	const rightPrefix = " "

	rightLen := width - len([]rune(leftPrefix)) - len([]rune(name)) - len([]rune(rightPrefix))
	if rightLen < 2 {
		rightLen = 2
	}

	left := dimStyle.Render(leftPrefix)
	middle := headerStyle.Render(name)
	right := dimStyle.Render(rightPrefix + strings.Repeat("─", rightLen))

	return left + middle + right
}

// statusPrefix returns a colored ✓ or ✗ symbol followed by a space.
func statusPrefix(passed bool) string {
	if passed {
		return PassStyle.Render("✓") + " "
	}
	return FailStyle.Render("✗") + " "
}
