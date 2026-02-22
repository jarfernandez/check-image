package commands

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestMain initializes the renderer in "never" mode before running any tests
// in the commands package, ensuring plain-text output for deterministic assertions.
func TestMain(m *testing.M) {
	initRenderer("never", os.Stdout)
	os.Exit(m.Run())
}

func TestInitRenderer_Never(t *testing.T) {
	var buf bytes.Buffer
	initRenderer("never", &buf)
	t.Cleanup(func() { initRenderer("never", os.Stdout) })

	rendered := PassStyle.Render("hello")
	assert.Equal(t, "hello", rendered)
	assert.False(t, strings.Contains(rendered, "\x1b["), "expected no ANSI codes with --color=never")
}

func TestInitRenderer_Always(t *testing.T) {
	if os.Getenv("NO_COLOR") != "" {
		t.Skip("NO_COLOR is set; --color=always defers to NO_COLOR")
	}
	t.Cleanup(func() { initRenderer("never", os.Stdout) })

	var buf bytes.Buffer
	initRenderer("always", &buf)

	rendered := PassStyle.Render("hello")
	assert.True(t, strings.Contains(rendered, "\x1b["), "expected ANSI codes with --color=always")
}

func TestInitRenderer_Always_WithNOCOLOR(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	t.Cleanup(func() { initRenderer("never", os.Stdout) })

	var buf bytes.Buffer
	initRenderer("always", &buf)

	rendered := PassStyle.Render("hello")
	assert.Equal(t, "hello", rendered)
	assert.False(t, strings.Contains(rendered, "\x1b["), "NO_COLOR must override --color=always")
}

func TestInitRenderer_Auto_NonTTY(t *testing.T) {
	t.Cleanup(func() { initRenderer("never", os.Stdout) })

	// bytes.Buffer is not a TTY, so auto mode must produce no ANSI codes.
	var buf bytes.Buffer
	initRenderer("auto", &buf)

	rendered := PassStyle.Render("hello")
	assert.Equal(t, "hello", rendered)
	assert.False(t, strings.Contains(rendered, "\x1b["), "expected no ANSI codes when output is not a TTY")
}

func TestStatusPrefix_Pass(t *testing.T) {
	prefix := statusPrefix(true)
	assert.Contains(t, prefix, "✓")
}

func TestStatusPrefix_Fail(t *testing.T) {
	prefix := statusPrefix(false)
	assert.Contains(t, prefix, "✗")
}
