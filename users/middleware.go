package users

import (
	"context"
	"net/http"
)

type UserTokenKey struct{}

func (m *Module) AuthenticatedFunc(next http.HandlerFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		userToken, err := m.Authenticate(req)
		if err != nil {
			http.Redirect(rw, req, "/", http.StatusForbidden)
			return
		}
		ctx := context.WithValue(req.Context(), UserTokenKey{}, userToken)
		next(rw, req.WithContext(ctx))
	}
}

func (m *Module) AuthenticatedHandler(next http.Handler) http.HandlerFunc {
	return m.AuthenticatedFunc(next.ServeHTTP)
}
