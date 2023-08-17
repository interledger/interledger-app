import type { CallOptions, ConnectError, Transport } from '@bufbuild/connect'
import { makeAnyClient } from '@bufbuild/connect'
import { createGrpcTransport } from '@bufbuild/connect-node'
import type {
  Message,
  MethodInfo,
  MethodInfoBiDiStreaming,
  MethodInfoClientStreaming,
  MethodInfoServerStreaming,
  MethodInfoUnary,
  PartialMessage,
  ServiceType
} from '@bufbuild/protobuf'
import { MethodKind } from '@bufbuild/protobuf'
import { redirect } from '@remix-run/node'
import { route } from 'routes-gen'
import { BackendService } from '~/generated/connect/backend/v1/backend_connect'
import { BadRequest } from '~/generated/connect/google/rpc/error_details_pb'
import { Code } from '~/generated/protobuf-ts/google/rpc/code'
import type { GrpcError } from '~/lib/proto.server'
import { codeMapping, isGrpcError } from '~/lib/proto.server'

const BACKEND_GRPC_URL = 'http://backend.backend:443'

let grpcConnectClient: PromiseCustomClient<typeof BackendService>

declare global {
  var __grpcConnectClient:
    | PromiseCustomClient<typeof BackendService>
    | undefined
}

const transport = createGrpcTransport({
  baseUrl: BACKEND_GRPC_URL,
  httpVersion: '2'
})

// this is needed because in development we don't want to restart
// the server with every change, but we want to make sure we don't
// create a new connection to the Client with every change either.
if (process.env.NODE_ENV === 'production') {
  grpcConnectClient = createCustomClient(BackendService, transport)
} else {
  if (!global.__grpcConnectClient) {
    global.__grpcConnectClient = createCustomClient(BackendService, transport)
  }
  grpcConnectClient = global.__grpcConnectClient
}

export type PromiseCustomClient<T extends ServiceType> = {
  [P in keyof T['methods']]: T['methods'][P] extends MethodInfoUnary<
    infer I,
    infer O
  >
    ? (
        input: PartialMessage<I>,
        request?: Request,
        options?: CallOptions
      ) => Promise<O>
    : T['methods'][P] extends MethodInfoServerStreaming<infer I, infer O>
    ? (request: PartialMessage<I>, options?: CallOptions) => AsyncIterable<O>
    : T['methods'][P] extends MethodInfoClientStreaming<infer I, infer O>
    ? (
        request: AsyncIterable<PartialMessage<I>>,
        options?: CallOptions
      ) => Promise<O>
    : T['methods'][P] extends MethodInfoBiDiStreaming<infer I, infer O>
    ? (
        request: AsyncIterable<PartialMessage<I>>,
        options?: CallOptions
      ) => AsyncIterable<O>
    : never
}

/**
 * Create a PromiseClient for the given service, invoking RPCs through the
 * given transport.
 */
export function createCustomClient<T extends ServiceType>(
  service: T,
  transport: Transport
) {
  return makeAnyClient(service, (method) => {
    switch (method.kind) {
      case MethodKind.Unary:
        return createUnaryFn(transport, service, method)
      default:
        return null
    }
  }) as PromiseCustomClient<T>
}

/**
 * UnaryFn is the method signature for a unary method of a PromiseClient.
 */
type UnaryFn<I extends Message<I>, O extends Message<O>> = (
  input: PartialMessage<I>,
  request?: Request,
  options?: CallOptions
) => Promise<O | GrpcError>

function createUnaryFn<I extends Message<I>, O extends Message<O>>(
  transport: Transport,
  service: ServiceType,
  method: MethodInfo<I, O>
): UnaryFn<I, O> {
  return async function (input, request?: Request, options?: CallOptions) {
    if (request) {
      const cookies = String(request.headers.get('cookie'))
      options = {
        ...options,
        headers: {
          ...options?.headers,
          cookies
        }
      }
    }

    const response = await transport
      .unary(
        service,
        method,
        options?.signal,
        options?.timeoutMs,
        options?.headers,
        input
      )
      .catch((err) => StatusConnectError(err))

    if (isGrpcError(response)) return response

    options?.onHeader?.(response.header)
    options?.onTrailer?.(response.trailer)

    return response.message
  }
}

function StatusConnectError(err: ConnectError): GrpcError {
  switch (codeMapping(err.code.toString())) {
    case Code.UNAUTHENTICATED:
      throw redirect(route('/login'))
  }

  const details = err.findDetails(BadRequest)

  if (details) {
    return {
      code: err.code,
      message: err.message,
      details
    }
  }

  return {
    code: codeMapping(err.code.toString()),
    message: err.message || 'Status without details.',
    details: []
  }
}

export { grpcConnectClient, StatusConnectError }
