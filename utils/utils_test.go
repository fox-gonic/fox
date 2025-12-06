package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNameOfFunction(t *testing.T) {
	// Test with a known function
	name := NameOfFunction(TestNameOfFunction)
	assert.Contains(t, name, "TestNameOfFunction")
	assert.Contains(t, name, "github.com/fox-gonic/fox/utils")
}

func TestNameOfFunction_AnonymousFunction(t *testing.T) {
	// Test with anonymous function
	anonFunc := func() {}
	name := NameOfFunction(anonFunc)
	assert.NotEmpty(t, name)
	assert.Contains(t, name, "utils")
}

func TestNameOfFunction_BuiltinFunction(t *testing.T) {
	// Test with a standard library function
	name := NameOfFunction(assert.Equal)
	assert.NotEmpty(t, name)
	assert.Contains(t, name, "Equal")
}

func TestJoinPaths(t *testing.T) {
	tests := []struct {
		name         string
		absolutePath string
		relativePath string
		expected     string
	}{
		{
			name:         "empty relative path",
			absolutePath: "/api",
			relativePath: "",
			expected:     "/api",
		},
		{
			name:         "simple join",
			absolutePath: "/api",
			relativePath: "v1",
			expected:     "/api/v1",
		},
		{
			name:         "relative path with leading slash",
			absolutePath: "/api",
			relativePath: "/v1",
			expected:     "/api/v1",
		},
		{
			name:         "relative path with trailing slash",
			absolutePath: "/api",
			relativePath: "v1/",
			expected:     "/api/v1/",
		},
		{
			name:         "both paths with slashes",
			absolutePath: "/api/",
			relativePath: "/v1/",
			expected:     "/api/v1/",
		},
		{
			name:         "multiple segments",
			absolutePath: "/api",
			relativePath: "v1/users",
			expected:     "/api/v1/users",
		},
		{
			name:         "multiple segments with trailing slash",
			absolutePath: "/api",
			relativePath: "v1/users/",
			expected:     "/api/v1/users/",
		},
		{
			name:         "dot segments",
			absolutePath: "/api",
			relativePath: "./v1",
			expected:     "/api/v1",
		},
		{
			name:         "parent directory",
			absolutePath: "/api/v1",
			relativePath: "../v2",
			expected:     "/api/v2",
		},
		{
			name:         "single slash relative",
			absolutePath: "/api",
			relativePath: "/",
			expected:     "/api/",
		},
		{
			name:         "empty absolute path",
			absolutePath: "",
			relativePath: "v1",
			expected:     "v1",
		},
		{
			name:         "empty absolute path with trailing slash",
			absolutePath: "",
			relativePath: "v1/",
			expected:     "v1/",
		},
		{
			name:         "root path",
			absolutePath: "/",
			relativePath: "api",
			expected:     "/api",
		},
		{
			name:         "root path with trailing slash in relative",
			absolutePath: "/",
			relativePath: "api/",
			expected:     "/api/",
		},
		{
			name:         "complex path",
			absolutePath: "/app/api/v1",
			relativePath: "users/profile/",
			expected:     "/app/api/v1/users/profile/",
		},
		{
			name:         "path cleaning",
			absolutePath: "/api//v1",
			relativePath: "users//profile",
			expected:     "/api/v1/users/profile",
		},
		{
			name:         "path cleaning with trailing slash",
			absolutePath: "/api//v1",
			relativePath: "users//profile/",
			expected:     "/api/v1/users/profile/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := JoinPaths(tt.absolutePath, tt.relativePath)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestJoinPaths_TrailingSlashPreservation(t *testing.T) {
	// Specifically test that trailing slashes in relative paths are preserved
	testCases := []struct {
		absolute string
		relative string
		hasSlash bool
	}{
		{"/api", "v1/", true},
		{"/api", "v1", false},
		{"/api/", "v1/", true},
		{"/api/", "v1", false},
	}

	for _, tc := range testCases {
		result := JoinPaths(tc.absolute, tc.relative)
		if tc.hasSlash {
			assert.Equal(t, uint8('/'), result[len(result)-1],
				"Expected trailing slash for absolute=%s, relative=%s", tc.absolute, tc.relative)
		} else {
			if len(result) > 0 {
				assert.NotEqual(t, uint8('/'), result[len(result)-1],
					"Did not expect trailing slash for absolute=%s, relative=%s", tc.absolute, tc.relative)
			}
		}
	}
}

func TestJoinPaths_EdgeCases(t *testing.T) {
	// Test edge cases
	t.Run("only slashes", func(t *testing.T) {
		result := JoinPaths("/", "/")
		assert.Equal(t, "/", result)
	})

	t.Run("multiple slashes in relative", func(t *testing.T) {
		result := JoinPaths("/api", "///v1///")
		assert.Equal(t, "/api/v1/", result)
	})

	t.Run("relative path is just slash", func(t *testing.T) {
		result := JoinPaths("/api", "/")
		assert.Equal(t, "/api/", result)
	})
}

// Benchmark tests
func BenchmarkNameOfFunction(b *testing.B) {
	fn := TestNameOfFunction
	for i := 0; i < b.N; i++ {
		_ = NameOfFunction(fn)
	}
}

func BenchmarkJoinPaths(b *testing.B) {
	absolutePath := "/api/v1"
	relativePath := "users/profile/"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = JoinPaths(absolutePath, relativePath)
	}
}

func BenchmarkJoinPaths_EmptyRelative(b *testing.B) {
	absolutePath := "/api/v1"
	relativePath := ""

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = JoinPaths(absolutePath, relativePath)
	}
}

func BenchmarkJoinPaths_WithTrailingSlash(b *testing.B) {
	absolutePath := "/api"
	relativePath := "v1/"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = JoinPaths(absolutePath, relativePath)
	}
}

// Helper function tests
func helperFunction() string {
	return "helper"
}

func TestNameOfFunction_Helper(t *testing.T) {
	name := NameOfFunction(helperFunction)
	assert.Contains(t, name, "helperFunction")
}

type testStruct struct{}

func (ts *testStruct) String() string {
	return "test"
}

func TestNameOfFunction_Method(t *testing.T) {
	ts := &testStruct{}

	// Test with method reference
	name := NameOfFunction(ts.String)
	assert.NotEmpty(t, name)
}

func TestJoinPaths_ConsistentOutput(t *testing.T) {
	// Test that multiple calls with same input produce same output
	absolutePath := "/api/v1"
	relativePath := "users/"

	result1 := JoinPaths(absolutePath, relativePath)
	result2 := JoinPaths(absolutePath, relativePath)
	result3 := JoinPaths(absolutePath, relativePath)

	assert.Equal(t, result1, result2)
	assert.Equal(t, result2, result3)
}

func TestJoinPaths_NoModificationOfInputs(t *testing.T) {
	// Ensure the function doesn't modify the input strings
	absolutePath := "/api"
	relativePath := "v1/"

	absolutePathCopy := absolutePath
	relativePathCopy := relativePath

	_ = JoinPaths(absolutePath, relativePath)

	assert.Equal(t, absolutePathCopy, absolutePath)
	assert.Equal(t, relativePathCopy, relativePath)
}
