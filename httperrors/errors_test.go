package httperrors

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestError(t *testing.T) {

	assert := assert.New(t)

	err := New(404, "not found")
	assert.Equal(404, err.HTTPCode)
	assert.Equal("(404): not found", err.Error())
}
