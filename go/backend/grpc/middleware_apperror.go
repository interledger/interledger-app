package grpc

import (
	"context"
	"fmt"

	"gitlab.com/fynbos/log"
	pb "gitlab.com/fynbos/proto/backend/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/protoadapt"
)

// MakeUnaryInterceptorAppError ensures that an AppError detail is present
// in every GRPC error response
func MakeUnaryInterceptorAppError() grpc.ServerOption {
	return grpc.ChainUnaryInterceptor(func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		result, err := handler(ctx, req)
		if err != nil {
			err = withAppError(ctx, err)
		}

		return result, err
	})
}

// withAppError ensures every gRPC error response carries exactly one AppError
// detail and augments it with the request id from the context.
//
// It handles two cases:
//   - A raw (non-gRPC) error: wrapped as codes.Internal with a new AppError.
//   - An existing gRPC status: the AppError detail is extracted and augmented;
//     if no AppError was present, a new one is created.
func withAppError(ctx context.Context, originalErr error) error {
	reqId := RequestIdFromContext(ctx)

	st, ok := status.FromError(originalErr)
	if !ok {
		// This is a raw error that wasn't handled by toGRPCError().
		// Most likely it is an internal error (e.g. database error)
		st = status.New(codes.Internal, originalErr.Error())
	}

	// Separate the AppError (if it exists) from the rest of the details
	var appError *pb.AppError
	var details []protoadapt.MessageV1

	for _, detail := range st.Details() {
		if existingAppError, ok := detail.(*pb.AppError); ok {
			if appError == nil {
				appError = existingAppError
			} else {
				log.Error("Found multiple AppError details")
			}
		} else if msg, ok := detail.(protoadapt.MessageV1); ok {
			details = append(details, msg)
		} else {
			log.Error("Trying to process a detail of an unknown type: " + fmt.Sprintf("%T", detail))
		}
	}

	// Create the AppError if it didn't exist.
	if appError == nil {
		appError = &pb.AppError{
			ErrorCode: pb.ErrorCode_ERROR_CODE_INTERNAL,
			Message:   st.Message(),
		}
	}

	if appError.ReqId != "" && appError.ReqId != reqId {
		panic("appError.ReqId is already present and different from the one in the context.")
	}

	// Augment the AppError with the request id
	appError.ReqId = reqId

	// Reconstruct the details
	details = append(details, appError)

	newSt := statusWithDetails(status.New(st.Code(), st.Message()), details...)

	return newSt.Err()
}
