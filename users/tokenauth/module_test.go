package tokenauth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/octavore/nagax/router/httperror"
	"github.com/shoenig/test"
	"github.com/shoenig/test/must"
)

type dummyTokenSource map[string]string

func (d *dummyTokenSource) Get(k string) *string {
	v, ok := (*d)[k]
	if !ok {
		return nil
	}
	return &v
}

func setup() (*Module, *httptest.ResponseRecorder, *http.Request) {
	m := &Module{
		tokenSource: &dummyTokenSource{"goodToken": "1234"},
		header:      defaultHTTPHeader,
		prefix:      defaultTokenPrefix,
	}
	return m, httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil)
}

func TestAuthenticateGoodToken(t *testing.T) {
	for _, token := range []string{"Token goodToken", "tOkEn goodToken"} {
		t.Run(token, func(t *testing.T) {
			m, rw, req := setup()
			req.Header.Set(defaultHTTPHeader, token)
			authenticated, userID, err := m.Authenticate(rw, req)

			test.True(t, authenticated, test.Sprint("expected authenticated to be true"))
			must.NotNil(t, userID)
			test.Eq(t, *userID, "1234")
			test.Nil(t, err)
		})
	}
}

func TestAuthenticateBadPrefix(t *testing.T) {
	m, rw, req := setup()
	req.Header.Set(defaultHTTPHeader, "Basic badToken")

	// returns false (not authenticated) but without an error
	authenticated, userID, err := m.Authenticate(rw, req)
	test.False(t, authenticated, test.Sprint("expected authenticated to be false"))
	test.Nil(t, userID, test.Sprint("expected userID to be nil"))
	test.Nil(t, err)
}

func TestAuthenticateBadToken(t *testing.T) {
	m, rw, req := setup()
	req.Header.Set(defaultHTTPHeader, "Token badToken")
	// returns false (not authenticated) with an error
	authenticated, userID, err := m.Authenticate(rw, req)

	code, _ := httperror.CodeFromErr(err)
	test.Eq(t, http.StatusUnauthorized, code)
	test.False(t, authenticated, test.Sprint("expected authenticated to be false"))
	test.Nil(t, userID, test.Sprint("expected userID to be nil"))
}
