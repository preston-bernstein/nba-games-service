package server

import (
	"context"
	"net/http"
)

// httpServer abstracts the HTTP server implementation for easier testing.
type httpServer interface {
	ListenAndServe() error
	Shutdown(context.Context) error
	Addr() string
	Handler() http.Handler
}

type netHTTPServer struct {
	srv *http.Server
}

func (s netHTTPServer) ListenAndServe() error              { return s.srv.ListenAndServe() }
func (s netHTTPServer) Shutdown(ctx context.Context) error { return s.srv.Shutdown(ctx) }
func (s netHTTPServer) Addr() string                       { return s.srv.Addr }
func (s netHTTPServer) Handler() http.Handler              { return s.srv.Handler }
