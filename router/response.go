package router

import (
	"encoding/json"
	"net/http"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
)

var jpb = &jsonpb.Marshaler{
	EnumsAsInts: false,
	Indent:      "  ",
}

// ProtoOK renders a 200 response with JSON-serialized proto
func ProtoOK(rw http.ResponseWriter, pb proto.Message) error {
	return Proto(rw, http.StatusOK, pb)
}

// Proto renders a response with given status code and JSON-serialized proto
func Proto(rw http.ResponseWriter, status int, pb proto.Message) error {
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(status)
	return jpb.Marshal(rw, pb)
}

// JSON renders a response with given status and JSON serialized data
func JSON(rw http.ResponseWriter, status int, v interface{}) error {
	if pb, ok := v.(proto.Message); ok {
		return Proto(rw, status, pb)
	}

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
