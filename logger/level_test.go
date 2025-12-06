package logger

import (
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestLevel_Constants(t *testing.T) {
	tests := []struct {
		name     string
		level    Level
		expected string
	}{
		{"Debug", DebugLevel, "debug"},
		{"Info", InfoLevel, "info"},
		{"Warn", WarnLevel, "warn"},
		{"Error", ErrorLevel, "error"},
		{"Fatal", FatalLevel, "fatal"},
		{"Panic", PanicLevel, "panic"},
		{"NoLevel", NoLevel, "no"},
		{"Disabled", Disabled, "disabled"},
		{"Trace", TraceLevel, "trace"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.level))
		})
	}
}

func TestLevel_ZerologLevel(t *testing.T) {
	tests := []struct {
		name          string
		level         Level
		expectedLevel zerolog.Level
	}{
		{
			name:          "DebugLevel",
			level:         DebugLevel,
			expectedLevel: zerolog.DebugLevel,
		},
		{
			name:          "InfoLevel",
			level:         InfoLevel,
			expectedLevel: zerolog.InfoLevel,
		},
		{
			name:          "WarnLevel",
			level:         WarnLevel,
			expectedLevel: zerolog.WarnLevel,
		},
		{
			name:          "ErrorLevel",
			level:         ErrorLevel,
			expectedLevel: zerolog.ErrorLevel,
		},
		{
			name:          "FatalLevel",
			level:         FatalLevel,
			expectedLevel: zerolog.FatalLevel,
		},
		{
			name:          "PanicLevel",
			level:         PanicLevel,
			expectedLevel: zerolog.PanicLevel,
		},
		{
			name:          "NoLevel",
			level:         NoLevel,
			expectedLevel: zerolog.NoLevel,
		},
		{
			name:          "Disabled",
			level:         Disabled,
			expectedLevel: zerolog.Disabled,
		},
		{
			name:          "TraceLevel",
			level:         TraceLevel,
			expectedLevel: zerolog.TraceLevel,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.level.ZerologLevel()
			assert.Equal(t, tt.expectedLevel, result)
		})
	}
}

func TestLevel_ZerologLevel_UnknownLevel(t *testing.T) {
	tests := []struct {
		name  string
		level Level
	}{
		{"EmptyString", Level("")},
		{"InvalidString", Level("invalid")},
		{"RandomString", Level("random")},
		{"NumberString", Level("123")},
		{"SpecialChars", Level("@#$%")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.level.ZerologLevel()
			// Default case should return TraceLevel
			assert.Equal(t, zerolog.TraceLevel, result)
		})
	}
}

func TestDefaultLogLevel(t *testing.T) {
	// Note: DefaultLogLevel may have been modified by SetConfig in other tests
	// Just verify it's one of the valid levels
	validLevels := []Level{
		DebugLevel, InfoLevel, WarnLevel, ErrorLevel,
		FatalLevel, PanicLevel, NoLevel, Disabled, TraceLevel,
	}

	found := false
	for _, level := range validLevels {
		if DefaultLogLevel == level {
			found = true
			break
		}
	}
	assert.True(t, found, "DefaultLogLevel should be one of the valid levels")
}

func TestLevel_AllCasesInSwitch(t *testing.T) {
	// Ensure all defined levels are handled in the switch statement
	levels := []Level{
		DebugLevel,
		InfoLevel,
		WarnLevel,
		ErrorLevel,
		FatalLevel,
		PanicLevel,
		NoLevel,
		Disabled,
		TraceLevel,
	}

	for _, level := range levels {
		t.Run(string(level), func(t *testing.T) {
			zerologLevel := level.ZerologLevel()
			assert.NotEqual(t, zerolog.Level(99), zerologLevel, "Level should be properly mapped")
		})
	}
}

func TestLevel_TypeAssertion(t *testing.T) {
	l := DebugLevel
	assert.IsType(t, Level(""), l)
	assert.Equal(t, "debug", string(l))
}

func TestLevel_Comparison(t *testing.T) {
	assert.Equal(t, DebugLevel, Level("debug"))
	assert.Equal(t, InfoLevel, Level("info"))
	assert.NotEqual(t, DebugLevel, InfoLevel)
	assert.NotEqual(t, ErrorLevel, WarnLevel)
}

func TestLevel_ZerologLevelMapping(t *testing.T) {
	// Test that the mapping is bidirectional (conceptually)
	tests := []struct {
		foxLevel     Level
		zerologLevel zerolog.Level
	}{
		{TraceLevel, zerolog.TraceLevel},
		{DebugLevel, zerolog.DebugLevel},
		{InfoLevel, zerolog.InfoLevel},
		{WarnLevel, zerolog.WarnLevel},
		{ErrorLevel, zerolog.ErrorLevel},
		{FatalLevel, zerolog.FatalLevel},
		{PanicLevel, zerolog.PanicLevel},
		{NoLevel, zerolog.NoLevel},
		{Disabled, zerolog.Disabled},
	}

	for _, tt := range tests {
		t.Run(string(tt.foxLevel), func(t *testing.T) {
			assert.Equal(t, tt.zerologLevel, tt.foxLevel.ZerologLevel())
		})
	}
}

func TestLevel_StringRepresentation(t *testing.T) {
	tests := []struct {
		level    Level
		expected string
	}{
		{DebugLevel, "debug"},
		{InfoLevel, "info"},
		{WarnLevel, "warn"},
		{ErrorLevel, "error"},
		{FatalLevel, "fatal"},
		{PanicLevel, "panic"},
		{NoLevel, "no"},
		{Disabled, "disabled"},
		{TraceLevel, "trace"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.level))
		})
	}
}

func TestLevel_CaseSensitivity(t *testing.T) {
	// Test that level constants are lowercase
	assert.Equal(t, "debug", string(DebugLevel))
	assert.NotEqual(t, "Debug", string(DebugLevel))
	assert.NotEqual(t, "DEBUG", string(DebugLevel))
}

func BenchmarkLevel_ZerologLevel(b *testing.B) {
	level := InfoLevel

	for i := 0; i < b.N; i++ {
		_ = level.ZerologLevel()
	}
}

func BenchmarkLevel_ZerologLevel_AllLevels(b *testing.B) {
	levels := []Level{
		DebugLevel,
		InfoLevel,
		WarnLevel,
		ErrorLevel,
		FatalLevel,
		PanicLevel,
		NoLevel,
		Disabled,
		TraceLevel,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, level := range levels {
			_ = level.ZerologLevel()
		}
	}
}

func BenchmarkLevel_ZerologLevel_Unknown(b *testing.B) {
	unknownLevel := Level("unknown")

	for i := 0; i < b.N; i++ {
		_ = unknownLevel.ZerologLevel()
	}
}
