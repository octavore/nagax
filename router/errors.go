package router

import (
	"fmt"
	"net/http"
	"path"

	"github.com/go-errors/errors"

	"github.com/octavore/nagax/proto/nagax/router/api"
)

var (
	ErrNotFound  = fmt.Errorf("not found")
	ErrForbidden = fmt.Errorf("forbidden")
	ErrInternal  = fmt.Errorf("internal server error")
)

type Error struct {
	err    api.Error
	source string
	silent bool
}

func (e *Error) Error() string {
	if e.source != "" {
		return e.source + ": " + e.err.String()
	}
	return e.err.String()
}

// NewError creates an Error with the appropriate enum for the code.
func NewError(code int32, detail string) *Error {
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
	// todo: wrap here?
	return &Error{
		err: api.Error{
			Code:   &code,
			Title:  codeEnum.Enum(),
			Detail: &detail,
		},
	}
}

// NewRequestError creates an Error with source set to the request url
func NewRequestError(req *http.Request, code int32, detail string) *Error {
	err := NewError(code, detail)
	err.source = req.URL.String()
	return err
}

// NewQuietWrap creates a quiet *wrapped* error that does not return a body
func NewQuietWrap(req *http.Request, code int32, detail string) error {
	err := NewError(code, detail)
	if req != nil {
		err.source = req.URL.String()
	}
	err.silent = true
	return errors.Wrap(err, 1)
}

// NewQuietError logs
// if e is wrapped:
func NewQuietError(req *http.Request, code int32, e error) error {
	// e is a wrapped error, updated the original error to be an *Error with silent=true
	wrapped, ok := e.(*errors.Error)
	if ok {
		wrapped.Err = NewError(code, wrapped.Err.Error())
		return wrapped
	}
	// e is not a wrapped error, so create a new quiet error
	return NewQuietWrap(req, code, e.Error())
}

// HandleError is the default error handler
func (m *Module) HandleError(rw http.ResponseWriter, req *http.Request, err error) {
	switch err {
	case ErrNotFound:
		err = NewRequestError(req, http.StatusNotFound, "not found")
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
		if e.silent {
			m.Logger.Error(e, "(quiet)")
			rw.WriteHeader(int(e.err.GetCode()))
		} else {
			m.SimpleError(rw, e)
		}
		return

	default:
		m.Logger.Errorf(`code:500: detail:"%v"`, e)
		err := NewRequestError(req, http.StatusInternalServerError, "internal server error")
		Proto(rw, http.StatusInternalServerError, &api.ErrorResponse{
			Errors: []*api.Error{&err.err},
		})
		return
	}
}

// SimpleError responds with err.Error() as the detail of a JSON error response.
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
	return fmt.Sprintf("[%s/%s:%d] %v", s.Package, f, s.LineNumber, e.Error())
}
