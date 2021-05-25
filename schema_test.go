package schema_test

import (
	"encoding/json"
	"reflect"
	"testing"

	schema "github.com/Kangaroux/go-map-schema"
	"github.com/stretchr/testify/require"
)

type mismatch schema.FieldMismatch

type TestStruct struct {
	Foo string
	Bar int
	Baz float64
}

type TestStructEmbedded struct {
	TestStruct
	Butt bool
}

type TestStructPtr struct {
	Ptr *string
}

type TestStructTags struct {
	LowercaseA  string `json:"a"`
	IgnoreMe    string `json:"-"`
	WithOptions string `json:",omitempty"`
	Hyphen      string `json:"-,"`
}

type TestStructUnsigned struct {
	Foo uint
}

func toJson(val interface{}) string {
	out, err := json.Marshal(val)

	if err != nil {
		panic(err)
	}

	return string(out)
}

// Tests that CompareMapToStruct returns an error if the dst isn't valid.
func TestCompareMapToStruct_BadDstErrors(t *testing.T) {
	var err error
	m := make(map[string]interface{})
	v := "hello"

	_, err = schema.CompareMapToStruct(123, m, nil)
	require.Equal(t, schema.ErrInvalidDst, err)

	_, err = schema.CompareMapToStruct(v, m, nil)
	require.Equal(t, schema.ErrInvalidDst, err)

	_, err = schema.CompareMapToStruct(&v, m, nil)
	require.Equal(t, schema.ErrInvalidDst, err)

	_, err = schema.CompareMapToStruct(nil, m, nil)
	require.Equal(t, schema.ErrInvalidDst, err)

	_, err = schema.CompareMapToStruct(TestStruct{}, m, nil)
	require.Equal(t, schema.ErrInvalidDst, err)

	_, err = schema.CompareMapToStruct(&TestStruct{}, m, nil)
	require.NoError(t, err)
}

// Tests that CompareMapToStruct returns an error if the src isn't valid.
func TestCompareMapToStruct_BadSrcErrors(t *testing.T) {
	var err error

	_, err = schema.CompareMapToStruct(&TestStruct{}, nil, nil)
	require.Equal(t, schema.ErrNilSrc, err)

	_, err = schema.CompareMapToStruct(&TestStruct{}, make(map[string]interface{}), nil)
	require.NoError(t, err)
}

// Tests that CompareMapToStruct uses the provided functions in the compare options.
func TestCompareMapToStruct_CompareOptsUsesProvidedFuncs(t *testing.T) {
	convertibleCalled := false
	convertibleFunc := func(t reflect.Type, v reflect.Value) bool { convertibleCalled = true; return false }

	typeNameCalled := false
	typeNameFunc := func(t reflect.Type) string { typeNameCalled = true; return "" }

	src := make(map[string]interface{})
	json.Unmarshal([]byte(`{"Foo":""}`), &src)

	opts := &schema.CompareOpts{
		ConvertibleFunc: convertibleFunc,
		TypeNameFunc:    typeNameFunc,
	}

	schema.CompareMapToStruct(&TestStruct{}, src, opts)

	require.True(t, convertibleCalled)
	require.True(t, typeNameCalled)
}

// Tests that CompareMapToStruct sets defaults if it receives a compare options instance
// but one or more of the functions are nil.
func TestCompareMapToStruct_CompareOptsSetsDefaults(t *testing.T) {
	srcJson := `{"Foo":true,"Baz":""}`
	expected := []mismatch{
		{
			Field:    "Foo",
			Expected: "string",
			Actual:   "bool",
		},
		{
			Field:    "Baz",
			Expected: "float64",
			Actual:   "string",
		},
	}

	src := make(map[string]interface{})
	json.Unmarshal([]byte(srcJson), &src)

	r, _ := schema.CompareMapToStruct(&TestStruct{}, src, &schema.CompareOpts{})
	require.JSONEq(t, toJson(expected), toJson(r.MismatchedFields))
}

// Tests that CompareMapToStruct identifies fields in src that can't be converted
// to the field in dst, due to a type mismatch (e.g. src:string -> dst:int).
// This only tests "simple" types (no pointers, lists, structs, etc.)
func TestCompareMapToStruct_MismatchedFieldsSimple(t *testing.T) {
	tests := []struct {
		srcJson  string
		expected []mismatch
	}{
		{
			srcJson:  `{}`,
			expected: []mismatch{},
		},
		{
			srcJson:  `{"Foo":"","Bar":0,"Baz":3.14}`,
			expected: []mismatch{},
		},
		{
			srcJson: `{"Foo":null}`,
			expected: []mismatch{
				{
					Field:    "Foo",
					Expected: "string",
					Actual:   "null",
				},
			},
		},
		{
			srcJson: `{"Foo":0}`,
			expected: []mismatch{
				{
					Field:    "Foo",
					Expected: "string",
					Actual:   "float64",
				},
			},
		},
		{
			srcJson: `{"Bar":"hi"}`,
			expected: []mismatch{
				{
					Field:    "Bar",
					Expected: "int",
					Actual:   "string",
				},
			},
		},
		{
			srcJson: `{"Bar":1.23}`,
			expected: []mismatch{
				{
					Field:    "Bar",
					Expected: "int",
					Actual:   "float64",
				},
			},
		},
		{
			srcJson: `{"Foo":true,"Baz":""}`,
			expected: []mismatch{
				{
					Field:    "Foo",
					Expected: "string",
					Actual:   "bool",
				},
				{
					Field:    "Baz",
					Expected: "float64",
					Actual:   "string",
				},
			},
		},
	}

	for _, test := range tests {
		// Unmarshal the json into a map.
		src := make(map[string]interface{})
		json.Unmarshal([]byte(test.srcJson), &src)

		r, _ := schema.CompareMapToStruct(&TestStruct{}, src, nil)
		require.JSONEq(t, toJson(test.expected), toJson(r.MismatchedFields), test.srcJson)
	}
}

// Tests that CompareMapToStruct treats embedded structs as if all the embedded fields
// were moved into the parent struct.
func TestCompareMapToStruct_MismatchedFieldsEmbedded(t *testing.T) {
	tests := []struct {
		srcJson  string
		expected []mismatch
	}{
		{
			srcJson:  `{}`,
			expected: []mismatch{},
		},
		{
			srcJson:  `{"Foo":"","Bar":0,"Baz":3.14,"Butt":false}`,
			expected: []mismatch{},
		},
		{
			srcJson: `{"Foo":null}`,
			expected: []mismatch{
				{
					Field:    "Foo",
					Expected: "string",
					Actual:   "null",
				},
			},
		},
		{
			srcJson: `{"Bar":"hi"}`,
			expected: []mismatch{
				{
					Field:    "Bar",
					Expected: "int",
					Actual:   "string",
				},
			},
		},
		{
			srcJson: `{"Foo":true,"Baz":""}`,
			expected: []mismatch{
				{
					Field:    "Foo",
					Expected: "string",
					Actual:   "bool",
				},
				{
					Field:    "Baz",
					Expected: "float64",
					Actual:   "string",
				},
			},
		},
		{
			srcJson: `{"Butt":"hi"}`,
			expected: []mismatch{
				{
					Field:    "Butt",
					Expected: "bool",
					Actual:   "string",
				},
			},
		},
	}

	for _, test := range tests {
		// Unmarshal the json into a map.
		src := make(map[string]interface{})
		json.Unmarshal([]byte(test.srcJson), &src)

		r, _ := schema.CompareMapToStruct(&TestStructEmbedded{}, src, nil)
		require.JSONEq(t, toJson(test.expected), toJson(r.MismatchedFields), test.srcJson)
	}
}

// Tests that CompareMapToStruct identifies pointer fields and checks if the src value can be
func TestCompareMapToStruct_MismatchedFieldsPtr(t *testing.T) {
	tests := []struct {
		srcJson  string
		expected []mismatch
	}{
		{
			srcJson:  `{}`,
			expected: []mismatch{},
		},
		{
			srcJson:  `{"Ptr":null}`,
			expected: []mismatch{},
		},
		{
			srcJson: `{"Ptr":0}`,
			expected: []mismatch{
				{
					Field:    "Ptr",
					Expected: "*string",
					Actual:   "float64",
				},
			},
		},
		{
			srcJson:  `{"Ptr":"hi"}`,
			expected: []mismatch{},
		},
	}

	for _, test := range tests {
		// Unmarshal the json into a map.
		src := make(map[string]interface{})
		json.Unmarshal([]byte(test.srcJson), &src)

		r, _ := schema.CompareMapToStruct(&TestStructPtr{}, src, nil)
		require.JSONEq(t, toJson(test.expected), toJson(r.MismatchedFields), test.srcJson)
	}
}

// Tests that CompareMapToStruct identifies the fields in dst that have a json struct tag.
func TestCompareMapToStruct_MismatchedFieldsTags(t *testing.T) {
	tests := []struct {
		srcJson  string
		expected []mismatch
	}{
		{
			srcJson:  `{}`,
			expected: []mismatch{},
		},
		{
			srcJson: `{"a":0}`,
			expected: []mismatch{
				{
					Field:    "a",
					Expected: "string",
					Actual:   "float64",
				},
			},
		},
		{
			srcJson:  `{"IgnoreMe":0}`,
			expected: []mismatch{},
		},
		{
			srcJson: `{"WithOptions":0}`,
			expected: []mismatch{
				{
					Field:    "WithOptions",
					Expected: "string",
					Actual:   "float64",
				},
			},
		},
		{
			srcJson: `{"-":0}`,
			expected: []mismatch{
				{
					Field:    "-",
					Expected: "string",
					Actual:   "float64",
				},
			},
		},
	}

	for _, test := range tests {
		// Unmarshal the json into a map.
		src := make(map[string]interface{})
		json.Unmarshal([]byte(test.srcJson), &src)

		r, _ := schema.CompareMapToStruct(&TestStructTags{}, src, nil)
		require.JSONEq(t, toJson(test.expected), toJson(r.MismatchedFields), test.srcJson)
	}
}

// Tests that CompareMapToStruct identifies negative numbers when the dst type
// is unsigned.
func TestCompareMapToStruct_MismatchedFieldsUnsigned(t *testing.T) {
	tests := []struct {
		srcJson  string
		expected []mismatch
	}{
		{
			srcJson:  `{}`,
			expected: []mismatch{},
		},
		{
			srcJson:  `{"Foo":0}`,
			expected: []mismatch{},
		},
		{
			srcJson:  `{"Foo":1}`,
			expected: []mismatch{},
		},
		{
			srcJson: `{"Foo":-1}`,
			expected: []mismatch{
				{
					Field:    "Foo",
					Expected: "uint",
					Actual:   "float64",
				},
			},
		},
		{
			srcJson: `{"Foo":1.5}`,
			expected: []mismatch{
				{
					Field:    "Foo",
					Expected: "uint",
					Actual:   "float64",
				},
			},
		},
	}

	for _, test := range tests {
		// Unmarshal the json into a map.
		src := make(map[string]interface{})
		json.Unmarshal([]byte(test.srcJson), &src)

		r, _ := schema.CompareMapToStruct(&TestStructUnsigned{}, src, nil)
		require.JSONEq(t, toJson(test.expected), toJson(r.MismatchedFields), test.srcJson)
	}
}

// Tests that CompareMapToStruct identifies and returns a list of fields that are in
// dst but not src.
func TestCompareMapToStruct_MissingFields(t *testing.T) {
	tests := []struct {
		srcJson  string
		expected []string
	}{
		{
			srcJson:  `{}`,
			expected: []string{"Foo", "Bar", "Baz"},
		},
		{
			srcJson:  `{"Foo":""}`,
			expected: []string{"Bar", "Baz"},
		},
		{
			srcJson:  `{"Foo":"","Bar":0}`,
			expected: []string{"Baz"},
		},
		{
			srcJson:  `{"Foo":"","Bar":0,"Baz":3.14}`,
			expected: []string{},
		},
	}

	for _, test := range tests {
		// Unmarshal the json into a map.
		src := make(map[string]interface{})
		json.Unmarshal([]byte(test.srcJson), &src)

		r, _ := schema.CompareMapToStruct(&TestStruct{}, src, nil)
		require.ElementsMatch(t, test.expected, r.MissingFields, test.srcJson)
	}
}

// Tests that CompareMapToStruct identifies and returns a list of fields that are in
// dst but not src, including embedded fields.
func TestCompareMapToStruct_MissingFieldsEmbedded(t *testing.T) {
	tests := []struct {
		srcJson  string
		expected []string
	}{
		{
			srcJson:  `{}`,
			expected: []string{"Foo", "Bar", "Baz", "Butt"},
		},
		{
			srcJson:  `{"Foo":""}`,
			expected: []string{"Bar", "Baz", "Butt"},
		},
		{
			srcJson:  `{"Foo":"","Bar":0}`,
			expected: []string{"Baz", "Butt"},
		},
		{
			srcJson:  `{"Foo":"","Bar":0,"Baz":3.14}`,
			expected: []string{"Butt"},
		},
		{
			srcJson:  `{"Foo":"","Bar":0,"Baz":3.14,"Butt":false}`,
			expected: []string{},
		},
	}

	for _, test := range tests {
		// Unmarshal the json into a map.
		src := make(map[string]interface{})
		json.Unmarshal([]byte(test.srcJson), &src)

		r, _ := schema.CompareMapToStruct(&TestStructEmbedded{}, src, nil)
		require.ElementsMatch(t, test.expected, r.MissingFields, test.srcJson)
	}
}

// Tests that CompareMapToStruct identifies and returns a list of fields that are in
// dst but not src, and correctly uses the json field name.
func TestCompareMapToStruct_MissingFieldsTags(t *testing.T) {
	tests := []struct {
		srcJson  string
		expected []string
	}{
		{
			srcJson:  `{}`,
			expected: []string{"a", "-", "WithOptions"},
		},
		{
			srcJson:  `{"a":""}`,
			expected: []string{"-", "WithOptions"},
		},
		{
			srcJson:  `{"-":""}`,
			expected: []string{"a", "WithOptions"},
		},
		{
			srcJson:  `{"WithOptions":""}`,
			expected: []string{"a", "-"},
		},
	}

	for _, test := range tests {
		// Unmarshal the json into a map.
		src := make(map[string]interface{})
		json.Unmarshal([]byte(test.srcJson), &src)

		r, _ := schema.CompareMapToStruct(&TestStructTags{}, src, nil)
		require.ElementsMatch(t, test.expected, r.MissingFields, test.srcJson)
	}
}

// Tests that Errors returns the expected error map.
func TestCompareResults_Errors(t *testing.T) {
	tests := []struct {
		srcJson  string
		expected error
	}{
		{
			srcJson: `{"Foo":null}`,
			expected: schema.MismatchError(map[string]interface{}{
				"Foo": `expected a string but it's null`,
			}),
		},
		{
			srcJson: `{"Foo":1.23,"Bar":true}`,
			expected: schema.MismatchError(map[string]interface{}{
				"Foo": `expected a string but it's a float64`,
				"Bar": `expected an int but it's a bool`,
			}),
		},
	}

	for _, test := range tests {
		// Unmarshal the json into a map.
		src := make(map[string]interface{})
		json.Unmarshal([]byte(test.srcJson), &src)

		r, _ := schema.CompareMapToStruct(&TestStruct{}, src, nil)

		require.Equal(t, test.expected, r.Errors(), test.srcJson)

		// Test marshaling the error to JSON.
		require.JSONEq(t, toJson(test.expected), toJson(r.Errors()), test.srcJson)
	}
}

// Tests that Errors returns nil when there are no type mismatches.
func TestCompareResults_ErrorsReturnsNil(t *testing.T) {
	tests := []struct {
		srcJson string
	}{
		{
			srcJson: `{}`,
		},
		{
			srcJson: `{"Foo":"hi"}`,
		},
		{
			srcJson: `{"Foo":"hi","Bar":1,"Baz":3.14}`,
		},
	}

	for _, test := range tests {
		// Unmarshal the json into a map.
		src := make(map[string]interface{})
		json.Unmarshal([]byte(test.srcJson), &src)

		r, _ := schema.CompareMapToStruct(&TestStruct{}, src, nil)

		require.Nil(t, r.Errors())
	}
}
