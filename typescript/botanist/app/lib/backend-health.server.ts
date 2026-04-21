import * as grpc from '@grpc/grpc-js'

function requireEnv(name: string): string {
  const value = process.env[name]
  if (!value) {
    throw new Error(`${name} environment variable is required`)
  }
  return value
}

const BACKEND_GRPC_URL = requireEnv('BACKEND_GRPC_URL')
const HEALTH_TIMEOUT_MS = 2000
const SERVING_STATUS = 1

function readVarint(buf: Buffer, offset: number): [number, number] {
  let result = 0
  let shift = 0
  let i = offset
  while (i < buf.length) {
    const byte = buf[i++]
    result |= (byte & 0x7f) << shift
    if ((byte & 0x80) === 0) {
      return [result, i]
    }
    shift += 7
  }
  return [0, i]
}

export function getServingStatus(response: Buffer): number {
  let i = 0
  while (i < response.length) {
    const byte = response[i++]
    const fieldNumber = byte >> 3
    const wireType = byte & 0x07

    if (fieldNumber === 1 && wireType === 0) {
      const [val] = readVarint(response, i)
      return val
    }

    if (wireType === 0) {
      const [, next] = readVarint(response, i)
      i = next
    } else if (wireType === 2) {
      const [len, next] = readVarint(response, i)
      i = next + len
    } else if (wireType === 1) {
      i += 8
    } else if (wireType === 5) {
      i += 4
    }
  }
  return 0
}

declare global {
  var __healthClient: grpc.Client | undefined
}

function getHealthClient(): grpc.Client {
  if (process.env.NODE_ENV === 'production') {
    return new grpc.Client(BACKEND_GRPC_URL, grpc.credentials.createInsecure())
  }
  if (!global.__healthClient) {
    global.__healthClient = new grpc.Client(
      BACKEND_GRPC_URL,
      grpc.credentials.createInsecure()
    )
  }
  return global.__healthClient
}

export async function getBackendHealth(): Promise<
  { ok: true } | { ok: false; error: string }
> {
  return new Promise((resolve) => {
    const deadline = new Date(Date.now() + HEALTH_TIMEOUT_MS)

    const requestBuffer = Buffer.from([0x0a, 0x07, ...Buffer.from('backend')])

    const client = getHealthClient()
    client.makeUnaryRequest(
      '/grpc.health.v1.Health/Check',
      (arg: unknown) => arg as Buffer,
      (arg: unknown) => arg as Buffer,
      requestBuffer,
      new grpc.Metadata(),
      { deadline },
      (error, response: Buffer | undefined) => {
        if (error) {
          resolve({ ok: false, error: error.message })
          return
        }
        if (!response) {
          resolve({ ok: false, error: 'Empty response' })
          return
        }

        const status = getServingStatus(response)
        if (status === SERVING_STATUS) {
          resolve({ ok: true })
        } else {
          resolve({
            ok: false,
            error: `Backend not serving (status: ${status})`
          })
        }
      }
    )
  })
}
