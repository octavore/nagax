package tokenauth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/octavore/nagax/users"
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
	m, rw, req := setup()
	req.Header.Set(defaultHTTPHeader, "Token goodToken")
	b, s, err := m.Authenticate(rw, req)
	if err != nil {
		t.Fatal("unexpected error", err)
	}
	if !b {
		t.Error("unexpected value", b)
	}
	if s == nil {
		t.Error("unexpected value", s)
	} else if *s != "1234" {
		t.Error("unexpected value", *s)
	}

	// different capitalization for prefix
	m, rw, req = setup()
	req.Header.Set(defaultHTTPHeader, "tOkEn goodToken")
	b, s, err = m.Authenticate(rw, req)
	if err != nil {
		t.Fatal("unexpected error", err)
	}
	if !b {
		t.Error("unexpected value", b)
	}
	if s == nil {
		t.Error("unexpected value", s)
	} else if *s != "1234" {
		t.Error("unexpected value", *s)
	}
}

func TestAuthenticateBadPrefix(t *testing.T) {
	m, rw, req := setup()
	req.Header.Set(defaultHTTPHeader, "Basic badToken")

	// returns false (not authenticated) but without an error
	b, s, err := m.Authenticate(rw, req)
	if err != nil {
		t.Fatal("unexpected error", nil)
	}
	if b {
		t.Error("unexpected value", b)
	}
	if s != nil {
		t.Error("unexpected value", *s)
	}
}

func TestAuthenticateBadToken(t *testing.T) {
	m, rw, req := setup()
	req.Header.Set(defaultHTTPHeader, "Token badToken")
	// returns false (not authenticated) with an error
	b, s, err := m.Authenticate(rw, req)
	if err != users.ErrNotAuthorized {
		t.Fatal("unexpected error", err)
	}
	if b {
		t.Error("unexpected value", b)
	}
	if s != nil {
		t.Error("unexpected value", *s)
	}
}
