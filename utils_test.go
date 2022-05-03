package fox

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilterFlags(t *testing.T) {
	result := filterFlags("text/html ")
	assert.Equal(t, "text/html", result)

	result = filterFlags("text/html;")
	assert.Equal(t, "text/html", result)
}

func TestParseAccept(t *testing.T) {
	parts := parseAccept("text/html , application/xhtml+xml,application/xml;q=0.9,  */* ;q=0.8")
	assert.Len(t, parts, 4)
	assert.Equal(t, "text/html", parts[0])
	assert.Equal(t, "application/xhtml+xml", parts[1])
	assert.Equal(t, "application/xml", parts[2])
	assert.Equal(t, "*/*", parts[3])
}

func TestLastChar(t *testing.T) {
	assert.Equal(t, uint8('a'), lastChar("hola"))
	assert.Equal(t, uint8('s'), lastChar("adios"))
	assert.Panics(t, func() { lastChar("") })
}

func TestJoinPaths(t *testing.T) {
	assert.Equal(t, "", joinPaths("", ""))
	assert.Equal(t, "/", joinPaths("", "/"))
	assert.Equal(t, "/a", joinPaths("/a", ""))
	assert.Equal(t, "/a/", joinPaths("/a/", ""))
	assert.Equal(t, "/a/", joinPaths("/a/", "/"))
	assert.Equal(t, "/a/", joinPaths("/a", "/"))
	assert.Equal(t, "/a/hola", joinPaths("/a", "/hola"))
	assert.Equal(t, "/a/hola", joinPaths("/a/", "/hola"))
	assert.Equal(t, "/a/hola/", joinPaths("/a/", "/hola/"))
	assert.Equal(t, "/a/hola/", joinPaths("/a/", "/hola//"))
}
