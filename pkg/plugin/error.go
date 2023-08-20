package plugin

import (
	"errors"
	"fmt"
	"strings"
)

type ValidationError struct {
	Errors []error
}

func NewValidationError(errs ...error) *ValidationError {
	return &ValidationError{
		Errors: errs,
	}
}

func (e *ValidationError) Append(errs ...error) {
	for _, err := range errs {
		if err == nil {
			continue
		}

		var valErr *ValidationError
		if errors.As(err, &valErr) {
			e.Errors = append(e.Errors, valErr.Errors...)
		} else {
			e.Errors = append(e.Errors, err)
		}
	}
}

func (e *ValidationError) Prefix(prefix string) {
	for i, err := range e.Errors {
		e.Errors[i] = fmt.Errorf("%s: %w", prefix, err)
	}
}

func Prefix(err error, prefix string) {
	var valErr *ValidationError
	if errors.As(err, &valErr) {
		valErr.Prefix(prefix)
	}
}

func (e *ValidationError) Return() error {
	if len(e.Errors) == 0 {
		return nil
	}
	return e
}

func (e *ValidationError) Error() string {
	out := &strings.Builder{}
	_, _ = fmt.Fprintf(out, "validation failed with %d errors:\n", len(e.Errors))
	for _, err := range e.Errors {
		_, _ = fmt.Fprintf(out, " - %v\n", err)
	}
	return out.String()
}

func (e *ValidationError) Unwrap() []error {
	return e.Errors
}
