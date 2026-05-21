package grpc

import (
	"google.golang.org/grpc/metadata"
)

const metaRequestIDKey = "X-Request-ID"

// RequestIDFromMetadata obtains the value of the headerRequestId key from the metadata
func RequestIDFromMetadata(meta metadata.MD) (string, error) {
	if meta == nil {
		return "", nil
	}

	requestID := ""
	headerRequestID := meta.Get(metaRequestIDKey)
	if len(headerRequestID) > 0 {
		requestID = headerRequestID[0]
	}

	return requestID, nil
}
