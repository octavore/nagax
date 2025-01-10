package router

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-errors/errors"
	"github.com/shoenig/test"
	"github.com/shoenig/test/must"

	"github.com/octavore/nagax/router/httperror"
	nerrors "github.com/octavore/nagax/util/errors"
)

type CustomError struct{}

func (c *CustomError) Error() string {
	return "custom error"
}

func TestHandleError(t *testing.T) {
	env := setup()
	defer env.stop()

	errWithStack := nerrors.New("has stack")
	customErrWithStack := nerrors.Wrap(&CustomError{})
	testCases := []struct {
		desc         string
		err          error
		expectedCode int
		expectedBody string
		expectedLog  string
	}{{
		desc:         "fmt-errorf",
		err:          fmt.Errorf("non-httperror"),
		expectedCode: 500,
		expectedLog:  `[500] /api/test error-json detail:"" error:"non-httperror" error-type:*errors.errorString`,
		expectedBody: `{
			"errors": [{
				"code": 500,
				"title": "internal_server_error"
			}]
		}`,
	}, {
		desc:         "wrapped-error",
		err:          errWithStack,
		expectedCode: 500,
		expectedLog:  `[500] /api/test error-json detail:"" error:"has stack" error-type:*errors.errorString loc:github.com/octavore/nagax/router/handle_error_test.go|27`,
		expectedBody: `{
			"errors": [{
				"code": 500,
				"title": "internal_server_error"
			}]
		}`,
	}, {
		desc:         "error-code-only",
		err:          httperror.HTTPErrorCode(403),
		expectedCode: 403,
		expectedLog:  `[403] /api/test error-code`,
	}, {
		desc:         "httperror",
		err:          httperror.NotFound("Resource not found."),
		expectedCode: 404,
		expectedBody: `{
			"errors": [{
				"code": 404,
				"title": "not_found",
				"detail":"Resource not found."
			}]
		}`,
		expectedLog: `[404] /api/test error-json detail:"Resource not found." error:<nil>`,
	}, {
		desc:         "httperror-with-error",
		err:          httperror.BadRequest("This is a bad request.").WithError(fmt.Errorf("hidden error")),
		expectedCode: 400,
		expectedBody: `{
			"errors": [{
				"code": 400,
				"title": "bad_request",
				"detail":"This is a bad request."
			}]
		}`,
		expectedLog: `[400] /api/test error-json detail:"This is a bad request." error:"hidden error" error-type:*errors.errorString`,
	}, {
		desc:         "httperror-with-error-with-stack",
		err:          httperror.InternalError().WithError(errWithStack),
		expectedCode: 500,
		expectedBody: `{
			"errors": [{
				"code": 500,
				"title": "internal_server_error"
			}]
		}`,
		expectedLog: `[500] /api/test error-json detail:"" error:"has stack" error-type:*errors.errorString loc:github.com/octavore/nagax/router/handle_error_test.go|27`,
	}, {
		desc:         "httperror-with-custom-error-with-stack",
		err:          httperror.InternalError().WithDetail("Another message.").WithError(customErrWithStack),
		expectedCode: 500,
		expectedBody: `{
			"errors": [{
				"code": 500,
				"title": "internal_server_error",
				"detail":"Another message."
			}]
		}`,
		expectedLog: `[500] /api/test error-json detail:"Another message." error:"custom error" error-type:*router.CustomError loc:github.com/octavore/nagax/router/handle_error_test.go|28`,
	}}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/test", nil)
			rr := httptest.NewRecorder()
			env.logger.Reset()
			env.module.HandleError(rr, req, tc.err)

			test.Eq(t, tc.expectedCode, rr.Code)
			if tc.expectedBody == "" {
				test.Eq(t, "", rr.Body.String())
			} else {
				test.EqJSON(t, tc.expectedBody, rr.Body.String())
			}

			if tc.expectedCode >= 500 {
				test.Eq(t, env.logger.Errors, []string{tc.err.Error()})
			} else {
				test.SliceEmpty(t, env.logger.Errors)
			}
			test.SliceEmpty(t, env.logger.Warnings)
			must.SliceLen(t, 1, env.logger.Infos)
			test.Eq(t, env.logger.Infos[0], tc.expectedLog)
		})
	}
}

func TestHandleErrorNonAPIRoute(t *testing.T) {
	env := setup()
	defer env.stop()

	testCases := []struct {
		err          error
		expectedCode int
	}{{
		err:          fmt.Errorf("non-httperror"),
		expectedCode: 500,
	}, {
		err:          httperror.NotFound("not found"),
		expectedCode: 404,
	}}

	for _, tc := range testCases {
		t.Run(tc.err.Error(), func(t *testing.T) {
			// track number of calls to ErrorPage
			errorPageCalls := 0
			env.module.ErrorPage = func(rw http.ResponseWriter, req *http.Request, status int, err error) {
				errorPageCalls++
				test.Eq(t, tc.expectedCode, status, test.Sprintf("expected status code %d from err but got %d", tc.expectedCode, status))

				// assert we got the unwrapped error
				test.ErrorIs(t, err, tc.err, test.Sprintf("expected baseErr but got wrappedErr"))
				http.Redirect(rw, req, "", http.StatusTemporaryRedirect)
			}

			req := httptest.NewRequest("GET", "/test-non-api-route", nil)
			rr := httptest.NewRecorder()
			env.logger.Reset()
			env.module.HandleError(rr, req, errors.New(tc.err)) // invoke with wrapped error

			test.Eq(t, http.StatusTemporaryRedirect, rr.Code)
			test.Eq(t, "<a href=\"/\">Temporary Redirect</a>.\n\n", rr.Body.String())
			test.Eq(t, 1, errorPageCalls)
			if tc.expectedCode >= 500 {
				test.Eq(t, env.logger.Errors, []string{tc.err.Error()})
			} else {
				test.SliceEmpty(t, env.logger.Errors)
			}
		})
	}

	// test.SliceEmpty(t, env.logger.Errors)
	// test.SliceEmpty(t, env.logger.Infos)
	// test.SliceEmpty(t, env.logger.Warnings)
	// expectedStr := `/test-redirected: code:302 title:found detail:"hidden error" (redirect)`
	// assert.EqualError(t, err, expectedStr)
	// assert.ElementsMatch(t, []string{expectedStr}, env.logger.Warnings)
}
