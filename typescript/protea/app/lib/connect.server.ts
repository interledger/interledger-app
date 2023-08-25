import type { CallOptions, Transport } from '@bufbuild/connect'
import { ConnectError, makeAnyClient } from '@bufbuild/connect'
import { createGrpcTransport } from '@bufbuild/connect-node'
import type {
  Message,
  MethodInfo,
  MethodInfoUnary,
  PartialMessage,
  ServiceType
} from '@bufbuild/protobuf'
import { MethodKind } from '@bufbuild/protobuf'
import { BackendService } from '~/generated/connect/backend/v1/backend_connect'
import type { Result } from './result.server'
import * as R from './result.server'

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
        request: Request,
        input: PartialMessage<I>,
        options?: CallOptions
      ) => Promise<Result<ConnectError, O>>
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
  request: Request,
  input: PartialMessage<I>,
  options?: CallOptions
) => Promise<Result<ConnectError, O>>

function createUnaryFn<I extends Message<I>, O extends Message<O>>(
  transport: Transport,
  service: ServiceType,
  method: MethodInfo<I, O>
): UnaryFn<I, O> {
  return async function (
    request: Request,
    input: PartialMessage<I>,
    options?: CallOptions
  ) {
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

    return await transport
      .unary(
        service,
        method,
        options?.signal,
        options?.timeoutMs,
        options?.headers,
        input
      )
      .then((res) => {
        options?.onHeader?.(res.header)
        options?.onTrailer?.(res.trailer)

        return R.ok(res.message)
      })
      .catch((err) => {
        return R.err(intoConnectError(err))
      })
  }
}

function intoConnectError(err: unknown): CustomConnectError {
  return ConnectError.from(err)
}

class CustomConnectError extends ConnectError {}

export { grpcConnectClient, intoConnectError }
