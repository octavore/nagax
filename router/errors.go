package router

import (
	"fmt"
	"net/http"
	"path"

	"github.com/go-errors/errors"

	"github.com/octavore/nagax/proto/nagax/router/api"
)

var (
	ErrNotFound      = fmt.Errorf("not found")
	ErrNotAuthorized = fmt.Errorf("not authorized")
	ErrForbidden     = fmt.Errorf("forbidden")
	ErrInternal      = fmt.Errorf("internal server error")
)

type Error struct {
	err      api.Error
	source   string
	silent   bool
	redirect bool
}

func (e *Error) Error() string {
	if e.source != "" {
		return e.source + ": " + e.err.String()
	}
	return e.err.String()
}

// newError creates an Error with the appropriate enum for the code.
func newError(code int32, detail, source string, silent, redirect bool) error {
	codeEnum := api.ErrorCode(code)
	_, ok := api.ErrorCode_name[code]
	if !ok {
		if code >= 400 && code < 499 {
			code = int32(api.ErrorCode_bad_request)
		} else {
			code = int32(api.ErrorCode_internal_server_error)
		}
		codeEnum = api.ErrorCode(code)
	}
	return &Error{
		silent:   silent,
		redirect: redirect,
		source:   source,
		err: api.Error{
			Code:   &code,
			Title:  codeEnum.Enum(),
			Detail: &detail,
		},
	}
}

// NewRequestError creates an Error with source set to the request url
func NewRequestError(req *http.Request, code int32, detail string) error {
	return newError(code, detail, req.URL.String(), false, false)
}

// NewQuietWrap creates a quiet *wrapped* error that does not return a body
func NewQuietWrap(req *http.Request, code int32, detail string) error {
	err := newError(code, detail, req.URL.String(), true, false)
	return errors.Wrap(err, 1)
}

// NewQuietError logs the error but does not show it to the user
func NewQuietError(req *http.Request, code int32, e error) error {
	err := newError(code, errString(e), req.URL.String(), true, false)
	return errors.Wrap(err, 1)
}

// NewRedirectingError creates an Error with source set to the request url
// and the redirect flag set to true, which will
func NewRedirectingError(req *http.Request, code int32, e error) error {
	err := newError(code, errString(e), req.URL.String(), false, true)
	return errors.Wrap(err, 1)
}

// HandleError is the default error handler
// TODO: needs to handle API vs non-API errors??
func (m *Module) HandleError(rw http.ResponseWriter, req *http.Request, err error) {
	switch err {
	case ErrNotFound:
		err = NewRequestError(req, http.StatusNotFound, "not found")
	case ErrNotAuthorized:
		err = NewRequestError(req, http.StatusUnauthorized, "not authenticated")
	case ErrForbidden:
		err = NewRequestError(req, http.StatusForbidden, "forbidden")
	case ErrInternal:
		err = NewRequestError(req, http.StatusInternalServerError, "internal server error")
	}

	switch e := err.(type) {
	// handle wrapped error created by errors.Wrap
	case *errors.Error:
		// log the stack trace and then recurse on the original err, which
		// is either a known *Error or unknown error
		m.Logger.Error(errString(err))
		m.HandleError(rw, req, e.Err)
		return

	case *Error:
		status := int(e.err.GetCode())
		if e.silent {
			m.Logger.Error(e, "(quiet)")
			rw.WriteHeader(status)
		} else if e.redirect {
			m.Logger.Error(e, "(redirect)")
			m.ErrorPage(rw, req, status)
		} else {
			m.SimpleError(rw, e)
		}
		return

	default:
		m.Logger.Errorf(`code:500: detail:"%v"`, e)
		err := newError(http.StatusInternalServerError, "internal server error", req.URL.String(), false, false).(*Error)

		Proto(rw, http.StatusInternalServerError, &api.ErrorResponse{
			Errors: []*api.Error{&err.err},
		})
		return
	}
}

// SimpleError responds with err.Error() as the "internal server error" of a JSON error response.
func (m *Module) SimpleError(rw http.ResponseWriter, err *Error) error {
	return m.Error(rw, err.err.GetCode(), &err.err)
}

func (m *Module) Error(rw http.ResponseWriter, status int32, errors ...*api.Error) error {
	for _, err := range errors {
		m.Logger.Error(err)
	}
	return Proto(rw, int(status), &api.ErrorResponse{
		Errors: errors,
	})
}

// errString prints out an error, with its location if appropriate
func errString(err error) string {
	e, ok := err.(*errors.Error)
	if !ok {
		return err.Error()
	}
	s := e.StackFrames()[0]
	f := path.Base(s.File)
	prefix := fmt.Sprintf("[%s/%s:%d]", s.Package, f, s.LineNumber)
	if e.Error() == "" {
		return prefix
	}
	return fmt.Sprint(prefix, " ", e.Error())
}
