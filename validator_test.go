package fox

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test DefaultValidator implementation

func TestDefaultValidator_ValidateStruct_ValidStruct(t *testing.T) {
	type validStruct struct {
		Name  string `validate:"required"`
		Email string `validate:"required,email"`
		Age   int    `validate:"gte=0,lte=130"`
	}

	v := &DefaultValidator{}

	tests := []struct {
		name string
		obj  validStruct
	}{
		{
			name: "valid data",
			obj: validStruct{
				Name:  "John Doe",
				Email: "john@example.com",
				Age:   30,
			},
		},
		{
			name: "minimum age",
			obj: validStruct{
				Name:  "Baby",
				Email: "baby@example.com",
				Age:   0,
			},
		},
		{
			name: "maximum age",
			obj: validStruct{
				Name:  "Elder",
				Email: "elder@example.com",
				Age:   130,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateStruct(tt.obj)
			assert.NoError(t, err)
		})
	}
}

func TestDefaultValidator_ValidateStruct_InvalidStruct(t *testing.T) {
	type validStruct struct {
		Name  string `validate:"required"`
		Email string `validate:"required,email"`
		Age   int    `validate:"gte=0,lte=130"`
	}

	v := &DefaultValidator{}

	tests := []struct {
		name        string
		obj         validStruct
		expectError bool
	}{
		{
			name: "missing required name",
			obj: validStruct{
				Name:  "",
				Email: "john@example.com",
				Age:   30,
			},
			expectError: true,
		},
		{
			name: "invalid email format",
			obj: validStruct{
				Name:  "John",
				Email: "not-an-email",
				Age:   30,
			},
			expectError: true,
		},
		{
			name: "age below minimum",
			obj: validStruct{
				Name:  "John",
				Email: "john@example.com",
				Age:   -1,
			},
			expectError: true,
		},
		{
			name: "age above maximum",
			obj: validStruct{
				Name:  "John",
				Email: "john@example.com",
				Age:   131,
			},
			expectError: true,
		},
		{
			name: "missing email",
			obj: validStruct{
				Name:  "John",
				Email: "",
				Age:   30,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateStruct(tt.obj)
			if tt.expectError {
				require.Error(t, err)
				// Verify it's a validator.ValidationErrors
				var validationErr validator.ValidationErrors
				assert.ErrorAs(t, err, &validationErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDefaultValidator_ValidateStruct_Pointer(t *testing.T) {
	type validStruct struct {
		Name string `validate:"required"`
	}

	v := &DefaultValidator{}

	t.Run("pointer to struct", func(t *testing.T) {
		obj := &validStruct{Name: "John"}
		err := v.ValidateStruct(obj)
		assert.NoError(t, err)
	})

	t.Run("pointer to invalid struct", func(t *testing.T) {
		obj := &validStruct{Name: ""}
		err := v.ValidateStruct(obj)
		assert.Error(t, err)
	})
}

func TestDefaultValidator_ValidateStruct_NonStruct(t *testing.T) {
	v := &DefaultValidator{}

	tests := []struct {
		name string
		obj  any
	}{
		{
			name: "string",
			obj:  "test string",
		},
		{
			name: "int",
			obj:  123,
		},
		{
			name: "slice",
			obj:  []string{"a", "b"},
		},
		{
			name: "map",
			obj:  map[string]string{"key": "value"},
		},
		{
			name: "pointer to string",
			obj:  func() *string { s := "test"; return &s }(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateStruct(tt.obj)
			// Non-struct types should not return error
			assert.NoError(t, err)
		})
	}
}

func TestDefaultValidator_ValidateStruct_NestedStruct(t *testing.T) {
	type address struct {
		Street string `validate:"required"`
		City   string `validate:"required"`
	}

	type person struct {
		Name    string  `validate:"required"`
		Address address `validate:"required"`
	}

	v := &DefaultValidator{}

	t.Run("valid nested struct", func(t *testing.T) {
		obj := person{
			Name: "John",
			Address: address{
				Street: "Main St",
				City:   "New York",
			},
		}
		err := v.ValidateStruct(obj)
		assert.NoError(t, err)
	})

	t.Run("invalid nested struct - missing street", func(t *testing.T) {
		obj := person{
			Name: "John",
			Address: address{
				Street: "",
				City:   "New York",
			},
		}
		err := v.ValidateStruct(obj)
		assert.Error(t, err)
	})
}

func TestDefaultValidator_ValidateStruct_SliceField(t *testing.T) {
	type structWithSlice struct {
		Tags []string `validate:"required,min=1"`
	}

	v := &DefaultValidator{}

	t.Run("valid slice", func(t *testing.T) {
		obj := structWithSlice{
			Tags: []string{"go", "testing"},
		}
		err := v.ValidateStruct(obj)
		assert.NoError(t, err)
	})

	t.Run("empty slice", func(t *testing.T) {
		obj := structWithSlice{
			Tags: []string{},
		}
		err := v.ValidateStruct(obj)
		assert.Error(t, err)
	})
}

func TestDefaultValidator_Engine(t *testing.T) {
	v := &DefaultValidator{}

	engine := v.Engine()
	require.NotNil(t, engine)

	// Verify it returns a validator.Validate instance
	validate, ok := engine.(*validator.Validate)
	assert.True(t, ok)
	assert.NotNil(t, validate)

	// Calling Engine multiple times should return the same instance
	engine2 := v.Engine()
	assert.Same(t, engine, engine2)
}

func TestDefaultValidator_LazyInit(t *testing.T) {
	v := &DefaultValidator{}

	// First call should initialize
	engine1 := v.Engine()
	require.NotNil(t, engine1)

	// Second call should return the same instance
	engine2 := v.Engine()
	assert.Same(t, engine1, engine2)

	// ValidateStruct should also work after initialization
	type testStruct struct {
		Name string `validate:"required"`
	}
	err := v.ValidateStruct(testStruct{Name: "test"})
	assert.NoError(t, err)
}

func TestDefaultValidator_ConcurrentAccess(t *testing.T) {
	v := &DefaultValidator{}

	type testStruct struct {
		Name string `validate:"required"`
	}

	// Test concurrent access to trigger lazy initialization
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			err := v.ValidateStruct(testStruct{Name: "test"})
			assert.NoError(t, err)
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestDefaultValidator_CustomValidationTag(t *testing.T) {
	type customStruct struct {
		Code string `validate:"required,len=6"`
	}

	v := &DefaultValidator{}

	t.Run("valid custom tag", func(t *testing.T) {
		obj := customStruct{Code: "ABC123"}
		err := v.ValidateStruct(obj)
		assert.NoError(t, err)
	})

	t.Run("invalid length", func(t *testing.T) {
		obj := customStruct{Code: "ABC"}
		err := v.ValidateStruct(obj)
		assert.Error(t, err)
	})
}

// Test kindOfData function

func TestKindOfData_DirectTypes(t *testing.T) {
	tests := []struct {
		name     string
		data     any
		expected string
	}{
		{
			name:     "struct",
			data:     struct{}{},
			expected: "struct",
		},
		{
			name:     "string",
			data:     "test",
			expected: "string",
		},
		{
			name:     "int",
			data:     42,
			expected: "int",
		},
		{
			name:     "slice",
			data:     []string{},
			expected: "slice",
		},
		{
			name:     "map",
			data:     map[string]string{},
			expected: "map",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kind := kindOfData(tt.data)
			assert.Equal(t, tt.expected, kind.String())
		})
	}
}

func TestKindOfData_PointerTypes(t *testing.T) {
	tests := []struct {
		name     string
		data     any
		expected string
	}{
		{
			name:     "pointer to struct",
			data:     &struct{}{},
			expected: "struct",
		},
		{
			name:     "pointer to string",
			data:     func() *string { s := "test"; return &s }(),
			expected: "string",
		},
		{
			name:     "pointer to int",
			data:     func() *int { i := 42; return &i }(),
			expected: "int",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kind := kindOfData(tt.data)
			assert.Equal(t, tt.expected, kind.String())
		})
	}
}

// Test IsValider interface

type validEntity struct {
	Name string
}

func (v *validEntity) IsValid() error {
	if v.Name == "" {
		return validator.ValidationErrors{}
	}
	return nil
}

type invalidEntity struct {
	Name string
}

func (v *invalidEntity) IsValid() error {
	return validator.ValidationErrors{}
}

func TestIsValider_Interface(t *testing.T) {
	t.Run("valid entity implements IsValider", func(t *testing.T) {
		var _ IsValider = &validEntity{}
	})

	t.Run("valid entity passes validation", func(t *testing.T) {
		entity := &validEntity{Name: "test"}
		err := entity.IsValid()
		assert.NoError(t, err)
	})

	t.Run("invalid entity fails validation", func(t *testing.T) {
		entity := &validEntity{Name: ""}
		err := entity.IsValid()
		assert.Error(t, err)
	})

	t.Run("always invalid entity", func(t *testing.T) {
		entity := &invalidEntity{Name: "test"}
		err := entity.IsValid()
		assert.Error(t, err)
	})
}

// Test global Validate variable

func TestGlobalValidate(t *testing.T) {
	assert.NotNil(t, Validate)

	type testStruct struct {
		Email string `validate:"required,email"`
	}

	t.Run("valid email", func(t *testing.T) {
		obj := testStruct{Email: "test@example.com"}
		err := Validate.Struct(obj)
		assert.NoError(t, err)
	})

	t.Run("invalid email", func(t *testing.T) {
		obj := testStruct{Email: "not-an-email"}
		err := Validate.Struct(obj)
		assert.Error(t, err)
	})
}

// Benchmark tests

func BenchmarkDefaultValidator_ValidateStruct_Simple(b *testing.B) {
	type simpleStruct struct {
		Name  string `validate:"required"`
		Email string `validate:"required,email"`
	}

	v := &DefaultValidator{}
	obj := simpleStruct{
		Name:  "John Doe",
		Email: "john@example.com",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = v.ValidateStruct(obj)
	}
}

func BenchmarkDefaultValidator_ValidateStruct_Complex(b *testing.B) {
	type address struct {
		Street string `validate:"required"`
		City   string `validate:"required"`
		Zip    string `validate:"required,len=5"`
	}

	type person struct {
		Name    string   `validate:"required,min=3,max=50"`
		Email   string   `validate:"required,email"`
		Age     int      `validate:"gte=0,lte=130"`
		Tags    []string `validate:"required,min=1,dive,required"`
		Address address  `validate:"required"`
	}

	v := &DefaultValidator{}
	obj := person{
		Name:  "John Doe",
		Email: "john@example.com",
		Age:   30,
		Tags:  []string{"go", "testing"},
		Address: address{
			Street: "Main St",
			City:   "New York",
			Zip:    "10001",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = v.ValidateStruct(obj)
	}
}

func BenchmarkDefaultValidator_Engine(b *testing.B) {
	v := &DefaultValidator{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = v.Engine()
	}
}

func BenchmarkKindOfData_Struct(b *testing.B) {
	data := struct{ Name string }{"test"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = kindOfData(data)
	}
}

func BenchmarkKindOfData_Pointer(b *testing.B) {
	data := &struct{ Name string }{"test"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = kindOfData(data)
	}
}
