import { ChannelCredentials } from '@grpc/grpc-js'
import { BackendServiceClient } from '~/generated/protobuf-ts/backend/v1/backend_client'
import { GrpcTransport } from '@protobuf-ts/grpc-transport'
import { base64decode } from '@protobuf-ts/runtime'
import {
  RetryInfo,
  DebugInfo,
  QuotaFailure,
  PreconditionFailure,
  BadRequest,
  RequestInfo,
  ResourceInfo,
  Help,
  LocalizedMessage,
  ErrorInfo
} from '~/generated/protobuf-ts/google/rpc/error_details'
import { Status } from '~/generated/protobuf-ts/google/rpc/status'
import type { RpcError } from '@protobuf-ts/runtime-rpc'
import { Any } from '~/generated/protobuf-ts/google/protobuf/any'

const transport = new GrpcTransport({
  host: 'dns:backend-admin:443',
  channelCredentials: ChannelCredentials.createInsecure()
})

let grpcClient: BackendServiceClient

declare global {
  var __grpcClient: BackendServiceClient | undefined
}

// this is needed because in development we don't want to restart
// the server with every change, but we want to make sure we don't
// create a new connection to the Client with every change either.
if (process.env.NODE_ENV === 'production') {
  grpcClient = new BackendServiceClient(transport)
} else {
  if (!global.__grpcClient) {
    global.__grpcClient = new BackendServiceClient(transport)
  }
  grpcClient = global.__grpcClient
}

export { grpcClient, StatusError, isGrpcError }

export interface GrpcError {
  code: number
  details: any[]
}

function isGrpcError(res: any): res is GrpcError {
  return (res as GrpcError).details !== undefined
}

/**
 * This function is intended to deserialise and parse the RpcError
 * returned by grpcClient functions. It will return the error message
 * in the same shape of successful rpc calls, with the error message in the response.
 * @param err The error received from a grpc call.
 * @returns GrpcError - the error code and details
 */
function StatusError(err: RpcError): GrpcError {
  if (!err.meta) {
    // return null
    throw new Error('No meta on error')
  }

  const buffer = base64decode(err.meta['grpc-status-details-bin'] as string)

  if (!buffer || typeof buffer === 'string') {
    // return null
  }

  let status: Status | undefined

  status = Status.fromBinary(buffer)
  const details: any[] = status.details
    .map((detail) => {
      return Any.unpack(detail, typeRegistry[detail.typeUrl]) || null
    })
    .filter(Boolean)

  return {
    code: status.code,
    details
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
