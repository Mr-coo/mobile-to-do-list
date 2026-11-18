package apperr

import (
	"encoding/json"
	"errors"
	"io"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// Registered here, where binding errors are translated, so FieldError.Field()
// is the JSON name in every importer (server and tests alike).
func init() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterTagNameFunc(func(f reflect.StructField) string {
			name, _, _ := strings.Cut(f.Tag.Get("json"), ",")
			if name == "-" {
				return ""
			}
			return name
		})
	}
}

// FromBinding turns a gin binding error into a ValidationError. Unknown
// error kinds are returned as nil so callers can treat them as internal.
func FromBinding(err error) *ValidationError {
	var verrs validator.ValidationErrors
	if errors.As(err, &verrs) {
		details := make(map[string]string, len(verrs))
		for _, fe := range verrs {
			details[fe.Field()] = ruleMessage(fe)
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
