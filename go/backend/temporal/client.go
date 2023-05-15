package temporal

import (
	"gitlab.com/fynbos/backend/temporal/context"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/contrib/opentelemetry"
	"go.temporal.io/sdk/interceptor"
	"go.temporal.io/sdk/workflow"
)

func NewTemporalClient(temporalUrl string) (client.Client, error) {
	traceInterceptor, err := opentelemetry.NewTracingInterceptor(opentelemetry.TracerOptions{})
	if err != nil {
		return nil, err
	}

	c, err := client.Dial(client.Options{
		HostPort:           temporalUrl,
		Interceptors:       []interceptor.ClientInterceptor{traceInterceptor},
		ContextPropagators: []workflow.ContextPropagator{context.NewHttpLogContextPropagator()},
	})
	if err != nil {
		return nil, err
	}

	return c, nil
}
