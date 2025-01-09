package httperror

import (
	"errors"
	"net/http"
)

type GetCoder interface {
	GetCode() int
}

func CodeFromErr(err error) (int, bool) {
	// if err implements GetCoder, return the code
	if e, ok := err.(GetCoder); ok {
		return e.GetCode(), true
	}
	// recurse if we can unwrap
	if u := errors.Unwrap(err); u != nil {
		return CodeFromErr(u)
	}
	return http.StatusInternalServerError, false
}
