package router

import (
	"log"
	"net/http"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/octavore/naga/service"
	"github.com/octavore/nagax/proto/nagax/router/api"
)

type Module struct {
	*http.ServeMux
}

func (r *Module) Init(c *service.Config) {
	c.Setup = func() error {
		r.ServeMux = http.NewServeMux()
		return nil
	}
	c.Start = func() {
		laddr := "127.0.0.1:8000"
		log.Println("listening on", laddr)
		go http.ListenAndServe(laddr, r)
	}
}

var json = &jsonpb.Marshaler{
	EnumsAsInts: false,
	Indent:      "  ",
}

func ProtoOK(rw http.ResponseWriter, pb proto.Message) error {
	return Proto(rw, http.StatusOK, pb)
}

func Proto(rw http.ResponseWriter, status int, pb proto.Message) error {
	rw.WriteHeader(status)
	return json.Marshal(rw, pb)
}

func Error(rw http.ResponseWriter, status int, errors ...*api.Error) error {
	return Proto(rw, status, &api.ErrorResponse{
		Errors: errors,
	})
}

func InternalError(rw http.ResponseWriter) error {
	return Proto(rw, http.StatusInternalServerError, &api.Error{
		Code:  proto.String("internal_server_error"),
		Title: proto.String("Internal server error"),
	})
}
