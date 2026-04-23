package grpc

import (
	"context"
	"fmt"
	"regexp"

	"google.golang.org/grpc/metadata"
)

const metaRequestIdKey = "x-request-id"

var requestIdValidationPattern = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
var requestIdValidationMaxLen = 64

type ctxKeyRequestIdType string

var ctxKeyRequestId = ctxKeyRequestIdType("request-id")

// RequestIdFromMetadata obtains the value of the headerRequestId key from the metadata
func RequestIdFromMetadata(meta metadata.MD) (string, error) {
	if meta == nil {
		return "", nil
	}

	requestId := ""
	headerRequestId := meta.Get(metaRequestIdKey)
	if len(headerRequestId) > 0 {
		requestId = headerRequestId[0]
		if len(requestId) > requestIdValidationMaxLen {
			return "", fmt.Errorf("invalid header %s: must be %d characters or fewer", metaRequestIdKey, requestIdValidationMaxLen)
		}
		if !requestIdValidationPattern.MatchString(requestId) {
			return "", fmt.Errorf("invalid header %s: must contain only ASCII letters, numbers, dashes, or underscores", metaRequestIdKey)
		}
	}

	return requestId, nil
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
