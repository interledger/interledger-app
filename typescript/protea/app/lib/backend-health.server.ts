import { createPromiseClient } from '@bufbuild/connect'
import { createGrpcTransport } from '@bufbuild/connect-node'
import { envValue } from '~/env.server'
import { Health } from '~/generated/connect/grpc/health/v1/health_connect'
import { HealthCheckResponse_ServingStatus } from '~/generated/connect/grpc/health/v1/health_pb'
import logger from './logger.server'

const BACKEND_GRPC_URL = envValue('BACKEND_GRPC_URL')

const transport = createGrpcTransport({
  baseUrl: BACKEND_GRPC_URL,
  httpVersion: '2'
})

const healthClient = createPromiseClient(Health, transport)

export async function getBackendHealth(): Promise<
  { ok: true } | { ok: false; error: string }
> {
  try {
    const response = await healthClient.check(
      { service: 'backend' },
      { timeoutMs: 2000 }
    )
    if (response.status === HealthCheckResponse_ServingStatus.SERVING) {
      return { ok: true }
    }
    return {
      ok: false,
      error: `Backend not serving (status: ${response.status})`
    }
  } catch (err) {
    const message = err instanceof Error ? err.message : String(err)
    logger.debug({ error: message }, 'Backend health check failed')
    return { ok: false, error: message }
  }
}
