import { ChannelCredentials } from '@grpc/grpc-js'
import { GrpcTransport } from '@protobuf-ts/grpc-transport'
import { HealthClient } from '~/generated/protobuf-ts/grpc/health/v1/health_client'
import { HealthCheckResponse_ServingStatus } from '~/generated/protobuf-ts/grpc/health/v1/health'

function requireEnv(name: string): string {
  const value = process.env[name]
  if (!value) {
    throw new Error(`${name} environment variable is required`)
  }
  return value
}

const BACKEND_GRPC_URL = requireEnv('BACKEND_GRPC_URL')
const HEALTH_TIMEOUT_MS = 2000

declare global {
  var __healthClient: HealthClient | undefined
}

function getHealthClient(): HealthClient {
  if (process.env.NODE_ENV === 'production') {
    const transport = new GrpcTransport({
      host: BACKEND_GRPC_URL,
      channelCredentials: ChannelCredentials.createInsecure()
    })
    return new HealthClient(transport)
  }
  if (!global.__healthClient) {
    const transport = new GrpcTransport({
      host: BACKEND_GRPC_URL,
      channelCredentials: ChannelCredentials.createInsecure()
    })
    global.__healthClient = new HealthClient(transport)
  }
  return global.__healthClient
}

export async function getBackendHealth(): Promise<
  { ok: true } | { ok: false; error: string }
> {
  try {
    const client = getHealthClient()
    const { response } = await client.check(
      { service: 'backend' },
      { timeout: HEALTH_TIMEOUT_MS }
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
    return { ok: false, error: message }
  }
}
