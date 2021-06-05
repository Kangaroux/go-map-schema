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

// Tests that TypeNameStartsWithVowel returns the expected result.
func TestTypeNameStartsWithVowel(t *testing.T) {
	tests := []struct {
		name     string
		expected bool
	}{
		{
			name:     "",
			expected: false,
		},
		{
			name:     "a",
			expected: true,
		},
		{
			name:     "e",
			expected: true,
		},
		{
			name:     "i",
			expected: true,
		},
		{
			name:     "o",
			expected: true,
		},
		{
			name:     "u",
			expected: false,
		},
		{
			name:     "int",
			expected: true,
		},
		{
			name:     "*int",
			expected: true,
		},
		{
			name:     "string",
			expected: false,
		},
	}

	for _, test := range tests {
		actual := schema.TypeNameStartsWithVowel(test.name)
		require.Equal(t, test.expected, actual)
	}
}

// Tests that TypeNameStartsWithVowel returns the expected result.
func TestTypeNameWithArticle(t *testing.T) {
	tests := []struct {
		name     string
		expected string
	}{
		{
			name:     "float",
			expected: "a float",
		},
		{
			name:     "int",
			expected: "an int",
		},
		{
			name:     "*string",
			expected: "a *string",
		},
		{
			name:     "*int",
			expected: "an *int",
		},
		{
			name:     "null",
			expected: "null",
		},
		{
			name:     "User",
			expected: "a User",
		},
	}

	for _, test := range tests {
		actual := schema.TypeNameWithArticle(test.name)
		require.Equal(t, test.expected, actual)
	}
}

// Tests that FieldNameWithPath returns the expected result.
func TestFieldNameWithPath(t *testing.T) {

	tests := []struct {
		name     string
		path     []string
		expected string
	}{
		{
			name:     "user",
			expected: "user",
		},
		{
			name:     "cat",
			path:     []string{},
			expected: "cat",
		},
		{
			name:     "email",
			path:     []string{"user"},
			expected: "user.email",
		},
		{
			name:     "is_primary",
			path:     []string{"user", "contacts", "email"},
			expected: "user.contacts.email.is_primary",
		},
	}

	for _, test := range tests {
		actual := schema.FieldNameWithPath(test.name, test.path)
		require.Equal(t, test.expected, actual)
	}
}
