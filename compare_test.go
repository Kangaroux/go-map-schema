package compare_test

import (
	"encoding/json"
	"testing"

	compare "github.com/Kangaroux/go-map-schema"
	"github.com/stretchr/testify/require"
)

type mismatch compare.FieldMismatch

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

	_, err = compare.CompareMapToStruct(123, m)
	require.Error(t, err)

	_, err = compare.CompareMapToStruct("hello", m)
	require.Error(t, err)

	_, err = compare.CompareMapToStruct(nil, m)
	require.Error(t, err)

	_, err = compare.CompareMapToStruct(TestStruct{}, m)
	require.Error(t, err)

	_, err = compare.CompareMapToStruct(&TestStruct{}, m)
	require.NoError(t, err)
}

// Tests that CompareMapToStruct returns an error if the src isn't valid.
func TestCompareMapToStruct_BadSrcErrors(t *testing.T) {
	var err error

	_, err = compare.CompareMapToStruct(&TestStruct{}, nil)
	require.Error(t, err)

	_, err = compare.CompareMapToStruct(&TestStruct{}, make(map[string]interface{}))
	require.NoError(t, err)
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
			srcJson: `{"Bar":"hi"}`,
			expected: []mismatch{
				{
					Field:    "Bar",
					Expected: "int",
					Actual:   "string",
				},
			},
		},
		// TODO: When converting to an int, check to see if the src value would be truncated.
		// If the value is modified as part of the conversion that should be a type mismatch.
		// {
		// 	srcJson: `{"Bar":1.23}`,
		// 	expected: []mismatch{
		// 		{
		// 			Field:    "Bar",
		// 			Expected: "int",
		// 			Actual:   "float64",
		// 		},
		// 	},
		// },
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

		r, _ := compare.CompareMapToStruct(&TestStruct{}, src)
		require.JSONEq(t, toJson(r.MismatchedFields), toJson(test.expected), test.srcJson)
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

		r, _ := compare.CompareMapToStruct(&TestStructEmbedded{}, src)
		require.JSONEq(t, toJson(r.MismatchedFields), toJson(test.expected), test.srcJson)
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

		r, _ := compare.CompareMapToStruct(&TestStructPtr{}, src)
		require.JSONEq(t, toJson(r.MismatchedFields), toJson(test.expected), test.srcJson)
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

		r, _ := compare.CompareMapToStruct(&TestStructTags{}, src)
		require.JSONEq(t, toJson(r.MismatchedFields), toJson(test.expected), test.srcJson)
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

		r, _ := compare.CompareMapToStruct(&TestStruct{}, src)
		require.ElementsMatch(t, r.MissingFields, test.expected, test.srcJson)
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

		r, _ := compare.CompareMapToStruct(&TestStructEmbedded{}, src)
		require.ElementsMatch(t, r.MissingFields, test.expected, test.srcJson)
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

		r, _ := compare.CompareMapToStruct(&TestStructTags{}, src)
		require.ElementsMatch(t, r.MissingFields, test.expected, test.srcJson)
	}
}
