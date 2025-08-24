package middleware

import (
	"log/slog"
	"os"
	"time"

	"github.com/danielgtaylor/huma/v2"
)

// actual logger
var requestLogger *slog.Logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
	Level: slog.LevelInfo,
}))

// middleware
func RequestLogger(ctx huma.Context, next func(huma.Context)) {
	start := time.Now()

	next(ctx)

	duration := time.Since(start)
	status := ctx.Status()
	path := ctx.Operation().Path
	method := ctx.Method()

	if status >= 400 {
		requestLogger.Error("request failed",
			slog.String("request_id", GetRequestID(ctx)),
			slog.String("method", method),
			slog.String("path", path),
			slog.Int("status", status),
			slog.Int64("duration_ms", duration.Milliseconds()),
		)
	} else {
		requestLogger.Debug("request completed",
			slog.String("request_id", GetRequestID(ctx)),
			slog.String("method", method),
			slog.String("path", path),
			slog.Int("status", status),
			slog.Int64("duration_ms", duration.Milliseconds()),
		)
	}
}
