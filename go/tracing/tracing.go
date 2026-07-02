package tracing

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/interledger/interledger-app/go/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"go.uber.org/zap"
)

// InitTraceProvider configures the global OTEL tracer provider.
//
// When enabled is false the provider uses a no-op exporter, so tracing can be
// switched off explicitly regardless of environment. When enabled, spans are
// exported over OTLP/gRPC using endpoint and headers from configuration.
//
// The standard OTEL_EXPORTER_OTLP_* environment variables take priority over
// the configured endpoint/headers: when such a var is set it is used instead of
// the config value, and a warning is logged so the override is visible.
func InitTraceProvider(serviceName, version string, enabled bool, endpoint string, headers map[string]string) (func(context.Context) error, error) {
	ctx := context.Background()

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
			semconv.ServiceVersionKey.String(version),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	var traceExporter sdktrace.SpanExporter

	if enabled {
		client := otlptracegrpc.NewClient(exporterOptions(endpoint, headers)...)
		traceExporter, err = otlptrace.New(ctx, client)
		if err != nil {
			return nil, fmt.Errorf("failed to create trace exporter: %w", err)
		}
	} else {
		traceExporter = tracetest.NewNoopExporter()
	}

	// Register the trace exporter with a TracerProvider, using a batch
	// span processor to aggregate spans before export.
	bsp := sdktrace.NewBatchSpanProcessor(traceExporter)
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)
	otel.SetTracerProvider(tracerProvider)

	// set global propagator to trace context (the default is no-op).
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	// Shutdown will flush any remaining spans and shut down the exporter.
	return tracerProvider.Shutdown, nil
}

// exporterOptions translates the configured endpoint and headers into gRPC
// exporter options.
//
// The standard OTEL_EXPORTER_OTLP_* environment variables take priority: for
// each setting, if the corresponding env var is set the option is omitted so
// the SDK reads the env var itself, and a warning is logged when this overrides
// a configured value. Settings with no env var and no config value yield no
// option (the SDK's own defaults apply).
func exporterOptions(endpoint string, headers map[string]string) []otlptracegrpc.Option {
	var opts []otlptracegrpc.Option

	if endpointEnvSet() {
		if endpoint != "" {
			log.Warn("OTEL endpoint environment variable is set; it overrides the configured otel.endpoint",
				zap.String("configured_endpoint", endpoint))
		}
	} else if endpoint != "" {
		opts = append(opts, endpointOptions(endpoint)...)
	}

	if headersEnvSet() {
		if len(headers) > 0 {
			log.Warn("OTEL headers environment variable is set; it overrides the configured otel.headers")
		}
	} else if len(headers) > 0 {
		opts = append(opts, otlptracegrpc.WithHeaders(headers))
	}

	return opts
}

// endpointOptions parses the endpoint URL into exporter options. Scheme handling
// matches the SDK's env-var parsing (withEndpointScheme): an http or unix scheme
// selects an insecure connection; anything else (including the grpc scheme used
// for Honeycomb) is secure/TLS. A URL that fails to parse is logged and skipped
// so a bad config value cannot take down startup.
func endpointOptions(endpoint string) []otlptracegrpc.Option {
	u, err := url.Parse(endpoint)
	if err != nil {
		log.Warn("failed to parse configured otel.endpoint; ignoring",
			zap.String("endpoint", endpoint), zap.Error(err))
		return nil
	}
	opts := []otlptracegrpc.Option{otlptracegrpc.WithEndpoint(path.Join(u.Host, u.Path))}
	switch strings.ToLower(u.Scheme) {
	case "http", "unix":
		opts = append(opts, otlptracegrpc.WithInsecure())
	}
	return opts
}

// endpointEnvSet reports whether an OTLP endpoint is configured via the standard
// environment variables the SDK consults (traces-specific var wins over the
// generic one, but either presence means the env config takes priority).
func endpointEnvSet() bool {
	return os.Getenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT") != "" ||
		os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT") != ""
}

// headersEnvSet reports whether OTLP headers are configured via the standard
// environment variables the SDK consults.
func headersEnvSet() bool {
	return os.Getenv("OTEL_EXPORTER_OTLP_TRACES_HEADERS") != "" ||
		os.Getenv("OTEL_EXPORTER_OTLP_HEADERS") != ""
}
