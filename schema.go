package schema

import (
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
	MissingFields []FieldMissing
}

// Errors returns a MismatchError containing the type errors. If there were no
// type errors, returns nil.
func (cr *CompareResults) Errors() error {
	if len(cr.MismatchedFields) == 0 {
		return nil
	}

	m := make(map[string]interface{})

	for _, f := range cr.MismatchedFields {
		m[f.Field] = f.Message()
	}

	return MismatchError(m)
}

type FieldMissing struct {
	// Field is the JSON name of the field.
	Field string

	// Path is the full path to the field.
	Path []string
}

// String returns the field name with its path.
// e.g: "Cat.Foo"
func (f FieldMissing) String() string {
	return FieldNameWithPath(f.Field, f.Path)
}

// FieldMismatch represents a type mismatch between a struct field and a map field.
type FieldMismatch struct {
	// Field is the JSON name of the field.
	Field string

	// Expected is the expected type (the type of the dst field).
	Expected string

	// Actual is the actual type (type of the src field).
	Actual string

	// Path is the full path to the field.
	Path []string
}

// Message returns the field mismatch error as a string.
// e.g: "expected an int but it's a string"
func (f FieldMismatch) Message() string {
	return fmt.Sprintf(
		`expected %s but it's %s`,
		TypeNameWithArticle(f.Expected),
		TypeNameWithArticle(f.Actual),
	)
}

// Message returns the field mismatch error as a string, and includes the field name
// with its path in the message.
// e.g: "expected Cat.Foo to be an int but it's a string"
func (f FieldMismatch) MessageWithField() string {
	return fmt.Sprintf(
		`expected "%s" to be %s but it's %s`,
		FieldNameWithPath(f.Field, f.Path),
		TypeNameWithArticle(f.Expected),
		TypeNameWithArticle(f.Actual),
	)
}

// String returns a user friendly message explaining the type mismatch.
func (f FieldMismatch) String() string {
	return f.MessageWithField()
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
			TypeNameFunc:    DetailedTypeName,
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
			opts.TypeNameFunc = DetailedTypeName
		}
	}

	v := reflect.ValueOf(dst)

	if !v.IsValid() || v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return nil, ErrInvalidDst
	} else if src == nil {
		return nil, ErrNilSrc
	}

	results := &CompareResults{
		MismatchedFields: []FieldMismatch{},
		MissingFields:    []FieldMissing{},
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
	isStruct := t.Kind() == reflect.Struct
	dstType := t

	// Check if v is a nil value.
	if !v.IsValid() || (v.CanAddr() && v.IsNil()) {
		return isPtr
	}

	// If the dst is a pointer, check if we can convert to the type it's pointing to.
	if isPtr {
		dstType = t.Elem()
		isStruct = t.Elem().Kind() == reflect.Struct
	}

	// If the dst is a struct, we should check its nested fields.
	if isStruct {
		return v.Kind() == reflect.Map
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

		// If the field is a nested struct also check its fields.
		shouldCheckNested := isStructType(f.Type)

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
				// There is no point to check nested fields if their parent is mismatched.
				shouldCheckNested = false
			}
		} else {
			missing := FieldMissing{Field: fieldName}

			results.MissingFields = append(results.MissingFields, missing)
			// There is no point to check nested fields if their parent is missing.
			shouldCheckNested = false
		}

		if shouldCheckNested {
			nested := src[fieldName].(map[string]interface{})
			nestedType := f.Type
			if f.Type.Kind() == reflect.Ptr {
				nestedType = nestedType.Elem()
			}

			checkNestedFields(nestedType, fieldName, nested, opts, results)
		}
	}
}

func checkNestedFields(t reflect.Type, fieldName string, src map[string]interface{}, opts *CompareOpts, results *CompareResults) {
	// Remember count of fields to check if new errors occured.
	mismatchCount := len(results.MismatchedFields)
	missingCount := len(results.MissingFields)

	compare(t, src, opts, results)

	// If there were new mismatched fields, add the current field name to their path.
	if mismatchCount != len(results.MismatchedFields) {
		for mi := mismatchCount; mi < len(results.MismatchedFields); mi++ {
			results.MismatchedFields[mi].Path = append([]string{fieldName}, results.MismatchedFields[mi].Path...)
		}
	}

	// If there were new missing fields, add the current field name to their path.
	if missingCount != len(results.MissingFields) {
		for mi := missingCount; mi < len(results.MissingFields); mi++ {
			results.MissingFields[mi].Path = append([]string{fieldName}, results.MissingFields[mi].Path...)
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

// isStructType returns whether the type is a struct or a pointer to it.
func isStructType(t reflect.Type) (yes bool) {
	switch t.Kind() {
	case reflect.Struct:
		yes = true
	case reflect.Ptr:
		yes = t.Elem().Kind() == reflect.Struct
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
