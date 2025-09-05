package context

import (
	"context"

	httplog "gitlab.com/fynbos/backend/providers/http"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/workflow"
)

// propagator implements the custom context propagator
type propagator struct{}

// TODO: Change?

// propagationKey is the key used by the propagator to pass values through the
// Temporal server headers
const propagationKey = "fynbos-http-log"

// NewHttpLogContextPropagator returns a context propagator that propagates a set of
// string key-value pairs across a workflow
func NewHttpLogContextPropagator() workflow.ContextPropagator {
	return &propagator{}
}

// Inject injects values from context into headers for propagation
func (s *propagator) Inject(ctx context.Context, writer workflow.HeaderWriter) error {
	value := ctx.Value(httplog.ContextKey)
	payload, err := converter.GetDefaultDataConverter().ToPayload(value)
	if err != nil {
		return err
	}
	writer.Set(propagationKey, payload)
	return nil
}

// InjectFromWorkflow injects values from context into headers for propagation
func (s *propagator) InjectFromWorkflow(ctx workflow.Context, writer workflow.HeaderWriter) error {
	value := ctx.Value(httplog.ContextKey)
	payload, err := converter.GetDefaultDataConverter().ToPayload(value)
	if err != nil {
		return err
	}
	writer.Set(propagationKey, payload)
	return nil
}

// Extract extracts values from headers and puts them into context
func (s *propagator) Extract(ctx context.Context, reader workflow.HeaderReader) (context.Context, error) {
	if value, ok := reader.Get(propagationKey); ok {
		var values httplog.Metadata
		if err := converter.GetDefaultDataConverter().FromPayload(value, &values); err != nil {
			return ctx, nil
		}
		ctx = context.WithValue(ctx, httplog.ContextKey, &values)
	}

	return ctx, nil
}

// ExtractToWorkflow extracts values from headers and puts them into context
func (s *propagator) ExtractToWorkflow(ctx workflow.Context, reader workflow.HeaderReader) (workflow.Context, error) {
	if value, ok := reader.Get(propagationKey); ok {
		var values httplog.Metadata
		if err := converter.GetDefaultDataConverter().FromPayload(value, &values); err != nil {
			return ctx, nil
		}
		ctx = workflow.WithValue(ctx, httplog.ContextKey, values)
	}

	return ctx, nil
}
