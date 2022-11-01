package temporal

import (
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/contrib/opentelemetry"
	"go.temporal.io/sdk/interceptor"
)

func NewTemporalClient(temporalUrl string) (client.Client, error) {
	traceInterceptor, err := opentelemetry.NewTracingInterceptor(opentelemetry.TracerOptions{})
	if err != nil {
		return nil, err
	}

	c, err := client.Dial(client.Options{
		HostPort:     temporalUrl,
		Interceptors: []interceptor.ClientInterceptor{traceInterceptor},
	})
	if err != nil {
		return nil, err
	}

	return c, nil
}
