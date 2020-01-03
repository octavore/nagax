package users

import (
	"context"
	"net/http"

	"github.com/octavore/nagax/router"
)

type UserTokenKey struct{}

func (m *Module) WithAuth(next router.Handle) router.Handle {
	return m.WithAuthList(m.Authenticators, next)
}

func (m *Module) MustWithAuth(next router.Handle) router.Handle {
	return m.MustWithAuthList(m.Authenticators, next)
}

func (m *Module) WithAuthList(authenticators []Authenticator, next router.Handle) router.Handle {
	return func(rw http.ResponseWriter, req *http.Request, par router.Params) error {
		_, userToken, err := m.AuthenticateWithList(authenticators, rw, req)
		if err != nil {
			return err
		}
		if userToken != nil {
			ctx := context.WithValue(req.Context(), UserTokenKey{}, *userToken)
			req = req.WithContext(ctx)
		}
		return next(rw, req, par)
	}
}

func (m *Module) MustWithAuthList(authenticators []Authenticator, next router.Handle) router.Handle {
	return func(rw http.ResponseWriter, req *http.Request, par router.Params) error {
		handled, userToken, err := m.AuthenticateWithList(authenticators, rw, req)
		if err != nil {
			return err
		}

		if !handled || userToken == nil {
			return router.ErrNotAuthorized
		}

		ctx := context.WithValue(req.Context(), UserTokenKey{}, *userToken)
		return next(rw, req.WithContext(ctx), par)
	}
}
