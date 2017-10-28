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

// Error wraps an api.Error return object
type Error struct {
	err      api.Error
	source   string
	silent   bool
	redirect bool
	request  *http.Request
}

func (e *Error) Error() string {
	errStr := e.err.String()
	if e.source != "" {
		errStr = e.source + ": " + errStr
	}
	if e.silent {
		errStr = errStr + " (silent)"
	}
	if e.redirect {
		errStr = errStr + " (redirect)"
	}
	return errStr
}

func (e *Error) GetRequest() *http.Request {
	return e.request
}

// newError creates an Error with the appropriate enum for the code.
func newAPIError(code int32, detail string) api.Error {
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
	return api.Error{
		Code:   &code,
		Title:  codeEnum.Enum(),
		Detail: &detail,
	}
}

// NewRequestError creates an Error with source set to the request url
func NewRequestError(req *http.Request, code int32, detail string) error {
	return &Error{
		silent:   false,
		redirect: false,
		source:   req.URL.String(),
		request:  req,
		err:      newAPIError(code, detail),
	}
}

// NewQuietWrap creates a quiet *wrapped* error that does not return a body
func NewQuietWrap(req *http.Request, code int32, detail string) error {
	return errors.Wrap(&Error{
		silent:   true,
		redirect: false,
		source:   req.URL.String(),
		request:  req,
		err:      newAPIError(code, detail),
	}, 1)
}

// NewQuietError logs the error but does not show it to the user
func NewQuietError(req *http.Request, code int32, e error) error {
	return errors.Wrap(&Error{
		silent:   true,
		redirect: false,
		source:   req.URL.String(),
		request:  req,
		err:      newAPIError(code, errString(e)),
	}, 1)
}

// NewRedirectingError creates an Error with source set to the request url
// and the redirect flag set to true, which will
func NewRedirectingError(req *http.Request, code int32, e error) error {
	return errors.Wrap(&Error{
		silent:   false,
		redirect: true,
		source:   req.URL.String(),
		request:  req,
		err:      newAPIError(code, errString(e)),
	}, 1)
}

// HandleError is the default error handler
func (m *Module) HandleError(rw http.ResponseWriter, req *http.Request, err error) int {
	switch err {
	case ErrNotFound:
		err = NewRequestError(req, http.StatusNotFound, "not found: "+req.URL.String())
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
		// log (warn) the stack trace and then recurse on the original err, which
		// is either a known *Error or unknown error
		m.Logger.Warning(errString(err))
		return m.HandleError(rw, req, e.Err)

	case *Error:
		status := int(e.err.GetCode())
		if status >= 500 {
			m.Logger.Error(err)
		} else {
			m.Logger.Warning(err)
		}
		if e.silent {
			rw.WriteHeader(status)
		} else if e.redirect {
			m.ErrorPage(rw, req, status)
		} else {
			Proto(rw, int(e.err.GetCode()), &api.ErrorResponse{Errors: []*api.Error{&e.err}})
		}
		return status

	default:
		// log error with request
		m.Logger.Error(&Error{
			request: req,
			err:     newAPIError(500, errString(e)),
		})
		ae := newAPIError(http.StatusInternalServerError, "internal server error")
		Proto(rw, http.StatusInternalServerError, &api.ErrorResponse{Errors: []*api.Error{&ae}})
		return http.StatusInternalServerError
	}
}

func (m *Module) Error(rw http.ResponseWriter, status int32, errors ...*api.Error) error {
	for _, err := range errors {
		if err.Code == nil || err.GetCode() >= 500 || status >= 500 {
			m.Logger.Error(err)
		} else {
			m.Logger.Warning(err)
		}
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
