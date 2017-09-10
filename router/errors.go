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

// handleError is the default error handler
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
	// handle wrapped error
	case *errors.Error:
		s := e.StackFrames()[0]
		f := path.Base(s.File)
		m.Logger.Errorf("[%s/%s:%d] %v", s.Package, f, s.LineNumber, e.Error())
		// recurse on the err: it is either a known Error or unknown error
		m.HandleError(rw, req, e.Err)
		return

	case *Error:
		m.SimpleError(rw, e)
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

// QuietError logs an error and returns the given status without a body
func (m *Module) QuietError(rw http.ResponseWriter, status int, err error) {
	m.Logger.Errorf("%d %s", status, err)
	rw.WriteHeader(status)
}

// SimpleError responds with err.Error() as the detail of a JSON error response.
func (m *Module) SimpleError(rw http.ResponseWriter, err *Error) error {
	return m.Error(rw, int(err.err.GetCode()), &err.err)
}

func (m *Module) Error(rw http.ResponseWriter, status int, errors ...*api.Error) error {
	for _, err := range errors {
		m.Logger.Error(err)
	}
	return Proto(rw, status, &api.ErrorResponse{
		Errors: errors,
	})
}
