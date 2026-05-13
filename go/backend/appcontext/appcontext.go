package appcontext

import "context"

type contextKey int

const (
	KeyRequestID contextKey = iota
)

func RequestIDFromContext(ctx context.Context) string {
	val := ctx.Value(KeyRequestID)
	if val == nil {
		return ""
	}
	return val.(string)
}

func WithRequestID(ctx context.Context, reqID string) context.Context {
	return context.WithValue(ctx, KeyRequestID, reqID)
}
