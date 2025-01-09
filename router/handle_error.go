package router

import (
	"net/http"

	"github.com/go-errors/errors"

	"github.com/octavore/nagax/proto/router/api"
	"github.com/octavore/nagax/router/httperror"
)

// HandleError is the base error handler for the router.
// 1. If err is a httperror.HTTPErrorCode, only the error status code is returned, without a body
// 2. If the route is not an API route, m.ErrorPage is called to show an error page
// 3. If err is a httperror.HTTPError, its ToProto function will be called for the return JSON
// 4. Otherwise, we will return a JSON response without any detail (probably a 500 unless err implements GetCode)
// *  If the final status code is 500, we will report the original err with m.Logger.ErrorCtx
func (m *Module) HandleError(rw http.ResponseWriter, req *http.Request, err error) int {
	statusCode, _ := httperror.CodeFromErr(err)
	logLine := newHandlerErrorLogBuilder(req, statusCode)

	var httpErrCode httperror.HTTPErrorCode

	if errors.As(err, &httpErrCode) {
		// 1. HTTPErrorCode: return error code only
		m.Logger.InfoCtx(req.Context(), logLine.WithAction("error-code"))
		rw.WriteHeader(statusCode)

	} else if !m.IsAPIRoute(req) {
		// 2. non-api routes show an error page
		m.Logger.InfoCtx(req.Context(), logLine.WithAction("error-page").WithError(err))
		m.ErrorPage(rw, req, statusCode, httperror.UnwrapAll(err))

	} else {
		// 3. api route
		var httpErr *httperror.HTTPError
		if !errors.As(err, &httpErr) {
			// if not a HTTPError, convert to one
			httpErr = &httperror.HTTPError{Code: statusCode, BaseError: err}
		}

		m.Logger.InfoCtx(req.Context(), logLine.WithAction("error-json").WithDetail(httpErr.Detail).WithError(err))
		protoErr := Proto(rw, statusCode, &api.ErrorResponse{Errors: []*api.Error{httpErr.ToProto()}})
		if protoErr != nil {
			m.Logger.ErrorCtx(req.Context(), protoErr)
		}
	}

	// log errors with errorCtx
	if statusCode >= 500 {
		// note: this returns the original error because it may have more context, eg the stack trace.
		m.Logger.ErrorCtx(req.Context(), err)
	}

	return statusCode
}
