import { Client, credentials } from '@grpc/grpc-js'

const BACKEND_GRPC_URL = process.env.BACKEND_GRPC_URL || 'dns:backend-admin:443'
const HEALTH_TIMEOUT_MS = 2000

// Pre-encoded protobuf: HealthCheckRequest { service: "backend" }
// Field 1 (string): tag=0x0a, length=7, value="backend"
const HEALTH_CHECK_REQUEST = Buffer.from([
  0x0a, 0x07, 0x62, 0x61, 0x63, 0x6b, 0x65, 0x6e, 0x64
])

const SERVING_STATUS = 1

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

function readVarint(
  buf: Buffer,
  start: number
): { value: number; next: number } | null {
  let value = 0
  let shift = 0
  let index = start

  while (index < buf.length && shift < 35) {
    const byte = buf[index]
    value |= (byte & 0x7f) << shift
    index += 1

    if ((byte & 0x80) === 0) {
      return { value, next: index }
    }

    shift += 7
  }

  return null
}

// Extract field 1 (status) from grpc.health.v1.HealthCheckResponse wire bytes.
function getServingStatus(response: Buffer): number | null {
  let index = 0

  while (index < response.length) {
    const tag = readVarint(response, index)
    if (!tag) {
      return null
    }

    index = tag.next
    const fieldNumber = tag.value >>> 3
    const wireType = tag.value & 0x07

    if (fieldNumber === 1 && wireType === 0) {
      const status = readVarint(response, index)
      return status ? status.value : null
    }

    if (wireType === 0) {
      const value = readVarint(response, index)
      if (!value) {
        return null
      }
      index = value.next
      continue
    }

    if (wireType === 2) {
      const length = readVarint(response, index)
      if (!length) {
        return null
      }
      index = length.next + length.value
      if (index > response.length) {
        return null
      }
      continue
    }

    return null
  }

  return null
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
        if (response && getServingStatus(response) === SERVING_STATUS) {
          resolve({ ok: true })
          return
        }
        resolve({ ok: false, error: 'Backend not serving' })
      }
    )
  })
}
