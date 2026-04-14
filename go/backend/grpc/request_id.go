package grpc

import (
	"context"

	"google.golang.org/grpc/metadata"
)

type ctxKeyRequestIdType string

var ctxKeyRequestId = ctxKeyRequestIdType("request-id")

// RequestIdFromMetadata obtains the value of the "x-request-id" key from the metadata
func RequestIdFromMetadata(meta metadata.MD) string {
	if meta == nil {
		return ""
	}

	requestId := ""
	xRequestId := meta.Get("x-request-id")
	if len(xRequestId) > 0 {
		requestId = xRequestId[0]
	}
	return requestId
}

// RequestIdFromContext returns the request id from the context
func RequestIdFromContext(ctx context.Context) string {
	val := ctx.Value(ctxKeyRequestId)
	if val == nil {
		return ""
	}

	reqId := val.(string)
	return reqId
}
