import { json } from '@remix-run/node'
import {
  grpcClient,
  httpMapping,
  isGrpcError,
  StatusError
} from '~/lib/proto.server'

export async function getConnection(request: Request, id: string) {
  let rpc = await grpcClient
    .getConnection(
      { id },
      {
        meta: {
          cookies: String(request.headers.get('cookie')) || ''
        }
      }
    )
    .then((v) => v)
    .catch(StatusError)

  if (isGrpcError(rpc)) {
    throw json({}, httpMapping(rpc.code))
  }

  return rpc.response
}
export async function getConnectionLimits(request: Request, id: string) {
  let rpc = await grpcClient
    .getConnectionLimits(
      { id },
      {
        meta: {
          cookies: String(request.headers.get('cookie')) || ''
        }
      }
    )
    .then((v) => v)
    .catch(StatusError)
  if (isGrpcError(rpc)) {
    throw json({}, httpMapping(rpc.code))
  }

  let limits = {
    daily:
      parseFloat(rpc.response.daily?.amount as string) *
      Math.pow(10, -(rpc.response.daily?.assetScale || 0)),
    monthly:
      parseFloat(rpc.response.monthly?.amount as string) *
      Math.pow(10, -(rpc.response.monthly?.assetScale || 0)),
    overall:
      parseFloat(rpc.response.overall?.amount as string) *
      Math.pow(10, -(rpc.response.overall?.assetScale || 0))
  }

  return limits
}
