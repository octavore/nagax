package middleware

import (
	"net/http"
	"sync"
)

type Middleware func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc)

type MiddlewareServer struct {
	serve http.HandlerFunc
	base  http.HandlerFunc

	mu             sync.Mutex // protects middleware list
	middlewareList []Middleware
}

func NewServer(base http.HandlerFunc) *MiddlewareServer {
	return &MiddlewareServer{
		serve:          base,
		base:           base,
		middlewareList: []Middleware{},
	}
}

func (m *MiddlewareServer) Prepend(middleware Middleware) {
	m.mu.Lock()
	m.middlewareList = append([]Middleware{middleware}, m.middlewareList...)
	m.mu.Unlock()
	m.Rebuild()
}

func (m *MiddlewareServer) Append(middleware Middleware) {
	m.mu.Lock()
	m.middlewareList = append(m.middlewareList, middleware)
	m.mu.Unlock()
	m.Rebuild()
}

// Remove middleware

func (m *MiddlewareServer) Set(middlewares ...Middleware) {
	m.mu.Lock()
	m.middlewareList = middlewares
	m.mu.Unlock()
	m.Rebuild()
}

func (m *MiddlewareServer) Rebuild() {
	m.mu.Lock()
	defer m.mu.Unlock()
	serve := m.base
	for i := len(m.middlewareList) - 1; i >= 0; i-- {
		mw := m.middlewareList[i]
		previous := serve
		serve = func(rw http.ResponseWriter, req *http.Request) {
			mw(rw, req, previous)
		}
	}
	m.serve = serve
}

func (m *MiddlewareServer) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	m.serve(rw, r)
}
