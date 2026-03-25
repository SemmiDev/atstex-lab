package apperrors

import (
	"errors"
	"fmt"
	"net/http"
)

var (
	ErrNotFound          = errors.New("resource not found")
	ErrConflict          = errors.New("resource already exists")
	ErrInvalidInput      = errors.New("invalid input")
	ErrValidation        = errors.New("validation failed")
	ErrUnauthorized      = errors.New("unauthorized")
	ErrForbidden         = errors.New("forbidden")
	ErrSubscriptionLimit = errors.New("subscription limit reached")
	ErrInternal          = errors.New("internal error")
)

type SafeError struct {
	Code       string
	HTTPStatus int
	UserMsg    string
	Internal   error
	Metadata   map[string]string
	Sentinel   error
	// Fields holds per-field validation errors (populated by NewValidationError).
	Fields map[string]string
}

func (e *SafeError) Error() string {
	return e.UserMsg
}

func (e *SafeError) Unwrap() error {
	return e.Sentinel
}

func (e *SafeError) LogString() string {
	return fmt.Sprintf(
		"code=%s http=%d msg=%q cause=%v meta=%v",
		e.Code, e.HTTPStatus, e.UserMsg, e.Internal, e.Metadata,
	)
}

func NewNotFound(resource string, internal error) *SafeError {
	return &SafeError{
		Code:       "RESOURCE_NOT_FOUND",
		HTTPStatus: http.StatusNotFound,
		UserMsg:    resource + " not found.",
		Internal:   internal,
		Sentinel:   ErrNotFound,
	}
}

func NewConflict(resource string, internal error, meta map[string]string) *SafeError {
	return &SafeError{
		Code:       "RESOURCE_CONFLICT",
		HTTPStatus: http.StatusConflict,
		UserMsg:    resource + " already exists.",
		Internal:   internal,
		Sentinel:   ErrConflict,
		Metadata:   meta,
	}
}

func NewInvalidInput(msg string) *SafeError {
	return &SafeError{
		Code:       "INVALID_INPUT",
		HTTPStatus: http.StatusBadRequest,
		UserMsg:    msg,
		Sentinel:   ErrInvalidInput,
	}
}

// NewValidationError returns a 422 Unprocessable Entity error that carries
// per-field validation messages produced by validate.Struct().
func NewValidationError(fields map[string]string) *SafeError {
	return &SafeError{
		Code:       "VALIDATION_ERROR",
		HTTPStatus: http.StatusUnprocessableEntity,
		UserMsg:    "Data tidak valid",
		Sentinel:   ErrValidation,
		Fields:     fields,
	}
}

func NewUnauthorized(internal error) *SafeError {
	return &SafeError{
		Code:       "UNAUTHORIZED",
		HTTPStatus: http.StatusUnauthorized,
		UserMsg:    "Invalid credentials.",
		Internal:   internal,
		Sentinel:   ErrUnauthorized,
	}
}

func NewForbidden() *SafeError {
	return &SafeError{
		Code:       "FORBIDDEN",
		HTTPStatus: http.StatusForbidden,
		UserMsg:    "You do not have permission to perform this action.",
		Sentinel:   ErrForbidden,
	}
}

func NewSubscriptionLimit(msg string) *SafeError {
	return &SafeError{
		Code:       "SUBSCRIPTION_LIMIT_REACHED",
		HTTPStatus: http.StatusPaymentRequired,
		UserMsg:    msg,
		Sentinel:   ErrSubscriptionLimit,
	}
}

func NewInternal(internal error) *SafeError {
	return &SafeError{
		Code:       "INTERNAL_ERROR",
		HTTPStatus: http.StatusInternalServerError,
		UserMsg:    "An internal error occurred. Please contact support.",
		Internal:   internal,
		Sentinel:   ErrInternal,
	}
}
