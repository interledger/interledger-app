package temporal

import (
	"context"
	"errors"
	"time"

	temporal_context "github.com/interledger/interledger-app/go/backend/temporal/context"
	"github.com/interledger/interledger-app/go/log"
	"go.temporal.io/api/serviceerror"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/contrib/opentelemetry"
	"go.temporal.io/sdk/interceptor"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/durationpb"
)

func NewTemporalClient(temporalUrl string) (client.Client, error) {
	traceInterceptor, err := opentelemetry.NewTracingInterceptor(opentelemetry.TracerOptions{})
	if err != nil {
		return nil, err
	}

	c, err := client.Dial(client.Options{
		HostPort:           temporalUrl,
		Logger:             newTemporalLogger(),
		Interceptors:       []interceptor.ClientInterceptor{traceInterceptor},
		ContextPropagators: []workflow.ContextPropagator{temporal_context.NewHttpLogContextPropagator()},
	})
	if err != nil {
		return nil, err
	}

	nc, err := client.NewNamespaceClient(client.Options{
		HostPort: temporalUrl,
		Logger:   newTemporalLogger(),
	})
	if err != nil {
		log.Error("Failed to create temporal namespace client", zap.Error(err))
		return c, nil
	}

	var existsError *serviceerror.NamespaceAlreadyExists

	var defaultRetentionPeriod durationpb.Duration = *durationpb.New(3 * 24 * time.Hour)
	err = nc.Register(context.Background(), &workflowservice.RegisterNamespaceRequest{
		Namespace:                        "default",
		WorkflowExecutionRetentionPeriod: &defaultRetentionPeriod,
	})
	if err != nil && !errors.As(err, &existsError) {
		log.Error("Failed to register temporal default namespace.", zap.Error(err))
	}

	var paymentsRetentionPeriod durationpb.Duration = *durationpb.New(10 * 365 * 24 * time.Hour)

	err = nc.Register(context.Background(), &workflowservice.RegisterNamespaceRequest{
		Namespace:                        "payments",
		WorkflowExecutionRetentionPeriod: &paymentsRetentionPeriod,
	})
	if err != nil && !errors.As(err, &existsError) {
		log.Error("Failed to register temporal payments namespace.", zap.Error(err))
	}

	return c, nil
}
