package apperr

import "errors"

var (
	ErrNotFound       = errors.New("not found")
	ErrDuplicateTitle = errors.New("duplicate title")
)

type ValidationError struct {
	Details map[string]string
}

func (e *ValidationError) Error() string { return "validation failed" }
