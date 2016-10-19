package users

import (
	"context"
	"errors"
	"net/http"
)

type UserTokenKey struct{}

func (m *Module) WithAuth(next http.HandlerFunc) http.HandlerFunc {
	return m.WithAuthList(m.Authenticators, next)
}

func (m *Module) MustWithAuth(next http.HandlerFunc) http.HandlerFunc {
	return m.MustWithAuthList(m.Authenticators, next)
}

func (m *Module) WithAuthList(authenticators []Authenticator, next http.HandlerFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		_, userToken, err := m.authenticate(authenticators, rw, req)
		if err != nil {
			m.ErrorHandler(rw, req, err)
			return
		}
		if userToken != nil {
			ctx := context.WithValue(req.Context(), UserTokenKey{}, *userToken)
			req = req.WithContext(ctx)
		}
		next(rw, req)
	}
}

var ErrNotAuthorized = errors.New("not authorized")

func (m *Module) MustWithAuthList(authenticators []Authenticator, next http.HandlerFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		handled, userToken, err := m.authenticate(authenticators, rw, req)
		if err != nil {
			m.ErrorHandler(rw, req, err)
			return
		}

		if !handled || userToken == nil {
			m.ErrorHandler(rw, req, ErrNotAuthorized)
			return
		}

		ctx := context.WithValue(req.Context(), UserTokenKey{}, *userToken)
		next(rw, req.WithContext(ctx))
	}
}
