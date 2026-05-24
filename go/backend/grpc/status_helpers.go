package grpc

import (
	"fmt"

	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/protoadapt"
)

// statusFindDetail returns the first detail of the specified type
// (e.g. *errdetails.BadRequest). If not found, it returns the zero value of that type.
func statusFindDetail[T protoadapt.MessageV1](s *status.Status) T {
	for _, detail := range s.Details() {
		if v, ok := detail.(T); ok {
			return v
		}
	}

	var zero T
	return zero
}

// statusWithDetails attaches details to a status, panicking if it fails.
func statusWithDetails(st *status.Status, details ...protoadapt.MessageV1) *status.Status {
	result, err := st.WithDetails(details...)
	if err != nil {
		// If this errored, it will always error
		// here, so better panic so we can figure
		// out why than have this silently passing.
		panic(fmt.Sprintf("Unexpected error attaching details: %v", err))
	}
	return result
}
