package server

import (
	"context"
	"net/http"
	"testing"
)

func TestNetHTTPServerHelpers(t *testing.T) {
	srv := &http.Server{
		Addr:    ":0",
		Handler: http.NewServeMux(),
	}
	wrapper := netHTTPServer{srv: srv}

	if wrapper.Handler() == nil {
		t.Fatalf("expected handler")
	}
	if wrapper.Addr() == "" {
		t.Fatalf("expected addr")
	}
	// Shutdown without starting should be safe.
	if err := wrapper.Shutdown(context.Background()); err != nil && err != http.ErrServerClosed {
		t.Fatalf("expected no fatal error on shutdown, got %v", err)
	}
}
