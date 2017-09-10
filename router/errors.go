package router

import (
	"net/http"

	"github.com/golang/protobuf/proto"
	"github.com/octavore/nagax/proto/nagax/router/api"
)

// QuietError logs an error and returns the given status without a body
func (m *Module) QuietError(rw http.ResponseWriter, status int, err error) {
	m.Logger.Errorf("%d %s", status, err)
	rw.WriteHeader(status)
}

// SimpleError responds with err.Error() as the detail of a JSON error response.
func (m *Module) SimpleError(rw http.ResponseWriter, status int, err error) error {
	m.Logger.Errorf("%d %s", status, err)
	return m.Error(rw, status, &api.Error{
		Code:   proto.String("error"),
		Title:  proto.String("error"),
		Detail: proto.String(err.Error()),
	})
}

func (m *Module) Error(rw http.ResponseWriter, status int, errors ...*api.Error) error {
	for _, err := range errors {
		m.Logger.Errorf("%d %s", status, err)
	}
	return m.Proto(rw, status, &api.ErrorResponse{
		Errors: errors,
	})
}

// InternalError returns an internal server error
func (m *Module) InternalError(rw http.ResponseWriter) error {
	m.Logger.Error("500 internal server error")
	return m.Proto(rw, http.StatusInternalServerError, &api.Error{
		Code:  proto.String("internal_server_error"),
		Title: proto.String("Internal server error"),
	})
}
