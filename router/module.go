package router

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/octavore/naga/service"

	"github.com/octavore/nagax/config"
	"github.com/octavore/nagax/logger"
	"github.com/octavore/nagax/proto/nagax/router/api"
)

// Config for the router module
type Config struct {
	Port int `json:"port"`
}

// Module router implements basic routing with helpers for protobuf-based responses.
type Module struct {
	*http.ServeMux
	Logger *logger.Module
	Config *config.Module
	config Config
}

// Init implements service.Init
func (m *Module) Init(c *service.Config) {
	c.Setup = func() error {
		m.ServeMux = http.NewServeMux()
		m.Config.ReadConfig(&m.config)
		return nil
	}
	c.Start = func() {
		port := 8000
		if m.config.Port != 0 {
			port = m.config.Port
		}
		laddr := fmt.Sprintf("127.0.0.1:%d", port)
		log.Println("listening on", laddr)
		go http.ListenAndServe(laddr, m)
	}
}

var jpb = &jsonpb.Marshaler{
	EnumsAsInts: false,
	Indent:      "  ",
}

// ProtoOK renders a 200 response with JSON-serialized proto
func (m *Module) ProtoOK(rw http.ResponseWriter, pb proto.Message) error {
	return m.Proto(rw, http.StatusOK, pb)
}

// Proto renders a response with given status code and JSON-serialized proto
func (m *Module) Proto(rw http.ResponseWriter, status int, pb proto.Message) error {
	rw.WriteHeader(status)
	return jpb.Marshal(rw, pb)
}

// JSON renders a response with given status and JSON serialized data
func (m *Module) JSON(rw http.ResponseWriter, status int, v interface{}) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	rw.Header().Add("Content-Type", "application/json")
	rw.WriteHeader(status)
	_, err = rw.Write(b)
	return err
}

// EmptyJSON renders a 200 response with JSON body `{}`
func (m *Module) EmptyJSON(rw http.ResponseWriter, status int) error {
	rw.Header().Add("Content-Type", "application/json")
	rw.WriteHeader(status)
	_, err := rw.Write([]byte(`{}`))
	return err
}

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
