package auth_router

import (
	"io"
	"net/http"
	"reflect"
	"strings"

	"github.com/julienschmidt/httprouter"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/octavore/nagax/router"
	"github.com/octavore/nagax/users"
	"github.com/octavore/nagax/util/errors"
)

type Route struct {
	method  string
	path    string
	handler interface{}
	version string
}

func Register[Req proto.Message, Res proto.Message, A interface{}](
	m *Module[A],
	pathSpec string,
	handler func(auth *A, par router.Params, reqpb Req) (Res, error),
	authenticators ...users.Authenticator,
) {
	method, path := parsePathSpec(pathSpec)
	reqType := reflect.TypeOf(handler).In(2)
	m.routeRegistry = append(m.routeRegistry, &Route{
		method:  method,
		path:    path,
		handler: handler,
		version: "proto",
	})

	h := func(rw http.ResponseWriter, req *http.Request, par httprouter.Params) error {
		reqpb := reflect.New(reqType.Elem()).Interface().(Req)
		auth, err := m.authAndParseRequest(req, reqpb, authenticators)
		if err != nil {
			return errors.Wrap(err)
		}
		res, err := handler(auth, par, reqpb)
		if err != nil {
			return errors.Wrap(err)
		}
		return router.ProtoOK(rw, res)
	}

	switch method {
	case http.MethodGet:
		m.Router.GET(path, m.RequireAuth(h, authenticators...))
	case http.MethodPost:
		m.Router.POST(path, m.RequireAuth(h, authenticators...))
	case http.MethodDelete:
		m.Router.DELETE(path, m.RequireAuth(h, authenticators...))
	case http.MethodPut:
		m.Router.PUT(path, m.RequireAuth(h, authenticators...))
	case http.MethodPatch:
		m.Router.PATCH(path, m.RequireAuth(h, authenticators...))
	default:
		panic("Unsupported method: " + method)
	}
}

var validMethods = map[string]bool{
	http.MethodGet:    true,
	http.MethodPost:   true,
	http.MethodDelete: true,
	http.MethodPatch:  true,
	http.MethodPut:    true,
}

func parsePathSpec(pathSpec string) (method, path string) {
	method, path, ok := strings.Cut(pathSpec, " ")
	if !ok {
		panic("auth_router: expected pathSpec to have format '<method> <path>'; got " + pathSpec)
	}
	path = strings.TrimSpace(path)
	if !validMethods[method] {
		panic("auth_router: invalid method " + method + " for pathSpec " + pathSpec)
	}
	return method, path
}

func (m *Module[A]) authAndParseRequest(req *http.Request, pb proto.Message, authenticators []users.Authenticator) (*A, error) {
	auth, err := m.GetAuthSession(req)
	if err != nil {
		// if nil auth is not allowed, you should configure m.GetAuthSession accordingly
		return nil, errors.Wrap(err)
	}
	_, isEmpty := pb.(*emptypb.Empty)
	if pb != nil && !isEmpty {
		data, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, errors.Wrap(err)
		}
		err = protojson.Unmarshal(data, pb)
		if err != nil {
			return nil, errors.Wrap(err)
		}
		if err != nil {
			return nil, errors.New("[%s %s] error decoding json body: %v", req.Method, req.URL.Path, err)
		}
	}
	return auth, nil
}
