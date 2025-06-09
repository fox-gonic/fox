package utils

import (
	"strings"
	"testing"
)

func TestLastChar(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		want      uint8
		wantPanic bool
	}{
		{
			name:      "normal string",
			input:     "hello",
			want:      'o',
			wantPanic: false,
		},
		{
			name:      "single character",
			input:     "a",
			want:      'a',
			wantPanic: false,
		},
		{
			name:      "empty string",
			input:     "",
			want:      0,
			wantPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				r := recover()
				if (r != nil) != tt.wantPanic {
					t.Errorf("lastChar() panic = %v, wantPanic %v", r, tt.wantPanic)
				}
			}()

			if got := lastChar(tt.input); got != tt.want {
				t.Errorf("lastChar() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNameOfFunction(t *testing.T) {
	// Define a test function
	testFunc := func() {}

	// Get the actual function name
	actualName := NameOfFunction(testFunc)

	t.Logf("actualName: %v", actualName)

	// Check if the name contains the expected parts
	if !strings.Contains(actualName, "TestNameOfFunction") {
		t.Errorf("NameOfFunction() = %v, want name containing 'TestNameOfFunction'", actualName)
	}
}

func TestJoinPaths(t *testing.T) {
	tests := []struct {
		name         string
		absolutePath string
		relativePath string
		want         string
	}{
		{
			name:         "empty relative path",
			absolutePath: "/api",
			relativePath: "",
			want:         "/api",
		},
		{
			name:         "normal join",
			absolutePath: "/api",
			relativePath: "users",
			want:         "/api/users",
		},
		{
			name:         "relative path with trailing slash",
			absolutePath: "/api",
			relativePath: "users/",
			want:         "/api/users/",
		},
		{
			name:         "both paths with trailing slash",
			absolutePath: "/api/",
			relativePath: "users/",
			want:         "/api/users/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := JoinPaths(tt.absolutePath, tt.relativePath); got != tt.want {
				t.Errorf("JoinPaths() = %v, want %v", got, tt.want)
			}
		})
	}
}
