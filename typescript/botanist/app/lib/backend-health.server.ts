import { Client, credentials, type Metadata } from '@grpc/grpc-js'

const BACKEND_GRPC_URL = process.env.BACKEND_GRPC_URL || 'dns:backend-admin:443'
const HEALTH_TIMEOUT_MS = 2000

// Pre-encoded protobuf: HealthCheckRequest { service: "backend" }
// Field 1 (string): tag=0x0a, length=7, value="backend"
const HEALTH_CHECK_REQUEST = Buffer.from([
  0x0a, 0x07, 0x62, 0x61, 0x63, 0x6b, 0x65, 0x6e, 0x64
])

// SERVING = 1 in HealthCheckResponse.ServingStatus
// Encoded as: field 1 varint, tag=0x08, value=0x01
const SERVING_RESPONSE = Buffer.from([0x08, 0x01])

declare global {
  var __healthClient: Client | undefined
}

function getHealthClient(): Client {
  if (process.env.NODE_ENV === 'production') {
    return new Client(BACKEND_GRPC_URL, credentials.createInsecure())
  }
  if (!global.__healthClient) {
    global.__healthClient = new Client(
      BACKEND_GRPC_URL,
      credentials.createInsecure()
    )
  }
  return global.__healthClient
}

export function getBackendHealth(): Promise<
  { ok: true } | { ok: false; error: string }
> {
  return new Promise((resolve) => {
    const client = getHealthClient()
    const deadline = new Date(Date.now() + HEALTH_TIMEOUT_MS)

    client.makeUnaryRequest(
      '/grpc.health.v1.Health/Check',
      (arg: Buffer) => arg,
      (buf: Buffer) => buf,
      HEALTH_CHECK_REQUEST,
      { deadline },
      (err: Error | null, response?: Buffer) => {
        if (err) {
          resolve({ ok: false, error: err.message })
          return
        }
        if (
          response &&
          response.length === SERVING_RESPONSE.length &&
          response[0] === SERVING_RESPONSE[0] &&
          response[1] === SERVING_RESPONSE[1]
        ) {
          resolve({ ok: true })
          return
        }
        resolve({ ok: false, error: 'Backend not serving' })
      }
    )
  })
}
