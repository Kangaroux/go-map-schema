package compare_test

import (
	"encoding/json"
	"testing"

	compare "github.com/Kangaroux/go-map-schema"
	"github.com/stretchr/testify/require"
)

type mismatch compare.FieldMismatch

type testStruct struct {
	Foo string
	Bar int
	Baz float64
}

func toJson(val interface{}) string {
	out, err := json.Marshal(val)

	if err != nil {
		panic(err)
	}

	return string(out)
}

// Tests that Compare identifies fields in src that can't be converted
// to the field in dst, due to a type mismatch (e.g. src:string -> dst:int)
func TestCompare_MismatchedFields(t *testing.T) {
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

		r := compare.Compare(&testStruct{}, src)
		require.JSONEq(t, toJson(r.MismatchedFields), toJson(test.expected))
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

		r := compare.Compare(&testStruct{}, src)
		require.ElementsMatch(t, r.MissingFields, test.expected)
	}
}
