package apperr

import (
	"encoding/json"
	"errors"
	"io"
	"strings"
	"unicode"

	"github.com/go-playground/validator/v10"
)

// FromBinding turns a gin binding error into a ValidationError. Unknown
// error kinds are returned as nil so callers can treat them as internal.
func FromBinding(err error) *ValidationError {
	var verrs validator.ValidationErrors
	if errors.As(err, &verrs) {
		details := make(map[string]string, len(verrs))
		for _, fe := range verrs {
			details[snakeCase(fe.Field())] = ruleMessage(fe)
		}
		return &ValidationError{Details: details}
	}

	var syntaxErr *json.SyntaxError
	var typeErr *json.UnmarshalTypeError
	if errors.As(err, &syntaxErr) || errors.As(err, &typeErr) ||
		errors.Is(err, io.ErrUnexpectedEOF) || errors.Is(err, io.EOF) {
		return &ValidationError{Details: map[string]string{"body": "invalid JSON"}}
	}
	return nil
}

func ruleMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "is required"
	case "max":
		return "is too long"
	case "oneof":
		return "must be one of " + strings.ReplaceAll(fe.Param(), " ", ", ")
	default:
		return "is invalid"
	}
}

// snakeCase maps a Go field name (CreatedAt) to its JSON name (created_at).
// Every request struct field uses the snake_case of its Go name as json tag.
func snakeCase(s string) string {
	var b strings.Builder
	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 {
				b.WriteByte('_')
			}
			r = unicode.ToLower(r)
		}
		b.WriteRune(r)
	}
	return b.String()
}
