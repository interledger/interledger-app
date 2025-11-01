import type { CallOptions, Transport } from '@bufbuild/connect'
import {
  ConnectError as BufConnectError,
  makeAnyClient
} from '@bufbuild/connect'
import { createGrpcTransport } from '@bufbuild/connect-node'
import {
  Message,
  MethodKind,
  type MethodInfo,
  type MethodInfoUnary,
  type PartialMessage,
  type ServiceType
} from '@bufbuild/protobuf'
import { BackendService } from '~/generated/connect/backend/v1/backend_connect'
import { ConnectError } from '~/lib/error.server'

const BACKEND_GRPC_URL = process.env.BACKEND_GRPC_URL || 'http://0.0.0.0:8443'

let grpc: PromiseClient<typeof BackendService>

declare global {
  var __grpcClient: PromiseClient<typeof BackendService> | undefined
}

const transport = createGrpcTransport({
  baseUrl: BACKEND_GRPC_URL,
  httpVersion: '2'
})

// this is needed because in development we don't want to restart
// the server with every change, but we want to make sure we don't
// create a new connection to the Client with every change either.
if (process.env.NODE_ENV === 'production') {
  grpc = createPromiseClient(BackendService, transport)
} else {
  if (!global.__grpcClient) {
    global.__grpcClient = createPromiseClient(BackendService, transport)
  }
  grpc = global.__grpcClient
}

/**
 * PromiseClient is a custom client for connect grpc. This allows us to have better
 * auth and error handling.
 *
 * This is a recommended practice as per the connect docs:
 * https://connectrpc.com/docs/web/using-clients/#roll-your-own-client
 *
 * We have only implemented the unary method. The others can be implemented
 * as needed.
 *
 * The original client code can be found at:
 * https://github.com/connectrpc/connect-es/blob/main/packages/connect/src/promise-client.ts
 */

type PromiseClient<T extends ServiceType> = {
  [P in keyof T['methods']]: T['methods'][P] extends MethodInfoUnary<
    infer I,
    infer O
  >
    ? (
        request: Request,
        input: PartialMessage<I>,
        options?: CallOptions
      ) => Promise<ConnectError | O>
    : never
}

/**
 * Create a PromiseClient for the given service, invoking RPCs through the
 * given transport.
 */
function createPromiseClient<T extends ServiceType>(
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
  }) as PromiseClient<T>
}

/**
 * UnaryFn is the method signature for a unary method of a PromiseClient.
 */
type UnaryFn<I extends Message<I>, O extends Message<O>> = (
  request: Request,
  input: PartialMessage<I>,
  options?: CallOptions
) => Promise<ConnectError | O>

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
    const cookies = request.headers.get('cookie') || ''
    options = {
      ...options,
      headers: {
        ...options?.headers,
        cookies
      }
    }

    return transport
      .unary(
        service,
        method,
        options?.signal,
        options?.timeoutMs,
        options?.headers,
        input
      )
      .then((response) => {
        options?.onHeader?.(response.header)
        options?.onTrailer?.(response.trailer)

        return response.message
      })
      .catch((err) => new ConnectError(request, BufConnectError.from(err)))
  }
}

// Keep a reference to the original protected method
const originalToJson = Message.prototype.toJson

// Monkeypatch to emit enums as integers. We need this so we can actually use
// the generated protobuf enums. Initially, when returning data structures
// from a loader or action that reference a protobuf generate typed, they were
// getting stringified with the enum key.
;(Message.prototype as any).toJSON = function () {
  return originalToJson.call(this, {
    enumAsInteger: true,
    emitDefaultValues: true
  })
}

export { grpc }
