package errors

import (
	"fmt"
	"net/http"
)

// Error codes
const (
	// General codes
	CodeSuccess         = 0
	CodeSystemError     = 1000
	CodeValidationError = 1001
	CodeBusinessError   = 1002
	CodeUnauthorized    = 1003
	CodeForbidden       = 1004
	CodeNotFound        = 1005
	CodeInternalError   = 1999

	// Device Type codes (2xxx)
	CodeDeviceTypeAlreadyExists = 2001
	CodeCannotRemoveCapability  = 2002
	CodeDeviceTypeInUse         = 2003

	// Device codes (21xx)
	CodeInvalidStatusTransition = 2101

	// Auth codes (30xx)
	CodeInvalidCredentials   = 3001
	CodeTokenExpired         = 3002
	CodeTokenInvalid         = 3003
	CodeMustChangePassword   = 3004
	CodeAccountDisabled      = 3005
	CodeAccountDeleted       = 3006

	// User codes (31xx)
	CodeUsernameExists    = 3101
	CodeUserNotFound      = 3102
	CodeSelfActionDenied = 3103
	CodeInvalidRole       = 3104
	CodeInvalidPassword   = 3105
)

// AppError represents an application error
type AppError struct {
	Code       int    `json:"code"`
	Message    string `json:"message"`
	Details    string `json:"details,omitempty"`
	HTTPStatus int    `json:"-"`
}

// Error implements error interface
func (e *AppError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("[%d] %s: %s", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// NewSystemError creates a system error
func NewSystemError(message string) *AppError {
	return &AppError{
		Code:       CodeSystemError,
		Message:    message,
		HTTPStatus: http.StatusInternalServerError,
	}
}

// NewValidationError creates a validation error
func NewValidationError(message string) *AppError {
	return &AppError{
		Code:       CodeValidationError,
		Message:    message,
		HTTPStatus: http.StatusBadRequest,
	}
}

// NewBusinessError creates a business error
func NewBusinessError(code int, message string) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: http.StatusBadRequest,
	}
}

// NewConflictError creates a conflict error (HTTP 409)
func NewConflictError(code int, message string) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: http.StatusConflict,
	}
}

// NewNotFoundError creates a not found error
func NewNotFoundError(message string) *AppError {
	return &AppError{
		Code:       CodeNotFound,
		Message:    message,
		HTTPStatus: http.StatusNotFound,
	}
}

// NewUnauthorizedError creates an unauthorized error
func NewUnauthorizedError(message string) *AppError {
	return &AppError{
		Code:       CodeUnauthorized,
		Message:    message,
		HTTPStatus: http.StatusUnauthorized,
	}
}

// NewForbiddenError creates a forbidden error
func NewForbiddenError(message string) *AppError {
	return &AppError{
		Code:       CodeForbidden,
		Message:    message,
		HTTPStatus: http.StatusForbidden,
	}
}

// NewInternalError creates an internal error
func NewInternalError(message string) *AppError {
	return &AppError{
		Code:       CodeInternalError,
		Message:    message,
		HTTPStatus: http.StatusInternalServerError,
	}
}

// WithDetails adds details to the error
func (e *AppError) WithDetails(details string) *AppError {
	return &AppError{
		Code:       e.Code,
		Message:    e.Message,
		Details:    details,
		HTTPStatus: e.HTTPStatus,
	}
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
