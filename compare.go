package compare

import (
	"fmt"
	"reflect"
	"strings"
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

func parseField(f reflect.StructField) (name string, ignore bool) {
	tag := f.Tag.Get("json")

	if tag == "" {
		return f.Name, false
	}

	if tag == "-" {
		return "", true
	}

	if i := strings.Index(tag, ","); i != -1 {
		if i == 0 {
			return f.Name, false
		} else {
			return tag[:i], false
		}
	}

	return tag, false
}

func typeNameFromType(t reflect.Type) string {
	if t.Kind() == reflect.Ptr {
		return fmt.Sprintf("*%s", t.Elem().Name())
	}

	return t.Name()
}

func typeNameFromValue(v reflect.Value) string {
	if !v.IsValid() {
		return "null"
	}

	return v.Type().Name()
}

// CanConvert returns whether value v is convertible to type t.
//
// If t is a pointer and v is not nil, it checks if v is convertible to the type that
// t points to.
func CanConvert(t reflect.Type, v reflect.Value) bool {
	isPtr := t.Kind() == reflect.Ptr

	// Check if v is a nil value.
	if !v.IsValid() || (v.CanAddr() && v.IsNil()) {
		return isPtr
	}

	// If the dst is a pointer, check if we can convert to the type it's pointing to.
	if isPtr {
		return t.Elem().ConvertibleTo(v.Type())
	}

	return v.Type().ConvertibleTo(t)
}

func Compare(dst interface{}, src map[string]interface{}) *CompareResults {
	results := &CompareResults{
		dst:              dst,
		src:              src,
		MismatchedFields: []FieldMismatch{},
		MissingFields:    []string{},
	}

	compare(reflect.ValueOf(dst).Elem().Type(), src, results)

	return results
}

func compare(t reflect.Type, src map[string]interface{}, results *CompareResults) {
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		fieldName, skip := parseField(f)

		if skip {
			continue
		}

		// If the field is embedded also check its fields.
		if f.Anonymous {
			compare(f.Type, src, results)
			continue
		}

		if srcField, ok := src[fieldName]; ok {
			srcValue := reflect.ValueOf(srcField)

			if !CanConvert(f.Type, srcValue) {
				mismatch := FieldMismatch{
					Field:    fieldName,
					Expected: typeNameFromType(f.Type),
					Actual:   typeNameFromValue(srcValue),
				}

				results.MismatchedFields = append(results.MismatchedFields, mismatch)
			}
		} else {
			results.MissingFields = append(results.MissingFields, fieldName)
		}
	}
}
