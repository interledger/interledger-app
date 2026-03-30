import { ChannelCredentials } from '@grpc/grpc-js'
import { GrpcTransport } from '@protobuf-ts/grpc-transport'
import { base64decode } from '@protobuf-ts/runtime'
import type { RpcError } from '@protobuf-ts/runtime-rpc'
import { BackendClient } from '~/generated/protobuf-ts/backend/admin/v1/backend_client'
import { Any } from '~/generated/protobuf-ts/google/protobuf/any'
import { Code } from '~/generated/protobuf-ts/google/rpc/code'
import {
  BadRequest,
  DebugInfo,
  ErrorInfo,
  Help,
  LocalizedMessage,
  PreconditionFailure,
  QuotaFailure,
  RequestInfo,
  ResourceInfo,
  RetryInfo
} from '~/generated/protobuf-ts/google/rpc/error_details'
import { Status } from '~/generated/protobuf-ts/google/rpc/status'

const BACKEND_GRPC_URL = process.env.BACKEND_GRPC_URL || 'dns:backend-admin:443'

const transport = new GrpcTransport({
  host: BACKEND_GRPC_URL,
  channelCredentials: ChannelCredentials.createInsecure(),
  timeout: 100000
})

let grpcClient: BackendClient

declare global {
  var __grpcClient: BackendClient | undefined
}

// this is needed because in development we don't want to restart
// the server with every change, but we want to make sure we don't
// create a new connection to the Client with every change either.
if (process.env.NODE_ENV === 'production') {
  grpcClient = new BackendClient(transport)
} else {
  if (!global.__grpcClient) {
    global.__grpcClient = new BackendClient(transport)
  }
  grpcClient = global.__grpcClient
}

interface GrpcError extends Status {
  code: number
  message: string
  details: any[]
}

function isGrpcError(res: any): res is GrpcError {
  return (
    (res as GrpcError).code !== undefined &&
    (res as GrpcError).message !== undefined &&
    (res as GrpcError).details !== undefined
  )
}

/**
 * This function is intended to deserialise and parse the RpcError
 * returned by grpcClient functions. It will return the error message
 * in the same shape of successful rpc calls, with the error message in the response.
 * @param err The error received from a grpc call.
 * @returns GrpcError - the error code and details
 */
function StatusError(err: RpcError): GrpcError {
  let status: Status | undefined
  let details: any[] | undefined
  if (!err.meta) {
    // return null
    throw new Error('No meta on error')
  }

  if (err.meta['grpc-status-details-bin']) {
    const buffer = base64decode(err.meta['grpc-status-details-bin'] as string)

    if (!buffer || typeof buffer === 'string') {
      // return null
    }

    status = Status.fromBinary(buffer)
    details = status.details
      .map((detail) => {
        return Any.unpack(detail, typeRegistry[detail.typeUrl]) || null
      })
      .filter(Boolean)

    return {
      code: status.code,
      message: status.message,
      details
    }
  }

  return {
    code: codeMapping(err.code),
    message: 'Status without details',
    details: []
  }
}

const typeRegistry: Record<string, any> = {
  'type.googleapis.com/google.rpc.RetryInfo': RetryInfo,
  'type.googleapis.com/google.rpc.DebugInfo': DebugInfo,
  'type.googleapis.com/google.rpc.QuotaFailure': QuotaFailure,
  'type.googleapis.com/google.rpc.PreconditionFailure': PreconditionFailure,
  'type.googleapis.com/google.rpc.BadRequest': BadRequest,
  'type.googleapis.com/google.rpc.RequestInfo': RequestInfo,
  'type.googleapis.com/google.rpc.ResourceInfo': ResourceInfo,
  'type.googleapis.com/google.rpc.Help': Help,
  'type.googleapis.com/google.rpc.LocalizedMessage': LocalizedMessage,
  'type.googleapis.com/google.rpc.ErrorInfo': ErrorInfo
}

function httpMapping(code: Code): ResponseInit | undefined {
  switch (code) {
    case Code.OK:
      return { status: 200, statusText: 'OK' }
    case Code.INVALID_ARGUMENT:
    case Code.FAILED_PRECONDITION:
    case Code.OUT_OF_RANGE:
      return { status: 400, statusText: 'Bad Request' }
    case Code.UNAUTHENTICATED:
      return { status: 401, statusText: 'Unauthorized' }
    case Code.PERMISSION_DENIED:
      return { status: 403, statusText: 'Forbidden' }
    case Code.NOT_FOUND:
      return { status: 404, statusText: 'Not Found' }
    case Code.ALREADY_EXISTS:
    case Code.ABORTED:
      return { status: 409, statusText: 'Conflict' }
    case Code.RESOURCE_EXHAUSTED:
      return { status: 429, statusText: 'Too Many Requests' }
    case Code.CANCELLED:
      return { status: 499, statusText: 'Client Closed Request' }
    case Code.UNKNOWN:
    case Code.INTERNAL:
    case Code.DATA_LOSS:
      return { status: 500, statusText: 'Internal Server Error' }
    case Code.UNIMPLEMENTED:
      return { status: 501, statusText: 'Not Implemented' }
    case Code.UNAVAILABLE:
      return { status: 503, statusText: 'Service Unavailable' }
    case Code.DEADLINE_EXCEEDED:
      return { status: 504, statusText: 'Gateway Timeout' }
  }
}

function codeMapping(code: string): Code {
  switch (code) {
    case 'INVALID_ARGUMENT':
      return Code.INVALID_ARGUMENT
    case 'FAILED_PRECONDITION':
      return Code.FAILED_PRECONDITION
    case 'OUT_OF_RANGE':
      return Code.OUT_OF_RANGE
    case 'UNAUTHENTICATED':
      return Code.UNAUTHENTICATED
    case 'PERMISSION_DENIED':
      return Code.PERMISSION_DENIED
    case 'NOT_FOUND':
      return Code.NOT_FOUND
    case 'ALREADY_EXISTS':
      return Code.ALREADY_EXISTS
    case 'ABORTED':
      return Code.ABORTED
    case 'RESOURCE_EXHAUSTED':
      return Code.RESOURCE_EXHAUSTED
    case 'CANCELLED':
      return Code.CANCELLED
    case 'UNKNOWN':
      return Code.UNKNOWN
    case 'INTERNAL':
      return Code.INTERNAL
    case 'DATA_LOSS':
      return Code.DATA_LOSS
    case 'UNIMPLEMENTED':
      return Code.UNIMPLEMENTED
    case 'UNAVAILABLE':
      return Code.UNAVAILABLE
    case 'DEADLINE_EXCEEDED':
      return Code.DEADLINE_EXCEEDED
    default:
      return Code.OK
  }
}

export { grpcClient, StatusError, httpMapping, isGrpcError }
export type { GrpcError }
