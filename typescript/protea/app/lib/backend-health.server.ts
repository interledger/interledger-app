import { MethodKind, proto3 } from '@bufbuild/protobuf'
import { createGrpcTransport } from '@bufbuild/connect-node'
import logger from './logger.server'

function requireEnv(name: string): string {
  const value = process.env[name]
  if (!value) {
    throw new Error(`${name} environment variable is required`)
  }
  return value
}

const BACKEND_GRPC_URL = requireEnv('BACKEND_GRPC_URL')
const SERVING_STATUS = 1

const HealthCheckRequest = proto3.makeMessageType('grpc.health.v1.HealthCheckRequest', [
  { no: 1, name: 'service', kind: 'scalar', T: 9 /* ScalarType.STRING */ }
])

const HealthCheckResponse = proto3.makeMessageType('grpc.health.v1.HealthCheckResponse', [
  { no: 1, name: 'status', kind: 'scalar', T: 5 /* ScalarType.INT32 */ }
])

const HealthService = {
  typeName: 'grpc.health.v1.Health',
  methods: {
    check: {
      name: 'Check',
      I: HealthCheckRequest,
      O: HealthCheckResponse,
      kind: MethodKind.Unary
    }
  }
} as const

const transport = createGrpcTransport({
  baseUrl: BACKEND_GRPC_URL,
  httpVersion: '2'
})

export async function getBackendHealth(): Promise<
  { ok: true } | { ok: false; error: string }
> {
  try {
    const response = await transport.unary(
      HealthService,
      HealthService.methods.check,
      undefined,
      2000,
      undefined,
      { service: 'backend' }
    )
    if (response.message.status === SERVING_STATUS) {
      return { ok: true }
    }
    return {
      ok: false,
      error: `Backend not serving (status: ${response.message.status})`
    }
  } catch (err) {
    const message = err instanceof Error ? err.message : String(err)
    logger.debug({ error: message }, 'Backend health check failed')
    return { ok: false, error: message }
  }
}
