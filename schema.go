package schema

import (
	"errors"
	"fmt"
	"math"
	"reflect"
	"strings"
)

// CompareResults contains the results of CompareMapToStruct.
type CompareResults struct {
	// MismatchedFields is a list of fields which have a type mismatch.
	MismatchedFields []FieldMismatch

	// MissingFields is a list of JSON field names which were not in src.
	MissingFields []string
}

// AsMap converts the mismatched fields to a map and returns it.
func (cr *CompareResults) AsMap() map[string]interface{} {
	m := make(map[string]interface{})

	for _, f := range cr.MismatchedFields {
		m[f.Field] = f.String()
	}

	return m
}

// FieldMismatch represents a type mismatch between a struct field and a map field.
type FieldMismatch struct {
	// Field is the JSON name of the field.
	Field string

	// Expected is the expected type (the type of the dst field).
	Expected string

	// Actual is the actual type (type of the src field).
	Actual string
}

// String returns a user friendly message explaining the type mismatch.
func (f FieldMismatch) String() string {
	return fmt.Sprintf(
		`expected "%s" to be %s but it's %s`,
		f.Field,
		typeNameWithArticle(f.Expected),
		typeNameWithArticle(f.Actual),
	)
}

// CompareOpts can be used to configure how CompareMapToStruct works.
type CompareOpts struct {
	// ConvertibleFunc is the function used to check if a value can be safely converted
	// to a type.
	ConvertibleFunc ConvertibleFunc

	// TypeNameFunc is the function used to convert a type into a string.
	TypeNameFunc TypeNameFunc
}

// ConvertibleFunc takes a dst type (t) and a src value (v) and returns true if
// v is convertible to t.
type ConvertibleFunc func(t reflect.Type, v reflect.Value) bool

// TypeNameFunc takes a reflection type and returns its name as a string.
type TypeNameFunc func(t reflect.Type) string

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
func CompareMapToStruct(dst interface{}, src map[string]interface{}, opts *CompareOpts) (*CompareResults, error) {
	if opts == nil {
		opts = &CompareOpts{
			ConvertibleFunc: DefaultCanConvert,
			TypeNameFunc:    TypeNameDetailed,
		}
	} else {
		// Create a copy of the options since we might need to modify it.
		opts = &CompareOpts{
			ConvertibleFunc: opts.ConvertibleFunc,
			TypeNameFunc:    opts.TypeNameFunc,
		}

		if opts.ConvertibleFunc == nil {
			opts.ConvertibleFunc = DefaultCanConvert
		}
		if opts.TypeNameFunc == nil {
			opts.TypeNameFunc = TypeNameDetailed
		}
	}

	v := reflect.ValueOf(dst)

	if !v.IsValid() || v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return nil, errors.New("dst must be a pointer to a struct")
	} else if src == nil {
		return nil, errors.New("src must not be nil")
	}

	results := &CompareResults{
		MismatchedFields: []FieldMismatch{},
		MissingFields:    []string{},
	}

	compare(v.Elem().Type(), src, opts, results)

	return results, nil
}

// DefaultCanConvert returns whether value v is convertible to type t.
//
// If t is a pointer and v is not nil, it checks if v is convertible to the type that
// t points to.
func DefaultCanConvert(t reflect.Type, v reflect.Value) bool {
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

// TypeNameDetailed takes a type and returns it verbatim.
func TypeNameDetailed(t reflect.Type) string {
	if t.Kind() == reflect.Ptr {
		return fmt.Sprintf("*%s", t.Elem().Name())
	}

	return t.Name()
}

// TypeNameSimple takes a type and returns a more universal/generic name.
// Floats are always "float", unsigned ints are always "uint", ints are always "int".
func TypeNameSimple(t reflect.Type) string {
	isPtr := t.Kind() == reflect.Ptr
	elemType := t

	if isPtr {
		elemType = t.Elem()
	}

	out := elemType.Name()

	switch elemType.Kind() {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		out = "uint"
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		out = "int"
	case reflect.Float32, reflect.Float64:
		out = "float"
	}

	if isPtr {
		out = "*" + out
	}

	return out
}

// compare performs the actual check between the map fields and the struct fields.
func compare(t reflect.Type, src map[string]interface{}, opts *CompareOpts, results *CompareResults) {
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		fieldName, skip := parseField(f)

		if skip {
			continue
		}

		// If the field is embedded also check its fields.
		if f.Anonymous {
			compare(f.Type, src, opts, results)
			continue
		}

		if srcField, ok := src[fieldName]; ok {
			srcValue := reflect.ValueOf(srcField)

			if !opts.ConvertibleFunc(f.Type, srcValue) {
				var srcTypeName string

				if !srcValue.IsValid() {
					srcTypeName = "null"
				} else {
					srcTypeName = opts.TypeNameFunc(srcValue.Type())
				}

				mismatch := FieldMismatch{
					Field:    fieldName,
					Expected: opts.TypeNameFunc(f.Type),
					Actual:   srcTypeName,
				}

				results.MismatchedFields = append(results.MismatchedFields, mismatch)
			}
		} else {
			results.MissingFields = append(results.MissingFields, fieldName)
		}
	}
}

// isFloatType returns true if the type is a floating point. Note that this doesn't
// care about the value -- unmarshaling the number "0" gives a float, not an int.
func isFloatType(t reflect.Type) (yes bool) {
	switch t.Kind() {
	case reflect.Float32, reflect.Float64:
		yes = true
	}

	return
}

// isIntegerType returns whether the type is an integer and if it's unsigned.
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

// typeNameStartsWithVowel returns true if the type name starts with a vowel.
// This doesn't include "u" as a vowel since words like "user" should be "a user"
// and not "an user".
func typeNameStartsWithVowel(t string) bool {
	t = strings.TrimLeft(t, "*")

	switch strings.ToLower(t[:1]) {
	case "a", "e", "i", "o":
		return true
	}

	return false
}

// typeNameWithArticle returns the type name with an indefinite article ("a" or "an").
// If the type name is "null" it just returns "null".
func typeNameWithArticle(t string) string {
	if t == "null" {
		return t
	}

	if typeNameStartsWithVowel(t) {
		return "an " + t
	} else {
		return "a " + t
	}
}
