package http

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"
)

type requestIDKey struct{}

func generateRequestID() string {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fallbackRequestID()
	}
	return hex.EncodeToString(b[:])
}

func fallbackRequestID() string {
	return hex.EncodeToString([]byte(time.Now().Format("20060102150405.000000000")))
}

func requestIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if val, ok := ctx.Value(requestIDKey{}).(string); ok {
		return val
	}
	return ""
}

func withRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey{}, id)
}
