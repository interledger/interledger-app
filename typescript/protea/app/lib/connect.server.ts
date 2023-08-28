import type { CallOptions, Transport } from '@bufbuild/connect'
import { Code, ConnectError, makeAnyClient } from '@bufbuild/connect'
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
import { BadRequest } from '~/generated/connect/google/rpc/error_details_pb'
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
      ) => Promise<Result<ExtConnectError<I>, O>>
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
) => Promise<Result<ExtConnectError<I>, O>>

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
        return R.err(ExtConnectError.from<I>(err))
      })
  }
}

class ExtConnectError<I extends Message<I>> extends ConnectError {
  static from<T extends Message<T>>(
    reason: unknown,
    code = Code.Unknown
  ): ExtConnectError<T> {
    if (reason instanceof ConnectError) {
      // @ts-ignore
      const details: Message[] = reason.details

      return new ExtConnectError(
        reason.message,
        reason.code,
        reason.metadata,
        details,
        reason.cause
      )
    }
    if (reason instanceof Error) {
      if (reason.name == 'AbortError') {
        return new ExtConnectError<T>(reason.message, Code.Canceled)
      }
      return new ExtConnectError<T>(
        reason.message,
        code,
        undefined,
        undefined,
        reason
      )
    }
    return new ExtConnectError<T>(
      String(reason),
      code,
      undefined,
      undefined,
      reason
    )
  }

  findViolations<E>(
    input: PartialMessage<I>,
    optFieldErrors?: E
  ): (E extends undefined ? PartialMessage<I> : E) {
    const violations = this.findDetails(BadRequest)[0].fieldViolations
    const fieldNames: { [p: string]: string } = {}
    const fieldErrors: PartialMessage<I> = {}

    Object.entries(input).forEach(([key, value]) =>
      Object.assign(fieldNames, {
        [key.charAt(0).toUpperCase() + key.slice(1)]: String(value)
      })
    ) // TODO: handle nested fields if needed

    violations?.forEach((violation) => {
      const fieldName = fieldNames[violation.field]
      if (!fieldName) {
        throw new Error(
          `Field ${violation.field} not found in field names: ${JSON.stringify(
            fieldNames
          )}`
        )
      }
      return Object.assign(fieldErrors, { [fieldName]: violation.description })
    })

    if (optFieldErrors) {
      Object.assign(fieldErrors, optFieldErrors)
    }

    return fieldErrors as E extends undefined ? PartialMessage<I> : E
  }
}

export { grpcConnectClient }
