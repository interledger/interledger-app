import {
  grpcClient,
  httpMapping,
  isGrpcError,
  openPaymentsClient,
  StatusError
} from '~/lib/proto.server'
import { json } from '@remix-run/node'
import type { PaymentPointer } from '~/generated/protobuf-ts/backend/v1/backend'
import { Code } from '~/generated/protobuf-ts/google/rpc/code'
import type { Amount } from '~/generated/protobuf-ts/backend/v1/backend'

export const PAYMENT_POINTER_BASE = process.env.PAYMENT_POINTER_BASE

export const formatAmount = (amount?: Amount): string => {
  if (typeof amount == 'undefined') return '$ 0.00'
  const symbol = amount.asset == 'USD' ? '$' : amount.asset
  return `${symbol} ${(parseInt(amount.amount) / 100).toFixed(
    amount.assetScale
  )}`
}

export async function getWalletPaymentPointer(
  request: Request
): Promise<PaymentPointer> {
  const cookie = String(request.headers.get('cookie'))
  let response = await openPaymentsClient
    .listWalletPaymentPointers(
      {},
      {
        meta: {
          cookies: cookie || ''
        }
      }
    )
    .then((v) => v)
    .catch(StatusError)
  if (isGrpcError(response)) {
    throw json({}, httpMapping(response.code))
  }

  return response.response.pointers[0]
}

export async function getWalletBalance(request: Request): Promise<string> {
  const cookie = String(request.headers.get('cookie'))
  let response = await grpcClient
    .getWalletBalance(
      {},
      {
        meta: {
          cookies: cookie || ''
        }
      }
    )
    .then((v) => v)
    .catch(StatusError)
  if (isGrpcError(response)) {
    if (response.code == Code.NOT_FOUND) return ''
    throw json({}, httpMapping(response.code))
  }

  return formatAmount({
    asset: 'USD',
    assetScale: 2,
    amount: response.response.available
  })
}
