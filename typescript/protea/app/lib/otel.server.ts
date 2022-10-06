import { NodeSDK } from '@opentelemetry/sdk-node'
// import { getNodeAutoInstrumentations } from '@opentelemetry/auto-instrumentations-node'
import { OTLPTraceExporter } from '@opentelemetry/exporter-trace-otlp-proto'
import { RemixInstrumentation } from 'opentelemetry-instrumentation-remix'
import { GrpcInstrumentation } from '@opentelemetry/instrumentation-grpc'

// The Trace Exporter exports the data to Honeycomb and uses
// the environment variables for endpoint, service name, and API Key.
const traceExporter = new OTLPTraceExporter()

export const sdk = new NodeSDK({
  traceExporter,
  instrumentations: [
    // getNodeAutoInstrumentations(),
    new RemixInstrumentation(),
    new GrpcInstrumentation()
  ]
})
