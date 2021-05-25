package schema

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// MismatchError is represented as a map of field mismatch errors.
type MismatchError map[string]interface{}

func (err MismatchError) Error() string {
	b := strings.Builder{}

	for key, val := range err {
		if s, ok := val.(string); ok {
			b.WriteString(fmt.Sprintf("%s: %s", key, s))
			b.WriteString(", ")
		}
	}

	s := b.String()

	// Remove the trailing ", "
	if s != "" {
		s = s[:len(s)-2]
	}

	return s
}

func (err MismatchError) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}(err))
}

var (
	ErrInvalidDst = errors.New("dst must be a pointer to a struct")
	ErrNilSrc     = errors.New("src must not be nil")
)
