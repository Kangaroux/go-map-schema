package compare

import (
	"fmt"
	"reflect"
)

type FieldMismatch struct {
	Field    string
	Expected string
	Actual   string
}

type CompareResults struct {
	MismatchedFields []FieldMismatch
	MissingFields    []string

	dst interface{}
	src map[string]interface{}
}

func (r *CompareResults) Dst() reflect.Type {
	return reflect.ValueOf(r.dst).Elem().Type()
}

func (r *CompareResults) String() string {
	return fmt.Sprintf("CompareResults<missing=%v, dst=%s, src=%v>", r.MissingFields, r.Dst().Name(), r.src)
}

func valueTypeName(v reflect.Value) string {
	if !v.IsValid() {
		return "null"
	}

	return v.Type().Name()
}

// CanConvert returns whether value v is convertible to type t.
func CanConvert(t reflect.Type, v reflect.Value) bool {
	if v.IsValid() {
		return v.Type().ConvertibleTo(t)
	}

	return t.Kind() == reflect.Ptr
}

func Compare(dst interface{}, src map[string]interface{}) *CompareResults {
	v := reflect.ValueOf(dst)
	t := v.Elem().Type()
	results := &CompareResults{
		dst:              dst,
		src:              src,
		MismatchedFields: []FieldMismatch{},
		MissingFields:    []string{},
	}

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)

		if srcField, ok := src[f.Name]; ok {
			srcValue := reflect.ValueOf(srcField)

			if !CanConvert(f.Type, srcValue) {
				mismatch := FieldMismatch{
					Field:    f.Name,
					Expected: f.Type.Name(),
					Actual:   valueTypeName(srcValue),
				}

				results.MismatchedFields = append(results.MismatchedFields, mismatch)
			}
		} else {
			results.MissingFields = append(results.MissingFields, f.Name)
		}
	}

	return results
}
