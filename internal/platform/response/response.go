package response

import (
	"net/http"

	"broadcast-platform/internal/platform/errors"

	"github.com/gin-gonic/gin"
)

// Response represents a unified HTTP response
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Success sends a success response
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    errors.CodeSuccess,
		Message: "success",
		Data:    data,
	})
}

// SuccessWithMessage sends a success response with custom message
func SuccessWithMessage(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    errors.CodeSuccess,
		Message: message,
		Data:    data,
	})
}

// Error sends an error response based on AppError
func Error(c *gin.Context, err error) {
	if appErr, ok := errors.AsAppError(err); ok {
		c.JSON(appErr.HTTPStatus, Response{
			Code:    appErr.Code,
			Message: appErr.Message,
		})
		return
	}

	// Unknown error - return internal error
	c.JSON(http.StatusInternalServerError, Response{
		Code:    errors.CodeInternalError,
		Message: "internal server error",
	})
}

// ErrorWithDetails sends an error response with additional details
func ErrorWithDetails(c *gin.Context, err error, details string) {
	if appErr, ok := errors.AsAppError(err); ok {
		c.JSON(appErr.HTTPStatus, Response{
			Code:    appErr.Code,
			Message: appErr.Message,
		})
		return
	}

	c.JSON(http.StatusInternalServerError, Response{
		Code:    errors.CodeInternalError,
		Message: "internal server error",
	})
}

// ValidationError sends a validation error response
func ValidationError(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, Response{
		Code:    errors.CodeValidationError,
		Message: message,
	})
}

// BusinessError sends a business error response
func BusinessError(c *gin.Context, code int, message string) {
	c.JSON(http.StatusBadRequest, Response{
		Code:    code,
		Message: message,
	})
}

// NotFound sends a not found error response
func NotFound(c *gin.Context, message string) {
	c.JSON(http.StatusNotFound, Response{
		Code:    errors.CodeNotFound,
		Message: message,
	})
}

// Unauthorized sends an unauthorized error response
func Unauthorized(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, Response{
		Code:    errors.CodeUnauthorized,
		Message: message,
	})
}

// Forbidden sends a forbidden error response
func Forbidden(c *gin.Context, message string) {
	c.JSON(http.StatusForbidden, Response{
		Code:    errors.CodeForbidden,
		Message: message,
	})
}
