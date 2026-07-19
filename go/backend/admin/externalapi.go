package admin

import (
	"context"
	"fmt"

	httplog "github.com/interledger/interledger-app/go/backend/providers/http"
	pb "github.com/interledger/interledger-app/go/proto/backend/admin/v1"
)

func (s *AdminRpcService) ListExternalApiCalls(
	ctx context.Context,
	req *pb.ListExternalApiCallsRequest,
) (*pb.ListExternalApiCallsResponse, error) {
	var logs []httplog.LogRecord
	err := s.b.DB().SelectContext(ctx, &logs, fmt.Sprintf("SELECT %s FROM external_api_logs WHERE context=$1 limit 100;", httplog.Fields), "paymentID="+req.GetPaymentId())
	if err != nil {
		return nil, toGRPCError(err)
	}

	ret := make([]*pb.ExternalApiCall, len(logs))
	for i, log := range logs {
		ret[i] = &pb.ExternalApiCall{
			Id:             log.ID,
			Provider:       log.Provider,
			Context:        log.Context,
			Method:         log.Method,
			RequestBody:    log.RequestBody,
			RequestPath:    log.RequestPath,
			ResponseBody:   log.ResponseBody,
			ResponseStatus: log.ResponseStatus,
		}
	}

	return &pb.ListExternalApiCallsResponse{
		List: ret,
	}, nil
}
