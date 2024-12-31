package errors

import (
	"errors"
	"net/http"
)

// AppError represents an application error
type AppError struct {
	Err        error
	Message    string
	StatusCode int
}

func (e *AppError) Error() string {
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// New creates a new AppError
func New(message string, statusCode int) *AppError {
	return &AppError{
		Message:    message,
		StatusCode: statusCode,
	}
}

// Wrap wraps an existing error with additional context
func Wrap(err error, message string, statusCode int) *AppError {
	return &AppError{
		Err:        err,
		Message:    message,
		StatusCode: statusCode,
	}
}

// Common errors
var (
	ErrNotFound          = New("Resource not found", http.StatusNotFound)
	ErrUnauthorized      = New("Unauthorized", http.StatusUnauthorized)
	ErrInvalidInput      = New("Invalid input", http.StatusBadRequest)
	ErrInternalServer    = New("Internal server error", http.StatusInternalServerError)
	ErrDuplicateResource = New("Resource already exists", http.StatusConflict)
)

// IsAppError checks if an error is an AppError
func IsAppError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr)
}

// StatusCode returns the HTTP status code for an error
func StatusCode(err error) int {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.StatusCode
	}
	return http.StatusInternalServerError
}

// Message returns a user-friendly error message
func Message(err error) string {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Message
	}
	return "An unexpected error occurred"
}
