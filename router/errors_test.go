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
	expectedStr := `/test-request-error: code:400 title:bad_request detail:"request error"`
	assert.EqualError(t, err, expectedStr)

	env := setup()
	defer env.stop()
	rr := httptest.NewRecorder()
	env.module.HandleError(rr, req, err)
	assert.Equal(t, len(env.logger.Warnings), 1)
	assert.Equal(t, env.logger.Warnings[0], expectedStr)
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
	expectedStr := `/test-quiet-error: code:400 title:bad_request detail:"request error" (silent)`
	assert.EqualError(t, err, expectedStr)

	env := setup()
	defer env.stop()
	rr := httptest.NewRecorder()
	env.module.HandleError(rr, req, err)
	assert.Equal(t, 0, len(env.logger.Errors))
	assert.Equal(t, 1, len(env.logger.Warnings))
	// log for wrapper
	assert.ElementsMatch(t, []string{expectedStr}, env.logger.Warnings)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Equal(t, "", rr.Body.String())
}

func TestNewQuietError_400(t *testing.T) {
	req := httptest.NewRequest("GET", "/test-quiet-wrap", nil)
	originalErr := fmt.Errorf("hidden error")
	err := NewQuietError(req, http.StatusBadRequest, originalErr)
	expectedStr := `/test-quiet-wrap: code:400 title:bad_request detail:"hidden error" (silent)`
	assert.EqualError(t, err, expectedStr)

	env := setup()
	defer env.stop()
	rr := httptest.NewRecorder()
	env.module.HandleError(rr, req, err)
	assert.Equal(t, 0, len(env.logger.Errors))
	// log for wrapper
	assert.ElementsMatch(t, []string{expectedStr}, env.logger.Warnings)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Equal(t, "", rr.Body.String())
}

func TestNewRedirectingError_400(t *testing.T) {
	req := httptest.NewRequest("GET", "/test-redirected", nil)
	originalErr := fmt.Errorf("hidden error")
	err := NewRedirectingError(req, originalErr)
	expectedStr := `/test-redirected: code:302 title:found detail:"hidden error" (redirect)`
	assert.EqualError(t, err, expectedStr)

	env := setup()
	defer env.stop()
	rr := httptest.NewRecorder()

	errorPageCalls := 0
	env.module.ErrorPage = func(rw http.ResponseWriter, req *http.Request, status int, err error) {
		assert.Equal(t, http.StatusFound, status)
		http.Redirect(rw, req, "", http.StatusTemporaryRedirect)
		errorPageCalls++
	}

	env.module.HandleError(rr, req, err)
	assert.Equal(t, 0, len(env.logger.Errors))
	assert.ElementsMatch(t, []string{expectedStr}, env.logger.Warnings)

	assert.Equal(t, http.StatusTemporaryRedirect, rr.Code)
	assert.Equal(t, "<a href=\"/\">Temporary Redirect</a>.\n\n", rr.Body.String())
	assert.Equal(t, 1, errorPageCalls)
}
