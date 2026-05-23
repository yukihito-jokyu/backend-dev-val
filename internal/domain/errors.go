package domain

import "fmt"

type NotFoundError struct {
	ID int64
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("todo not found: id=%d", e.ID)
}

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error: %s - %s", e.Field, e.Message)
}
