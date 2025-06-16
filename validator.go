package fox

import (
	"reflect"
	"sync"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// Validate global validator
var Validate = validator.New()

// DefaultValidator is the default implementation of Validator.
type DefaultValidator struct {
	once     sync.Once
	validate *validator.Validate
}

var _ binding.StructValidator = &DefaultValidator{}

// ValidateStruct check struct
func (v *DefaultValidator) ValidateStruct(obj any) error {
	if kindOfData(obj) == reflect.Struct {

		v.lazyinit()

		if err := v.validate.Struct(obj); err != nil {
			return error(err)
		}
	}

	return nil
}

// Engine return validate
func (v *DefaultValidator) Engine() any {
	v.lazyinit()
	return v.validate
}

func (v *DefaultValidator) lazyinit() {
	v.once.Do(func() {
		v.validate = validator.New()

		// add any custom validations etc. here
	})
}

func kindOfData(data any) reflect.Kind {
	value := reflect.ValueOf(data)
	valueType := value.Kind()

	if valueType == reflect.Ptr {
		valueType = value.Elem().Kind()
	}
	return valueType
}

// IsValider interface
type IsValider interface {
	IsValid() error
}
