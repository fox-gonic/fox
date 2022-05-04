package fox

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestError(t *testing.T) {

	var err interface{}

	err = Error{}
	errType := reflect.ValueOf(err).Type()

	fmt.Println(errType == errorType)

	err = &Error{}
	errType = reflect.Indirect(reflect.ValueOf(err)).Type()

	assert.True(t, errType == errorType)
}
