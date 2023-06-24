package router

import (
	"fmt"
	"net/http"
	"path"

	"github.com/go-errors/errors"

	"github.com/octavore/nagax/proto/router/api"
	"github.com/octavore/nagax/router/httperror"
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
	} else if e.request != nil {
		errStr = e.request.URL.Path + ": " + errStr
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

func (e *Error) GetCode() int {
	return int(e.err.GetCode())
}

func (e *Error) Source() api.Error {
	return e.err
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
func NewRedirectingError(req *http.Request, e error) error {
	return errors.Wrap(&Error{
		silent:   false,
		redirect: true,
		source:   req.URL.String(),
		request:  req,
		err:      newAPIError(http.StatusFound, errString(e)),
	}, 1)
}

type GetCoder interface {
	GetCode() int
}

var _ GetCoder = &Error{}

func GetErrorCode(err error) int {
	err, _ = unwrapGoError(err)
	if e, ok := err.(GetCoder); ok {
		return e.GetCode()
	}

	// check for err constants
	switch err {
	case ErrNotFound:
		return http.StatusNotFound
	case ErrNotAuthorized:
		return http.StatusUnauthorized
	case ErrForbidden:
		return http.StatusForbidden
	case ErrInternal:
		return http.StatusInternalServerError
	}
	return http.StatusInternalServerError
}

// HandleError is the base error handler for the router
func (m *Module) HandleError(rw http.ResponseWriter, req *http.Request, err error) int {
	// 1. unwrap error
	originalErr := err
	unwrappedErr, _ := unwrapGoError(err)

	// 2. convert known errors to router error
	switch unwrappedErr {
	case ErrNotFound:
		unwrappedErr = httperror.NotFound("Not found: " + req.URL.String())
	case ErrNotAuthorized:
		unwrappedErr = httperror.NotAuthorized("Not authenticated")
	case ErrForbidden:
		unwrappedErr = httperror.Forbidden("Forbidden")
	case ErrInternal:
		unwrappedErr = httperror.InternalError("Internal server error")
	}

	// 3. convert errors to router.Error and handle unknown errors
	var routerErr *Error
	if h, ok := unwrappedErr.(*httperror.HTTPError); ok {
		routerErr = &Error{
			source:  req.URL.String(),
			request: req,
			err:     newAPIError(h.Code, h.Detail),
		}
	} else {
		routerErr, ok = unwrappedErr.(*Error)
		if !ok /* handling an unknown error */ {
			routerErr = &Error{
				source:  req.URL.String(),
				request: req,
				err:     newAPIError(500, "internal server error"),
			}
		}
	}

	// 4. log errors
	status := routerErr.GetCode()
	if status >= 500 {
		routerErr.err = newAPIError(int32(status), "internal server error")
		m.Logger.Error(originalErr) // log the original error (with request)
	} else {
		m.Logger.Warning(originalErr)
	}

	// 5.handle errors
	switch {
	case routerErr.silent:
		rw.WriteHeader(status)
	case routerErr.redirect || !m.IsAPIRoute(req):
		m.ErrorPage(rw, req, status, unwrappedErr)
	default:
		Proto(rw, status, &api.ErrorResponse{Errors: []*api.Error{&routerErr.err}})
	}
	return status
}

// Error writes an an API error to rw
func (m *Module) Error(rw http.ResponseWriter, status int, errors ...*api.Error) error {
	for _, err := range errors {
		if err.Code == nil || err.GetCode() >= 500 || status >= 500 {
			m.Logger.Error(err)
		} else {
			m.Logger.Warning(err)
		}
	}
	return Proto(rw, status, &api.ErrorResponse{
		Errors: errors,
	})
}

// errString prints out an error, with its location if appropriate
func errString(err error) string {
	err, ok := unwrapGoError(err)
	if !ok {
		return err.Error()
	}
	e := err.(*errors.Error)
	s := e.StackFrames()[0]
	f := path.Base(s.File)
	prefix := fmt.Sprintf("[%s/%s:%d]", s.Package, f, s.LineNumber)
	if e.Error() == "" {
		return prefix
	}
	return fmt.Sprint(prefix, " ", e.Error())
}

func unwrapGoError(err error) (error, bool) {
	wrappedErr, isWrapped := err.(*errors.Error)
	if isWrapped {
		unwrapped, _ := unwrapGoError(wrappedErr.Err)
		return unwrapped, true
	}
	return err, false
}
