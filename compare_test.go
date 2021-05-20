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

// Tests that Compare identifies fields in src that can't be converted
// to the field in dst, due to a type mismatch (e.g. src:string -> dst:int).
// This only tests "simple" types (no pointers, lists, structs, etc.)
func TestCompare_MismatchedFieldsSimple(t *testing.T) {
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

		r := compare.Compare(&TestStruct{}, src)
		require.JSONEq(t, toJson(r.MismatchedFields), toJson(test.expected), test.srcJson)
	}
}

// Tests that Compare identifies pointer fields and checks if the src value can be
func TestCompare_MismatchedFieldsPtr(t *testing.T) {
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

		r := compare.Compare(&TestStructPtr{}, src)
		require.JSONEq(t, toJson(r.MismatchedFields), toJson(test.expected), test.srcJson)
	}
}

// Tests that Compare identifies the fields in dst that have a json struct tag.
func TestCompare_MismatchedFieldsTags(t *testing.T) {
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

		r := compare.Compare(&TestStructTags{}, src)
		require.JSONEq(t, toJson(r.MismatchedFields), toJson(test.expected), test.srcJson)
	}
}

// Tests that Compare identifies and returns a list of fields that are in
// dst but not src.
func TestCompare_MissingFields(t *testing.T) {
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

		r := compare.Compare(&TestStruct{}, src)
		require.ElementsMatch(t, r.MissingFields, test.expected, test.srcJson)
	}
}

// Tests that Compare identifies and returns a list of fields that are in
// dst but not src, and correctly uses the json field name.
func TestCompare_MissingFieldsTags(t *testing.T) {
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

		r := compare.Compare(&TestStructTags{}, src)
		require.ElementsMatch(t, r.MissingFields, test.expected, test.srcJson)
	}
}
