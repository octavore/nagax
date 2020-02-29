package router

import "net/http"

func BadRequest(req *http.Request, message string) error {
	return NewRequestError(req, http.StatusBadRequest, message)
}

func NotFound(req *http.Request, message string) error {
	return NewRequestError(req, http.StatusNotFound, message)
}

func Forbidden(req *http.Request) error {
	return NewRequestError(req, http.StatusForbidden, "forbidden")
}

func Internal(req *http.Request) error {
	return NewRequestError(req, http.StatusInternalServerError, "internal server error")
}
