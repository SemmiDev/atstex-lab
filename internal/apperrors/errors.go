package apperrors

import (
	"errors"
	"fmt"
	"net/http"
)

var (
	ErrNotFound     = errors.New("resource not found")
	ErrConflict     = errors.New("resource already exists")
	ErrInvalidInput = errors.New("invalid input")
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
	ErrInternal     = errors.New("internal error")
)

type SafeError struct {
	Code       string
	HTTPStatus int
	UserMsg    string
	Internal   error
	Metadata   map[string]string
	Sentinel   error
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

func NewInternal(internal error) *SafeError {
	return &SafeError{
		Code:       "INTERNAL_ERROR",
		HTTPStatus: http.StatusInternalServerError,
		UserMsg:    "An internal error occurred. Please contact support.",
		Internal:   internal,
		Sentinel:   ErrInternal,
	}
}
