package middleware

import (
	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
)

type ctxKey string

const requestIDKey ctxKey = "requestID"

// Middleware that generates a request ID and sets it in the context + response header
func RequestID(ctx huma.Context, next func(huma.Context)) {
	reqID := uuid.New().String()
	ctx.SetHeader("X-Request-ID", reqID)
	next(ctx)
}

// Helper function to retrieve request ID from context
func GetRequestID(ctx huma.Context) string {
	id, _ := ctx.Context().Value(requestIDKey).(string)
	return id
}
