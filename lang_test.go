package schema_test

import (
	"reflect"
	"testing"

	schema "github.com/Kangaroux/go-map-schema"
	"github.com/stretchr/testify/require"
)

// Tests that SimpleTypeName returns the expected string.
func TestSimpleTypeName(t *testing.T) {
	tests := []struct {
		val      interface{}
		expected string
	}{
		// uint types
		{
			val:      new(uint),
			expected: "*uint",
		},
		{
			val:      uint(0),
			expected: "uint",
		},
		{
			val:      uint8(0),
			expected: "uint",
		},
		{
			val:      uint16(0),
			expected: "uint",
		},
		{
			val:      uint32(0),
			expected: "uint",
		},
		{
			val:      uint64(0),
			expected: "uint",
		},

		// int types
		{
			val:      new(int),
			expected: "*int",
		},
		{
			val:      int(0),
			expected: "int",
		},
		{
			val:      int8(0),
			expected: "int",
		},
		{
			val:      int16(0),
			expected: "int",
		},
		{
			val:      int32(0),
			expected: "int",
		},
		{
			val:      int64(0),
			expected: "int",
		},

		// float types
		{
			val:      new(float32),
			expected: "*float",
		},
		{
			val:      float32(0),
			expected: "float",
		},
		{
			val:      float64(0),
			expected: "float",
		},

		// other
		{
			val:      "hello",
			expected: "string",
		},
		{
			val:      &TestStruct{},
			expected: "*TestStruct",
		},
	}

	for _, test := range tests {
		name := schema.SimpleTypeName(reflect.TypeOf(test.val))
		require.Equal(t, test.expected, name)
	}
}
