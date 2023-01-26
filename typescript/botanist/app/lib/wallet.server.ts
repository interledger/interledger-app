import {
  grpcClient,
  httpMapping,
  isGrpcError,
  StatusError
} from '~/lib/proto.server'
import { json } from '@remix-run/node'
import type { Amount } from '~/generated/protobuf-ts/backend/v1/backend'
import type {
  ListWalletsResponse,
  ListWalletTransactionsResponse,
  PaginationRequest,
  WalletDetails
} from '~/generated/protobuf-ts/backend/admin/v1/backend'

export const PAYMENT_POINTER_BASE = process.env.PAYMENT_POINTER_BASE

export const formatAmount = (amount?: Amount): string => {
  if (typeof amount == 'undefined') return '$ 0.00'
  const symbol = amount.asset == 'USD' ? '$' : amount.asset
  return `${symbol} ${(parseInt(amount.amount) / 100).toFixed(
    amount.assetScale
  )}`
}

export async function ListWallets(
  request: Request,
  page: PaginationRequest
): Promise<ListWalletsResponse> {
  const cookie = String(request.headers.get('cookie'))
  let response = await grpcClient
    .listWallets(page, {
      meta: {
        cookies: cookie || ''
      }
    })
    .then((v) => v)
    .catch(StatusError)
  console.log('HELP')
  if (isGrpcError(response)) {
    throw json({}, httpMapping(response.code))
  }

  return response.response
}

export async function GetWalletDetails(
  request: Request,
  walletID: string
): Promise<WalletDetails> {
  const cookie = String(request.headers.get('cookie'))
  let response = await grpcClient
    .getWalletDetails(
      { walletID },
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

  return response.response
}

export async function ListWalletTransactions(
  request: Request,
  walletID: string,
  page: PaginationRequest
): Promise<ListWalletTransactionsResponse> {
  const cookie = String(request.headers.get('cookie'))
  let response = await grpcClient
    .listWalletTransactions(
      { walletID, page },
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

  return response.response
}
