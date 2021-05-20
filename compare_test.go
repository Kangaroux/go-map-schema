package compare_test

import (
	"encoding/json"
	"testing"

	compare "github.com/Kangaroux/go-map-schema"
	"github.com/stretchr/testify/require"
)

type testStruct struct {
	Foo string
	Bar int
	Baz float64
}

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
