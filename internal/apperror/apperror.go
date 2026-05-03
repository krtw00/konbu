package apperror

import (
	"fmt"
	"net/http"
)

type Error struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	HTTPStatus int    `json:"-"`
	Err        error  `json:"-"`
}

func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *Error) Unwrap() error {
	return e.Err
}

func NotFound(resource string) *Error {
	return &Error{
		Code:       "not_found",
		Message:    resource + " not found",
		HTTPStatus: http.StatusNotFound,
	}
}

func BadRequest(message string) *Error {
	return &Error{
		Code:       "bad_request",
		Message:    message,
		HTTPStatus: http.StatusBadRequest,
	}
}

func Unauthorized(message string) *Error {
	return &Error{
		Code:       "unauthorized",
		Message:    message,
		HTTPStatus: http.StatusUnauthorized,
	}
}

func Forbidden(message string) *Error {
	return &Error{
		Code:       "forbidden",
		Message:    message,
		HTTPStatus: http.StatusForbidden,
	}
}

func Internal(err error) *Error {
	return &Error{
		Code:       "internal_error",
		Message:    "internal server error",
		HTTPStatus: http.StatusInternalServerError,
		Err:        err,
	}
}

func ServiceUnavailable(message string) *Error {
	return &Error{
		Code:       "service_unavailable",
		Message:    message,
		HTTPStatus: http.StatusServiceUnavailable,
	}
}
