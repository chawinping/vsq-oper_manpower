package errors

import (
	"fmt"
	"net/http"
)

// ErrorCode represents a standardized error code
type ErrorCode string

const (
	// Validation errors
	ErrorCodeValidation   ErrorCode = "VALIDATION_ERROR"
	ErrorCodeInvalidInput ErrorCode = "INVALID_INPUT"
	ErrorCodeMissingField ErrorCode = "MISSING_FIELD"

	// Business logic errors
	ErrorCodeBusinessLogic ErrorCode = "BUSINESS_LOGIC_ERROR"
	ErrorCodeConstraint    ErrorCode = "CONSTRAINT_VIOLATION"
	ErrorCodeDuplicate     ErrorCode = "DUPLICATE_ENTRY"

	// Resource errors
	ErrorCodeNotFound      ErrorCode = "NOT_FOUND"
	ErrorCodeAlreadyExists ErrorCode = "ALREADY_EXISTS"

	// Permission errors
	ErrorCodeUnauthorized ErrorCode = "UNAUTHORIZED"
	ErrorCodeForbidden    ErrorCode = "FORBIDDEN"

	// System errors
	ErrorCodeInternal ErrorCode = "INTERNAL_ERROR"
	ErrorCodeDatabase ErrorCode = "DATABASE_ERROR"
	ErrorCodeTimeout  ErrorCode = "TIMEOUT"
)

// AppError represents an application error with structured information
type AppError struct {
	Code       ErrorCode         `json:"code"`
	Message    string            `json:"error"`
	Details    map[string]string `json:"details,omitempty"`
	StatusCode int               `json:"-"`
	Err        error             `json:"-"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// NewAppError creates a new application error
func NewAppError(code ErrorCode, message string, statusCode int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
		Details:    make(map[string]string),
	}
}

// WithError wraps an underlying error
func (e *AppError) WithError(err error) *AppError {
	e.Err = err
	return e
}

// WithDetail adds a detail field to the error
func (e *AppError) WithDetail(key, value string) *AppError {
	if e.Details == nil {
		e.Details = make(map[string]string)
	}
	e.Details[key] = value
	return e
}

// Predefined error constructors

func NewValidationError(message string) *AppError {
	return NewAppError(ErrorCodeValidation, message, http.StatusBadRequest)
}

func NewValidationErrorWithDetails(message string, details map[string]string) *AppError {
	err := NewValidationError(message)
	for k, v := range details {
		err.Details[k] = v
	}
	return err
}

func NewNotFoundError(resource string) *AppError {
	return NewAppError(ErrorCodeNotFound, fmt.Sprintf("%s not found", resource), http.StatusNotFound)
}

func NewBusinessLogicError(message string) *AppError {
	return NewAppError(ErrorCodeBusinessLogic, message, http.StatusBadRequest)
}

func NewConstraintError(message string) *AppError {
	return NewAppError(ErrorCodeConstraint, message, http.StatusConflict)
}

func NewDuplicateError(message string) *AppError {
	return NewAppError(ErrorCodeDuplicate, message, http.StatusConflict)
}

func NewForbiddenError(message string) *AppError {
	return NewAppError(ErrorCodeForbidden, message, http.StatusForbidden)
}

func NewInternalError(message string) *AppError {
	return NewAppError(ErrorCodeInternal, message, http.StatusInternalServerError)
}

func NewDatabaseError(err error) *AppError {
	return NewAppError(ErrorCodeDatabase, "Database operation failed", http.StatusInternalServerError).WithError(err)
}

// IsAppError checks if an error is an AppError
func IsAppError(err error) bool {
	_, ok := err.(*AppError)
	return ok
}

// AsAppError converts an error to AppError if possible
func AsAppError(err error) (*AppError, bool) {
	appErr, ok := err.(*AppError)
	return appErr, ok
}

// SanitizeError returns a user-friendly error message, hiding internal details in production
func SanitizeError(err error, isDevelopment bool) string {
	if appErr, ok := AsAppError(err); ok {
		return appErr.Message
	}

	// In production, don't expose internal error details
	if !isDevelopment {
		return "An unexpected error occurred. Please try again or contact support."
	}

	// In development, show the actual error
	return err.Error()
}
