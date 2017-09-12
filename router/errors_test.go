package router

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRequestError_400(t *testing.T) {
	req := httptest.NewRequest("GET", "/test-request-error", nil)
	err := NewRequestError(req, http.StatusBadRequest, "request error")
	expectedStr := `code:400 title:bad_request detail:"request error" `
	assert.EqualError(t, err, "/test-request-error: "+expectedStr)

	env := setup()
	defer env.stop()
	rr := httptest.NewRecorder()
	env.module.HandleError(rr, req, err)
	assert.Equal(t, len(env.logger.Errors), 1)
	assert.Equal(t, env.logger.Errors[0], expectedStr)
	assert.Equal(t, rr.Code, http.StatusBadRequest)
	assert.JSONEq(t, rr.Body.String(), `{
		"errors": [{
			"code": 400,
			"title": "bad_request",
			"detail":"request error"
		}]
	}`)
}

func TestNewQuietWrap_400(t *testing.T) {
	req := httptest.NewRequest("GET", "/test-quiet-error", nil)
	err := NewQuietWrap(req, http.StatusBadRequest, "request error")
	expectedStr := `code:400 title:bad_request detail:"request error" `
	prefixedStr := "/test-quiet-error: " + expectedStr
	assert.EqualError(t, err, prefixedStr)

	env := setup()
	defer env.stop()
	rr := httptest.NewRecorder()
	env.module.HandleError(rr, req, err)
	assert.Equal(t, len(env.logger.Errors), 2)
	// log for wrapper
	assert.Equal(t, env.logger.Errors[0],
		`[github.com/octavore/nagax/router/errors_test.go:36] `+prefixedStr)
	// log for error
	assert.Equal(t, env.logger.Errors[1], prefixedStr+"(quiet)")
	assert.Equal(t, rr.Code, http.StatusBadRequest)
	assert.Equal(t, rr.Body.String(), "")
}

func TestNewQuietError_400(t *testing.T) {
	req := httptest.NewRequest("GET", "/test-quiet-wrap", nil)
	originalErr := fmt.Errorf("hidden error")
	err := NewQuietError(req, http.StatusBadRequest, originalErr)
	expectedStr := `code:400 title:bad_request detail:"hidden error" `
	prefixedStr := "/test-quiet-wrap: " + expectedStr
	assert.EqualError(t, err, prefixedStr)

	env := setup()
	defer env.stop()
	rr := httptest.NewRecorder()
	env.module.HandleError(rr, req, err)
	assert.Equal(t, len(env.logger.Errors), 2)
	// log for wrapper
	assert.Equal(t, env.logger.Errors[0],
		`[github.com/octavore/nagax/router/errors_test.go:58] `+prefixedStr)
	// log for error
	assert.Equal(t, env.logger.Errors[1], prefixedStr+"(quiet)")
	assert.Equal(t, rr.Code, http.StatusBadRequest)
	assert.Equal(t, rr.Body.String(), "")
}

func TestNewRedirectingError_400(t *testing.T) {
	req := httptest.NewRequest("GET", "/test-redirected", nil)
	originalErr := fmt.Errorf("hidden error")
	err := NewRedirectingError(req, http.StatusBadRequest, originalErr)
	expectedStr := `code:400 title:bad_request detail:"hidden error" `
	prefixedStr := "/test-redirected: " + expectedStr
	assert.EqualError(t, err, prefixedStr)

	env := setup()
	defer env.stop()
	rr := httptest.NewRecorder()

	errorPageCalls := 0
	env.module.ErrorPage = func(rw http.ResponseWriter, req *http.Request, status int) {
		assert.Equal(t, status, http.StatusBadRequest)
		http.Redirect(rw, req, "", http.StatusTemporaryRedirect)
		errorPageCalls++
	}

	env.module.HandleError(rr, req, err)
	assert.Equal(t, len(env.logger.Errors), 2)
	assert.Equal(t, env.logger.Errors[0],
		`[github.com/octavore/nagax/router/errors_test.go:80] `+prefixedStr)
	assert.Equal(t, env.logger.Errors[1], prefixedStr+"(redirect)")
	assert.Equal(t, rr.Code, http.StatusTemporaryRedirect)
	assert.Equal(t, rr.Body.String(), "<a href=\"/\">Temporary Redirect</a>.\n\n")
	assert.Equal(t, errorPageCalls, 1)
}
