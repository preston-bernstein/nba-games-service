package metrics

import (
	"context"
	"testing"
	"time"
)

func TestSetupDisabledReturnsNoHandler(t *testing.T) {
	rec, handler, shutdown, err := Setup(context.Background(), TelemetryConfig{
		Enabled: false,
	})
	if err != nil {
		t.Fatalf("expected no error when disabled, got %v", err)
	}
	if rec == nil {
		t.Fatalf("expected recorder")
	}
	if handler != nil {
		t.Fatalf("expected nil handler when disabled")
	}
	if shutdown == nil {
		t.Fatalf("expected shutdown function")
	}
}

func TestSetupEnabledInitializesRecorderAndHandler(t *testing.T) {
	rec, handler, shutdown, err := Setup(context.Background(), TelemetryConfig{
		Enabled:     true,
		ServiceName: "nba-games-service",
		// No OTLP endpoint; uses Prometheus exporter only.
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if rec == nil {
		t.Fatalf("expected recorder")
	}
	if handler == nil {
		t.Fatalf("expected handler when enabled")
	}
	if shutdown == nil {
		t.Fatalf("expected shutdown function")
	}

	// Exercise otel-backed recorders to ensure no panic.
	rec.RecordHTTPRequest("GET", "/health", 200, time.Millisecond)
	rec.RecordPollerCycle(time.Millisecond, nil)
	rec.RecordProviderAttempt("balldontlie", time.Millisecond, nil)
	rec.RecordRateLimit("balldontlie", time.Second)
}
