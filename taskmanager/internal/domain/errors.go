package domain

import "fmt"

// InvalidInputError says which field is wrong and why.
type InvalidInputError struct {
	Field  string
	Reason string
}

func (e *InvalidInputError) Error() string {
	return fmt.Sprintf("invalid %s: %s", e.Field, e.Reason)
}

// Is makes errors.Is(err, ErrInvalidInput) true for this type too.
func (e *InvalidInputError) Is(target error) bool { return target == ErrInvalidInput }