package httperror

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/octavore/nagax/proto/router/api"
)

type HTTPError struct {
	Detail    string
	Code      int
	BaseError error
}

func (e *HTTPError) GetCode() int {
	return e.Code
}

func (e *HTTPError) GetDetail() string {
	return e.Detail
}

func (e *HTTPError) Error() string {
	msg := fmt.Sprintf("code=%d detail=%q", e.Code, e.Detail)
	if e.BaseError != nil {
		msg += fmt.Sprintf(" error=%q", UnwrapAll(e.BaseError))
	}
	return msg
}

func (e *HTTPError) Unwrap() error {
	return e.BaseError
}

func (e *HTTPError) Is(target error) bool {
	if t, ok := target.(*HTTPError); ok {
		return e.Code == t.Code && e.Detail == t.Detail && errors.Is(e.BaseError, t.BaseError)
	}
	return false
}

func (e *HTTPError) WithDetail(format string, args ...any) *HTTPError {
	e.Detail = fmt.Sprintf(format, args...)
	return e
}

func (e *HTTPError) WithError(err error) *HTTPError {
	e.BaseError = err
	return e
}

func (e *HTTPError) ToProto() *api.Error {
	code := int32(e.Code)
	err := &api.Error{
		Title: api.ErrorCode_internal_server_error.Enum(),
		Code:  &code,
	}
	if e.Detail != "" {
		err.Detail = &e.Detail
	}
	if _, ok := api.ErrorCode_name[code]; ok {
		// supported error codes
		err.Title = api.ErrorCode(e.Code).Enum()
	} else if e.Code >= 400 && e.Code < 499 {
		err.Title = api.ErrorCode_bad_request.Enum()
	}
	return err
}

// BadRequest is a helper to return a 400 error
func BadRequest(format string, args ...any) *HTTPError {
	return (&HTTPError{Code: http.StatusBadRequest}).WithDetail(format, args...)
}

// NotAuthorized is a helper to return a 401 error
func NotAuthorized(format string, args ...any) *HTTPError {
	return (&HTTPError{Code: http.StatusUnauthorized}).WithDetail(format, args...)
}

// Forbidden is a helper to return a 403 error
func Forbidden(format string, args ...any) *HTTPError {
	return (&HTTPError{Code: http.StatusForbidden}).WithDetail(format, args...)
}

// NotFound is a helper to return a 404 error
func NotFound(format string, args ...any) *HTTPError {
	return (&HTTPError{Code: http.StatusNotFound}).WithDetail(format, args...)
}

// Internal is a helper to return a 500 error
func InternalError() *HTTPError {
	return &HTTPError{Code: http.StatusInternalServerError}
}
