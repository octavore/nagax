package httperror

import (
	"fmt"
	"net/http"
)

type HTTPError struct {
	Detail string
	Code   int32
}

func (e *HTTPError) Error() string {
	return e.Detail
}

// BadRequest is a helper to return a 400 error
func BadRequest(format string, args ...interface{}) *HTTPError {
	return &HTTPError{
		Detail: fmt.Sprintf(format, args...),
		Code:   http.StatusBadRequest,
	}
}

// NotAuthorized is a helper to return a 401 error
func NotAuthorized(format string, args ...interface{}) *HTTPError {
	return &HTTPError{
		Detail: fmt.Sprintf(format, args...),
		Code:   http.StatusUnauthorized,
	}
}

// Forbidden is a helper to return a 403 error
func Forbidden(format string, args ...interface{}) *HTTPError {
	return &HTTPError{
		Detail: fmt.Sprintf(format, args...),
		Code:   http.StatusForbidden,
	}
}

// NotFound is a helper to return a 404 error
func NotFound(format string, args ...interface{}) *HTTPError {
	return &HTTPError{
		Detail: fmt.Sprintf(format, args...),
		Code:   http.StatusNotFound,
	}
}

// Internal is a helper to return a 404 error
func InternalError(format string, args ...interface{}) *HTTPError {
	return &HTTPError{
		Detail: fmt.Sprintf(format, args...),
		Code:   http.StatusInternalServerError,
	}
}
