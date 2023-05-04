import {
  grpcClient,
  httpMapping,
  isGrpcError,
  StatusError
} from '~/lib/proto.server'
import { json } from '@remix-run/node'
import type { Amount } from '~/generated/protobuf-ts/backend/v1/backend'
import type {
  GetTransactionDetailsResponse,
  ListWalletsResponse,
  PaginationRequest,
  WalletDetails
} from '~/generated/protobuf-ts/backend/admin/v1/backend'
import { DateTime } from 'luxon'

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

interface Transaction {
  walletID: string
  id: string
  title: string
  type: string
  status: string
  date: string
  amount: string
  source: string
  destination: string
}

export async function GetWalletTransactions(
  request: Request,
  walletID: string
): Promise<Transaction[]> {
  const cookie = String(request.headers.get('cookie'))
  let response = await grpcClient
    .listTransactions(
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

  return response.response.transactions.map((transaction) => {
    return {
      walletID,
      id: transaction.id,
      title: 'TODO',
      status: 'Complete',
      type: transaction.type,
      date: DateTime.fromSeconds(
        parseInt(transaction.timestamp?.seconds ?? '')
      ).toFormat('dd MMM yyyy'),
      amount: transaction.amount.toString(),
      source: transaction.source,
      destination: transaction.destination
    }
  })
}

export async function GetWalletTransactionDetails(
  request: Request,
  walletID: string,
  transactionID: string
): Promise<GetTransactionDetailsResponse> {
  const cookie = String(request.headers.get('cookie'))
  let response = await grpcClient
    .getTransactionDetails(
      { walletID, transactionID },
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

/**
 * RPC endpoints required for the admin panel
 *
 * - Transactions
 *     - List
 *     - Get detailed
 * - Identities
 *     - List
 *     - Get Detailed
 * - KYC Status
 * - Linked Accounts
 *     - List
 *     - Detailed
 * - Audit
 *     - List
 */

export async function GetWalletIdentities(
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

export async function GetWalletIdentityDetails(
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

export async function GetWalletLinkedAccounts(
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

export async function GetWalletLinkedAccountDetails(
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

export async function GetWalletAudits(
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
