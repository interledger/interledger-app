import type {
  BinaryReadOptions,
  FieldList,
  JsonReadOptions,
  JsonValue,
  PartialMessage,
  PlainMessage
} from '@bufbuild/protobuf'
import { Message, MethodKind, proto3 } from '@bufbuild/protobuf'
import { createGrpcTransport } from '@bufbuild/connect-node'
import logger from './logger.server'

const BACKEND_GRPC_URL = process.env.BACKEND_GRPC_URL || 'http://0.0.0.0:8443'

// -- Inline gRPC health proto types (grpc.health.v1) --

enum ServingStatus {
  UNKNOWN = 0,
  SERVING = 1,
  NOT_SERVING = 2,
  SERVICE_UNKNOWN = 3
}
proto3.util.setEnumType(
  ServingStatus,
  'grpc.health.v1.HealthCheckResponse.ServingStatus',
  [
    { no: 0, name: 'UNKNOWN' },
    { no: 1, name: 'SERVING' },
    { no: 2, name: 'NOT_SERVING' },
    { no: 3, name: 'SERVICE_UNKNOWN' }
  ]
)

class HealthCheckRequest extends Message<HealthCheckRequest> {
  service = ''
  constructor(data?: PartialMessage<HealthCheckRequest>) {
    super()
    proto3.util.initPartial(data, this)
  }
  static readonly runtime = proto3
  static readonly typeName = 'grpc.health.v1.HealthCheckRequest'
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: 'service', kind: 'scalar' as const, T: 9 /* STRING */ }
  ])
  static fromBinary(
    bytes: Uint8Array,
    options?: Partial<BinaryReadOptions>
  ): HealthCheckRequest {
    return new HealthCheckRequest().fromBinary(bytes, options)
  }
  static fromJson(
    jsonValue: JsonValue,
    options?: Partial<JsonReadOptions>
  ): HealthCheckRequest {
    return new HealthCheckRequest().fromJson(jsonValue, options)
  }
  static fromJsonString(
    jsonString: string,
    options?: Partial<JsonReadOptions>
  ): HealthCheckRequest {
    return new HealthCheckRequest().fromJsonString(jsonString, options)
  }
  static equals(
    a:
      | HealthCheckRequest
      | PlainMessage<HealthCheckRequest>
      | undefined,
    b:
      | HealthCheckRequest
      | PlainMessage<HealthCheckRequest>
      | undefined
  ): boolean {
    return proto3.util.equals(HealthCheckRequest, a, b)
  }
}

class HealthCheckResponse extends Message<HealthCheckResponse> {
  status: ServingStatus = ServingStatus.UNKNOWN
  constructor(data?: PartialMessage<HealthCheckResponse>) {
    super()
    proto3.util.initPartial(data, this)
  }
  static readonly runtime = proto3
  static readonly typeName = 'grpc.health.v1.HealthCheckResponse'
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    {
      no: 1,
      name: 'status',
      kind: 'enum' as const,
      T: proto3.getEnumType(ServingStatus)
    }
  ])
  static fromBinary(
    bytes: Uint8Array,
    options?: Partial<BinaryReadOptions>
  ): HealthCheckResponse {
    return new HealthCheckResponse().fromBinary(bytes, options)
  }
  static fromJson(
    jsonValue: JsonValue,
    options?: Partial<JsonReadOptions>
  ): HealthCheckResponse {
    return new HealthCheckResponse().fromJson(jsonValue, options)
  }
  static fromJsonString(
    jsonString: string,
    options?: Partial<JsonReadOptions>
  ): HealthCheckResponse {
    return new HealthCheckResponse().fromJsonString(jsonString, options)
  }
  static equals(
    a:
      | HealthCheckResponse
      | PlainMessage<HealthCheckResponse>
      | undefined,
    b:
      | HealthCheckResponse
      | PlainMessage<HealthCheckResponse>
      | undefined
  ): boolean {
    return proto3.util.equals(HealthCheckResponse, a, b)
  }
}

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

// -- Health check transport (reuses BACKEND_GRPC_URL) --

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
      undefined, // signal
      2000, // timeoutMs
      undefined, // headers
      { service: 'backend' }
    )
    if (response.message.status === ServingStatus.SERVING) {
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
