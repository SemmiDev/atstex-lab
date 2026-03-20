package apperrors

import (
	"errors"
	"net/http"

	"github.com/semmidev/problem"
)

// Translate maps an error to a *problem.Problem.
// It is the final firewall before any error reaches the network.
func Translate(err error) *problem.Problem {
	var safe *SafeError
	if errors.As(err, &safe) {
		return &problem.Problem{
			Type:   safe.Code,
			Title:  http.StatusText(safe.HTTPStatus),
			Status: safe.HTTPStatus,
			Detail: safe.UserMsg,
		}
	}

	// Fall back to sentinel matching
	var (
		status int
		detail string
		code   string
	)

	switch {
	case errors.Is(err, ErrNotFound):
		status = http.StatusNotFound
		detail = "Resource not found."
		code = "RESOURCE_NOT_FOUND"
	case errors.Is(err, ErrConflict):
		status = http.StatusConflict
		detail = "Resource already exists."
		code = "RESOURCE_CONFLICT"
	case errors.Is(err, ErrInvalidInput):
		status = http.StatusBadRequest
		detail = "The request contains invalid data."
		code = "INVALID_INPUT"
	case errors.Is(err, ErrUnauthorized):
		status = http.StatusUnauthorized
		detail = "Invalid credentials."
		code = "UNAUTHORIZED"
	case errors.Is(err, ErrForbidden):
		status = http.StatusForbidden
		detail = "Permission denied."
		code = "FORBIDDEN"
	default:
		status = http.StatusInternalServerError
		detail = "An internal error occurred. Please contact support."
		code = "INTERNAL_ERROR"
	}

	return &problem.Problem{
		Type:   code,
		Title:  http.StatusText(status),
		Status: status,
		Detail: detail,
	}
}
