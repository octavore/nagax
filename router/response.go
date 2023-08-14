package router

import (
	"encoding/json"
	"net/http"

	"github.com/octavore/nagax/util/errors"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

var jpb = &protojson.MarshalOptions{
	UseEnumNumbers: false,
	Indent:         "  ",
}

// ProtoOK renders a 200 response with JSON-serialized proto
func ProtoOK(rw http.ResponseWriter, pb proto.Message) error {
	return Proto(rw, http.StatusOK, pb)
}

// Proto renders a response with given status code and JSON-serialized proto
func Proto(rw http.ResponseWriter, status int, pb proto.Message) error {
	if pb == nil {
		return EmptyJSON(rw, status)
	}
	data, err := jpb.Marshal(pb)
	if err != nil {
		return errors.Wrap(err)
	}
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(status)
	_, err = rw.Write(data)
	return errors.Wrap(err)
}

// JSON renders a response with given status and JSON serialized data
func JSON(rw http.ResponseWriter, status int, v interface{}) error {
	if v == nil {
		return EmptyJSON(rw, status)
	}
	if pb, ok := v.(proto.Message); ok {
		return Proto(rw, status, pb)
	}
	rw.Header().Add("Content-Type", "application/json")
	rw.WriteHeader(status)
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	_, err = rw.Write(b)
	return errors.Wrap(err)
}

// EmptyJSON renders a response with the given status and JSON body `{}`
func EmptyJSON(rw http.ResponseWriter, status int) error {
	rw.Header().Add("Content-Type", "application/json")
	rw.WriteHeader(status)
	_, err := rw.Write([]byte(`{}`))
	return errors.Wrap(err)
}
