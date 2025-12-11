package errors

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Sentinel errors
var (
	ErrNotFound     = errors.New("not found")
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
	ErrConflict     = errors.New("conflict")
	ErrValidation   = errors.New("validation")
	ErrInternal     = errors.New("internal")
)

// ErrorResponse is the API error response structure
type ErrorResponse struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id,omitempty"`
}

// Wrap creates an error with a message that wraps a sentinel
// Usage: Wrap(ErrNotFound, "user with id %d not found", id)
func Wrap(sentinel error, format string, args ...any) error {
	msg := fmt.Sprintf(format, args...)
	return &wrappedError{msg: msg, sentinel: sentinel}
}

type wrappedError struct {
	msg      string
	sentinel error
}

func (e *wrappedError) Error() string {
	return e.msg
}

func (e *wrappedError) Unwrap() error {
	return e.sentinel
}

// Code returns HTTP status code for an error
func Code(err error) int {
	switch {
	case errors.Is(err, ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, ErrUnauthorized):
		return http.StatusUnauthorized
	case errors.Is(err, ErrForbidden):
		return http.StatusForbidden
	case errors.Is(err, ErrConflict):
		return http.StatusConflict
	case errors.Is(err, ErrValidation):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

// ErrorCode returns a string code for an error
func ErrorCode(err error) string {
	switch {
	case errors.Is(err, ErrNotFound):
		return "NOT_FOUND"
	case errors.Is(err, ErrUnauthorized):
		return "UNAUTHORIZED"
	case errors.Is(err, ErrForbidden):
		return "FORBIDDEN"
	case errors.Is(err, ErrConflict):
		return "CONFLICT"
	case errors.Is(err, ErrValidation):
		return "VALIDATION_ERROR"
	default:
		return "INTERNAL_ERROR"
	}
}

// Response sends error as JSON response
func Response(c *gin.Context, err error) {
	resp := ErrorResponse{
		Code:      ErrorCode(err),
		Message:   err.Error(),
		RequestID: c.GetString("request_id"),
	}
	c.JSON(Code(err), resp)
}

// Is exposes errors.Is for convenience
func Is(err, target error) bool {
	return errors.Is(err, target)
}
