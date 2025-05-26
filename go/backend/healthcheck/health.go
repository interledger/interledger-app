package healthcheck

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
)

// This implements the grpc health-checking protocol
// https://github.com/grpc/grpc/blob/master/doc/health-checking.md
// At the moment this is a naive implementation that does not check
// if any of the other services are ready.

type Service interface {
	Check(ctx context.Context, req *grpc_health_v1.HealthCheckRequest) (*grpc_health_v1.HealthCheckResponse, error)
	Watch(req *grpc_health_v1.HealthCheckRequest, server grpc_health_v1.Health_WatchServer) error
	List(ctx context.Context, req *grpc_health_v1.HealthListRequest) (*grpc_health_v1.HealthListResponse, error)
}

func NewService() (Service, error) {
	return &service{}, nil
}

type service struct{}

func (s *service) Check(ctx context.Context, req *grpc_health_v1.HealthCheckRequest) (*grpc_health_v1.HealthCheckResponse, error) {
	svc := req.GetService()
	if svc != "backend" {
		return nil, status.Error(codes.Unknown, "Service unknown.")
	}

	return &grpc_health_v1.HealthCheckResponse{
		Status: grpc_health_v1.HealthCheckResponse_SERVING,
	}, nil
}

func (s *service) Watch(req *grpc_health_v1.HealthCheckRequest, server grpc_health_v1.Health_WatchServer) error {
	svc := req.GetService()
	if svc != "backend" {
		return status.Error(codes.Unknown, "Service unknown.")
	}

	return server.Send(&grpc_health_v1.HealthCheckResponse{
		Status: grpc_health_v1.HealthCheckResponse_SERVING,
	})
}

func (s *service) List(ctx context.Context, req *grpc_health_v1.HealthListRequest) (*grpc_health_v1.HealthListResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method List not implemented")
}
