package utils

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ErrorType represents different types of errors
type ErrorType string

const (
	ErrorTypeValidation ErrorType = "validation"
	ErrorTypeNotFound   ErrorType = "not_found"
	ErrorTypeForbidden  ErrorType = "forbidden"
	ErrorTypeInternal   ErrorType = "internal"
	ErrorTypeExternal   ErrorType = "external"
)

// AppError wraps application errors with context
type AppError struct {
	Type     ErrorType `json:"type"`
	Message  string    `json:"message,omitempty"` // User-facing message (safe to expose)
	Details  string    `json:"-"`                 // Internal details (NEVER expose to client)
	Code     int       `json:"code,omitempty"`    // HTTP status code
	Internal error     `json:"-"`                 // The underlying error
}

func (e *AppError) Error() string {
	if e.Internal != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Type, e.Message, e.Internal)
	}
	return fmt.Sprintf("[%s] %s", e.Type, e.Message)
}

// NewValidationError creates a validation error
func NewValidationError(message string, details ...interface{}) *AppError {
	detailStr := ""
	if len(details) > 0 {
		detailStr = fmt.Sprintf("%v", details[0])
	}
	return &AppError{
		Type:     ErrorTypeValidation,
		Message:  message,
		Details:  detailStr,
		Code:     http.StatusBadRequest,
		Internal: nil,
	}
}

// NewNotFoundError creates a not found error
func NewNotFoundError(resource string) *AppError {
	return &AppError{
		Type:    ErrorTypeNotFound,
		Message: fmt.Sprintf("%s not found", resource),
		Details: fmt.Sprintf("Resource '%s' does not exist", resource),
		Code:    http.StatusNotFound,
	}
}

// NewForbiddenError creates a forbidden error
func NewForbiddenError(message string) *AppError {
	return &AppError{
		Type:    ErrorTypeForbidden,
		Message: message,
		Code:    http.StatusForbidden,
	}
}

// NewInternalError creates an internal error
func NewInternalError(operation string, err error) *AppError {
	return &AppError{
		Type:     ErrorTypeInternal,
		Message:  fmt.Sprintf("Failed to %s", operation),
		Details:  fmt.Sprintf("Operation '%s' failed with error: %v", operation, err),
		Code:     http.StatusInternalServerError,
		Internal: err,
	}
}

// NewExternalError creates an external service error
func NewExternalError(service string, err error) *AppError {
	return &AppError{
		Type:     ErrorTypeExternal,
		Message:  fmt.Sprintf("%s service temporarily unavailable", service),
		Details:  fmt.Sprintf("External service '%s' error: %v", service, err),
		Code:     http.StatusServiceUnavailable,
		Internal: err,
	}
}

// IsNotFound checks if the error is a not found error
func IsNotFound(err error) bool {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Type == ErrorTypeNotFound
	}
	return false
}

// HandleError handles errors in controllers and sends appropriate HTTP responses
// This function ensures internal error details are NEVER exposed to clients
func HandleError(c *gin.Context, err error, operation string) bool {
	if err == nil {
		return false
	}

	var appErr *AppError

	// Check if it's already an AppError
	if ae, ok := err.(*AppError); ok {
		appErr = ae
	} else {
		// Wrap unknown errors as internal errors
		appErr = NewInternalError(operation, err)
	}

	// Log the full error details for debugging (includes internal info)
	if appErr.Internal != nil {
		log.Printf("[ERROR] %s failed: %v (details: %s)", operation, appErr.Internal, appErr.Details)
	} else {
		log.Printf("[ERROR] %s: %s", operation, appErr.Message)
	}

	// Create response without exposing internal details
	response := gin.H{
		"error": appErr.Message,
		"type":  string(appErr.Type),
	}

	// Add error code if it's validation error
	if appErr.Type == ErrorTypeValidation {
		response["code"] = "VALIDATION_ERROR"
	}

	c.JSON(appErr.Code, response)
	return true
}
