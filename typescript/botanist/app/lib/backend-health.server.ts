import { ChannelCredentials } from '@grpc/grpc-js'
import { GrpcTransport } from '@protobuf-ts/grpc-transport'
import {
  HealthCheckResponse_ServingStatus
} from '~/generated/protobuf-ts/grpc/health/v1/health'
import { HealthClient } from '~/generated/protobuf-ts/grpc/health/v1/health_client'

const BACKEND_GRPC_URL = process.env.BACKEND_GRPC_URL || 'dns:backend-admin:443'
const HEALTH_TIMEOUT_MS = 2000

const transport = new GrpcTransport({
  host: BACKEND_GRPC_URL,
  channelCredentials: ChannelCredentials.createInsecure(),
  timeout: HEALTH_TIMEOUT_MS
})

declare global {
  var __healthClient: HealthClient | undefined
}

function getHealthClient(): HealthClient {
  if (process.env.NODE_ENV === 'production') {
    return new HealthClient(transport)
  }
  if (!global.__healthClient) {
    global.__healthClient = new HealthClient(transport)
  }
  return global.__healthClient
}

export async function getBackendHealth(): Promise<
  { ok: true } | { ok: false; error: string }
> {
  try {
    const client = getHealthClient()
    const response = await client.check({ service: 'backend' })

    if (response.response.status === HealthCheckResponse_ServingStatus.SERVING) {
      return { ok: true }
    }

    return {
      ok: false,
      error: `Backend not serving (status: ${response.response.status})`
    }
  } catch (error) {
    return {
      ok: false,
      error: error instanceof Error ? error.message : String(error)
    }
  }
}
