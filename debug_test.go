package fox

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestIsDebugging tests IsDebugging function
func TestIsDebugging(t *testing.T) {
	// Save original mode
	originalMode := foxMode
	defer func() {
		foxMode = originalMode
	}()

	t.Run("debug mode", func(t *testing.T) {
		SetMode(DebugMode)
		assert.True(t, IsDebugging())
	})

	t.Run("release mode", func(t *testing.T) {
		SetMode(ReleaseMode)
		assert.False(t, IsDebugging())
	})

	t.Run("test mode", func(t *testing.T) {
		SetMode(TestMode)
		assert.False(t, IsDebugging())
	})
}

// TestDebugPrint tests debugPrint function
func TestDebugPrint(t *testing.T) {
	// Save original mode and writer
	originalMode := foxMode
	originalWriter := DefaultWriter
	defer func() {
		foxMode = originalMode
		DefaultWriter = originalWriter
	}()

	t.Run("debug mode with newline", func(t *testing.T) {
		SetMode(DebugMode)
		buf := &bytes.Buffer{}
		DefaultWriter = buf

		debugPrint("test message\n")

		output := buf.String()
		assert.Contains(t, output, "[FOX-debug] test message")
	})

	t.Run("debug mode without newline", func(t *testing.T) {
		SetMode(DebugMode)
		buf := &bytes.Buffer{}
		DefaultWriter = buf

		debugPrint("test message")

		output := buf.String()
		assert.Contains(t, output, "[FOX-debug] test message")
		assert.True(t, strings.HasSuffix(output, "\n"))
	})

	t.Run("debug mode with format", func(t *testing.T) {
		SetMode(DebugMode)
		buf := &bytes.Buffer{}
		DefaultWriter = buf

		debugPrint("test %s %d", "message", 123)

		output := buf.String()
		assert.Contains(t, output, "[FOX-debug] test message 123")
	})

	t.Run("release mode", func(t *testing.T) {
		SetMode(ReleaseMode)
		buf := &bytes.Buffer{}
		DefaultWriter = buf

		debugPrint("test message")

		output := buf.String()
		assert.Empty(t, output)
	})
}

// TestDebugPrintRoute tests debugPrintRoute function
func TestDebugPrintRoute(t *testing.T) {
	// Save original mode and writer
	originalMode := foxMode
	originalWriter := DefaultWriter
	defer func() {
		foxMode = originalMode
		DefaultWriter = originalWriter
	}()

	t.Run("debug mode", func(t *testing.T) {
		SetMode(DebugMode)
		buf := &bytes.Buffer{}
		DefaultWriter = buf

		router := New()

		handler := func() string { return "test" }
		handlers := HandlersChain{handler}

		debugPrintRoute(&router.RouterGroup, "GET", "/test", handlers)

		output := buf.String()
		assert.Contains(t, output, "[FOX-debug]")
		assert.Contains(t, output, "GET")
		assert.Contains(t, output, "/test")
		assert.Contains(t, output, "handlers)")
	})

	t.Run("release mode", func(t *testing.T) {
		SetMode(ReleaseMode)
		buf := &bytes.Buffer{}
		DefaultWriter = buf

		router := New()

		handler := func() string { return "test" }
		handlers := HandlersChain{handler}

		debugPrintRoute(&router.RouterGroup, "GET", "/test", handlers)

		output := buf.String()
		assert.Empty(t, output)
	})
}
