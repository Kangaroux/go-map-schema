package schema

import (
	"fmt"
	"reflect"
	"strings"
)

// DetailedTypeName takes a type and returns it verbatim.
func DetailedTypeName(t reflect.Type) string {
	if t.Kind() == reflect.Ptr {
		return fmt.Sprintf("*%s", t.Elem().Name())
	}

	return t.Name()
}

// SimpleTypeName takes a type and returns a more universal/generic name.
// Floats are always "float", unsigned ints are always "uint", ints are always "int".
func SimpleTypeName(t reflect.Type) string {
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

// TypeNameStartsWithVowel returns true if the type name starts with a vowel.
// This doesn't include "u" as a vowel since words like "user" should be "a user"
// and not "an user".
func TypeNameStartsWithVowel(t string) bool {
	if t == "" {
		return false
	}

	t = strings.TrimLeft(t, "*")

	switch strings.ToLower(t[:1]) {
	case "a", "e", "i", "o":
		return true
	}

	return false
}

// TypeNameWithArticle returns the type name with an indefinite article ("a" or "an").
// If the type name is "null" it just returns "null".
func TypeNameWithArticle(t string) string {
	if t == "null" {
		return t
	}

	if TypeNameStartsWithVowel(t) {
		return "an " + t
	} else {
		return "a " + t
	}
}
