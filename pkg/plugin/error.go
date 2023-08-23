package plugin

import (
	"errors"
	"fmt"
	"strings"
)

// ValidationError is a collection of errors that occurred during validation.
type ValidationError struct {
	Errors []error
}

// NewValidationError creates a new validation error with the given errors.
func NewValidationError(errs ...error) *ValidationError {
	return &ValidationError{
		Errors: errs,
	}
}

// Append adds the given errors to the validation error.
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

// Prefix adds the given prefix to all errors in the validation error.
func (e *ValidationError) Prefix(prefix string) {
	for i, err := range e.Errors {
		e.Errors[i] = fmt.Errorf("%s: %w", prefix, err)
	}
}

// Return returns an error if any errors have been collected or nil.
func (e *ValidationError) Return() error {
	if len(e.Errors) == 0 {
		return nil
	}
	return e
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	out := &strings.Builder{}
	_, _ = fmt.Fprintf(out, "validation failed with %d errors:\n", len(e.Errors))
	for _, err := range e.Errors {
		_, _ = fmt.Fprintf(out, " - %v\n", err)
	}
	return out.String()
}

// Unwrap implements the errors wrapper interface.
func (e *ValidationError) Unwrap() []error {
	return e.Errors
}
