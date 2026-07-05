// Package apperrors mirrors src/errors of the Node backend.
// errors.go ≈ app-error.ts: one AppError type + constructors per case,
// producing the same codes and HTTP status codes.
package apperrors

import "fmt"

// AppError ≈ class AppError. It implements the `error` interface and travels
// through handlers via c.Error(err) — the Go twin of Express's next(error).
type AppError struct {
	Code       string
	Message    string
	StatusCode int
	Details    any   // validation details or extra context
	Err        error // wrapped cause (shown as "stack" in development)
}

func (e *AppError) Error() string { return e.Message }

// Unwrap lets errors.Is / errors.As inspect the wrapped cause.
func (e *AppError) Unwrap() error { return e.Err }

func NewNotFound(resource string, identifier any) *AppError {
	return &AppError{
		Code:       "NOT_FOUND",
		Message:    fmt.Sprintf("%s with identifier %v not found", resource, identifier),
		StatusCode: 404,
	}
}

func NewConflict(message string) *AppError {
	return &AppError{Code: "CONFLICT", Message: message, StatusCode: 409}
}

func NewValidation(message string, details any) *AppError {
	return &AppError{Code: "VALIDATION_ERROR", Message: message, StatusCode: 400, Details: details}
}

func NewBadRequest(message string) *AppError {
	return &AppError{Code: "BAD_REQUEST", Message: message, StatusCode: 400}
}

func NewBusinessRule(message string) *AppError {
	return &AppError{Code: "BUSINESS_RULE_VIOLATION", Message: message, StatusCode: 422}
}

func NewInternal(message string) *AppError {
	if message == "" {
		message = "An internal server error occurred"
	}
	return &AppError{Code: "INTERNAL_ERROR", Message: message, StatusCode: 500}
}

func NewUnauthorized(message string) *AppError {
	return &AppError{Code: "UNAUTHORIZED", Message: message, StatusCode: 401}
}

func NewForbidden(message string) *AppError {
	return &AppError{Code: "FORBIDDEN", Message: message, StatusCode: 403}
}
