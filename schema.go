package schema

import (
	"errors"
	"fmt"
	"math"
	"reflect"
	"strings"
)

// FieldMismatch represents a type mismatch between a struct field and a map field.
type FieldMismatch struct {
	// Field is the JSON name of the field.
	Field string

	// Expected is the expected type (the type of the dst field).
	Expected string

	// Actual is the actual type (type of the src field).
	Actual string
}

func (f FieldMismatch) String() string {
	return fmt.Sprintf(`expected "%s" to be a %s but it's a %s`, f.Field, f.Expected, f.Actual)
}

// CompareResults contains the results of CompareMapToStruct.
type CompareResults struct {
	// MismatchedFields is a list of fields which have a type mismatch.
	MismatchedFields []FieldMismatch

	// MissingFields is a list of JSON field names which were not in src.
	MissingFields []string
}

/*
CompareMapToStruct takes a pointer to a struct (dst) and a map (src). For each field
in dst, it checks if: the field is in src, and if the types are compatible. The name
of the field is the same as the JSON name.

Fields that have a type mismatch are added to MismatchedFields in the returned
CompareResults. Any fields in dst that are not in src are added to MissingFields.

A type mismatch occurs if a value cannot be converted to a different type without
modifying it, parsing it, etc.

Examples of a type mismatch (src -> dst):
    string -> int
    int    -> string
    bool   -> int
    float  -> int
    null   -> string

Examples of allowed type conversions (src -> dst):
    int  -> float
    <T>  -> *<T>
    null -> *<T>

Embedded structs work as you might expect. The fields of the struct are treated as
if they were hardcoded into dst. In other words, embedding does not change how
src should be structured.
*/
func CompareMapToStruct(dst interface{}, src map[string]interface{}) (*CompareResults, error) {
	v := reflect.ValueOf(dst)

	if !v.IsValid() || v.Kind() != reflect.Ptr {
		return nil, errors.New("dst must be a pointer to a struct")
	} else if v.IsNil() {
		return nil, errors.New("dst must not be nil")
	} else if v.Elem().Kind() != reflect.Struct {
		return nil, errors.New("dst must be a pointer to a struct")
	} else if src == nil {
		return nil, errors.New("src must not be nil")
	}

	results := &CompareResults{
		MismatchedFields: []FieldMismatch{},
		MissingFields:    []string{},
	}

	compare(v.Elem().Type(), src, results)

	return results, nil
}

// canConvert returns whether value v is convertible to type t.
//
// If t is a pointer and v is not nil, it checks if v is convertible to the type that
// t points to.
func canConvert(t reflect.Type, v reflect.Value) bool {
	isPtr := t.Kind() == reflect.Ptr
	dstType := t

	// Check if v is a nil value.
	if !v.IsValid() || (v.CanAddr() && v.IsNil()) {
		return isPtr
	}

	// If the dst is a pointer, check if we can convert to the type it's pointing to.
	if isPtr {
		dstType = t.Elem()
	}

	if !v.Type().ConvertibleTo(dstType) {
		return false
	}

	// Handle converting to an integer type.
	if dstInt, unsigned := isIntegerType(dstType); dstInt {
		if isFloatType(v.Type()) {
			f := v.Float()

			if math.Trunc(f) != f {
				return false
			} else if unsigned && f < 0 {
				return false
			}
		} else if srcInt, _ := isIntegerType(v.Type()); srcInt {
			if unsigned && v.Int() < 0 {
				return false
			}
		}
	}

	return true
}

// compare performs the actual check between the map fields and the struct fields.
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

			if !canConvert(f.Type, srcValue) {
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

func isFloatType(t reflect.Type) (yes bool) {
	switch t.Kind() {
	case reflect.Float32, reflect.Float64:
		yes = true
	}

	return
}

func isIntegerType(t reflect.Type) (yes bool, unsigned bool) {
	switch t.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		yes = true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		yes = true
		unsigned = true
	}

	return
}

// parseField returns the field's JSON name.
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

// typeNameFromType returns the type's name, and handles pointer types.
func typeNameFromType(t reflect.Type) string {
	if t.Kind() == reflect.Ptr {
		return fmt.Sprintf("*%s", t.Elem().Name())
	}

	return t.Name()
}

// typeNameFromValue returns the name of the value's type. If the value is invalid
// (usually from unmarshaling a null) returns "null".
func typeNameFromValue(v reflect.Value) string {
	if !v.IsValid() {
		return "null"
	}

	return v.Type().Name()
}
