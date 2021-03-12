package logging

import (
	"context"

	"go.uber.org/zap"
)

type contextKeyLogging int

const keyLogger = contextKeyLogging(0)
const keyRequestID = contextKeyLogging(1)

// WithLogger will attach the given logger to a parent context.
func WithLogger(parent context.Context, logger *zap.Logger) context.Context {
	if parent == nil {
		parent = context.Background()
	}
	if logger == nil {
		return parent
	}
	return context.WithValue(parent, keyLogger, logger)
}

// GetLogger will retrieve a logger from the given context.
func GetLogger(ctx context.Context) *zap.Logger {
	if ctx == nil {
		return zap.NewNop()
	}
	if logger, ok := ctx.Value(keyLogger).(*zap.Logger); ok {
		return logger
	}
	return zap.NewNop()
}

// GetRequestID will return the generated request id that the logging.Middleware
// has generated and attached to the http.Request's context.
func GetRequestID(ctx context.Context) string {
	if ctx == nil || ctx.Value(keyRequestID) == nil {
		return ""
	}
	if id, ok := ctx.Value(keyRequestID).(string); ok {
		return id
	}
	return ""
}
