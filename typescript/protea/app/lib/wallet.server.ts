import {
  grpcClient,
  httpMapping,
  isGrpcError,
  openPaymentsClient,
  StatusError
} from '~/lib/proto.server'
import { json, redirect } from '@remix-run/node'
import type {
  Amount,
  KYCStatusResponse,
  LinkedAccount,
  PaginationRequest,
  PaymentPointer,
  Transaction as GrpcTransaction
} from '~/generated/protobuf-ts/backend/v1/backend'
import { Code } from '~/generated/protobuf-ts/google/rpc/code'
import { DateTime } from 'luxon'
import { route } from 'routes-gen'

export const PAYMENT_POINTER_BASE = process.env.PAYMENT_POINTER_BASE

export const formatAmount = (amount?: Amount): string => {
  if (typeof amount == 'undefined') return '$ 0.00'
  const symbol = amount.asset == 'USD' ? '$' : amount.asset
  return `${symbol} ${(parseInt(amount.amount) / 100).toFixed(
    amount.assetScale
  )}`
}

export async function getKycStatus(
  request: Request
): Promise<KYCStatusResponse> {
  const response = await grpcClient
    .kYCStatus(
      {},
      {
        meta: {
          cookies: String(request.headers.get('cookie')) || ''
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
  } else if (response.response.pointers.length == 0) {
    throw redirect(route('/payment-pointer'))
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
    if (response.code == Code.NOT_FOUND) return '$ 0.00'
    throw json({}, httpMapping(response.code))
  }

  return formatAmount({
    asset: 'USD',
    assetScale: 2,
    amount: response.response.available
  })
}

type FormattedLinkedAccount = {
  id: string
  name: string
  type: string
  icon: string
}

type LinkedAccountsResponse = {
  linkedAccounts: Array<FormattedLinkedAccount>
  canTopUp: boolean
  canWithdraw: boolean
}

export async function getLinkedAccount(
  request: Request,
  id: string
): Promise<FormattedLinkedAccount> {
  const cookie = String(request.headers.get('cookie'))
  const response = await grpcClient
    .getLinkedAccount(
      {
        id
      },
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

  return formatLinkedAccount(response.response)
}

export async function getLinkedAccounts(
  request: Request
): Promise<LinkedAccountsResponse> {
  const cookie = String(request.headers.get('cookie'))
  const response = await grpcClient
    .getLinkedAccounts(
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

  const linkedAccounts =
    response.response.linkedAccounts.map(formatLinkedAccount)

  return {
    linkedAccounts,
    canTopUp: linkedAccounts.filter(({ type }) => type == 'card').length > 0,
    canWithdraw: linkedAccounts.filter(({ type }) => type == 'bank').length > 0
  }
}

const formatLinkedAccount = (
  linkedAccount: LinkedAccount
): FormattedLinkedAccount => {
  let type = '',
    name = '',
    icon = ''
  switch (linkedAccount.type) {
    case 'sendCard':
      type = 'card'
      name = `Card ending ${linkedAccount.mask}`
      icon = 'credit_card'
      break
    case 'bankAccount':
      type = 'bank'
      name = `${linkedAccount.name} ${linkedAccount.mask}`
      icon = 'account_balance'
      break
    case 'wallet':
      type = 'wallet'
      name = 'Cash balance'
      icon = 'wallet'
      break
  }
  return {
    id: linkedAccount.id,
    name,
    type,
    icon
  }
}

export type Transaction = {
  id: string
  type: string
  icon: string
  title: string
  total: string
  time: string
  date: string
}

export async function getPendingTransactions(
  request: Request
): Promise<Transaction[]> {
  const cookie = String(request.headers.get('cookie'))
  return openPaymentsClient
    .listPendingTransactions(
      { page: 1, pageSize: 20 },
      {
        meta: {
          cookies: cookie || ''
        }
      }
    )
    .then((response) =>
      response.response.transactions.map((trx) => ({
        id: trx.id,
        type: trx.type,
        icon: 'schedule',
        title: trx.type == 'outgoing' ? trx.destination : trx.source,
        total: formatAmount(trx.amount),
        time: 'Pending',
        date: DateTime.fromSeconds(
          parseInt(trx.timestamp?.seconds ?? '')
        ).toFormat('dd MMM yyyy')
      }))
    )
    .catch((error) => {
      const status = StatusError(error)
      if (isGrpcError(status)) {
        throw json({}, httpMapping(status.code))
      }
      return []
    })
}

function transactionIcon(type: string): string {
  switch (type) {
    case 'open_payments_outgoing':
      return 'north_east'
    case 'open_payments_incoming':
      return 'south_west'
    case 'machnet_wallet_topup':
      return 'credit_card'
    case 'machnet_wallet_withdrawal':
      return 'account_balance'
    default:
      return ''
  }
}

function transactionTitle(trx: GrpcTransaction): string {
  switch (trx.type) {
    case 'open_payments_outgoing':
      return trx.destination
    case 'open_payments_incoming':
      return trx.source
    case 'machnet_wallet_topup':
      return 'Top up'
    case 'machnet_wallet_withdrawal':
      return 'Withdrawal'
    default:
      return ''
  }
}

export async function getTransactions(
  request: Request,
  input: PaginationRequest = { page: 1, pageSize: 20 }
): Promise<Transaction[]> {
  const cookie = String(request.headers.get('cookie'))
  return grpcClient
    .listTransactions(input, {
      meta: {
        cookies: cookie || ''
      }
    })
    .then((response) =>
      response.response.transactions.map((trx) => ({
        id: trx.id,
        type: trx.type,
        icon: trx.state == 'Pending' ? 'schedule' : transactionIcon(trx.type),
        title: transactionTitle(trx),
        total: formatAmount(trx.amount),
        time: DateTime.fromSeconds(
          parseInt(trx.timestamp?.seconds ?? '')
        ).toFormat('HH:mm'),
        date: DateTime.fromSeconds(
          parseInt(trx.timestamp?.seconds ?? '')
        ).toFormat('dd MMM yyyy')
      }))
    )
    .catch((error) => {
      const status = StatusError(error)
      if (isGrpcError(status)) {
        throw json({}, httpMapping(status.code))
      }
      return []
    })
}

export type DetailedTransaction = {
  id: string
  type: string
  foreignId: string
  status: string
  subTotal: string
  fees: string
  total: string
  date: string
  transfers: Array<DetailedTransfer>
}

export type DetailedTransfer = {
  linkedAccountId: string
  type: string
  amount: string
  status: string
}

export async function getTransaction(
  request: Request,
  type: string,
  id: string
): Promise<DetailedTransaction> {
  const cookie = String(request.headers.get('cookie'))

  return grpcClient
    .lookupTransaction(
      { id },
      {
        meta: {
          cookies: cookie || ''
        }
      }
    )
    .then((resp) => {
      return {
        id: resp.response.id,
        type: resp.response.type,
        foreignId: resp.response.foreignId,
        status: resp.response.state,
        subTotal: formatAmount(resp.response.amount),
        fees: '$ 0.00',
        total: formatAmount(resp.response.amount),
        date: DateTime.fromSeconds(
          parseInt(resp.response.timestamp?.seconds ?? '')
        ).toLocaleString(DateTime.DATETIME_FULL),
        transfers: resp.response.transfers.map((transfer) => {
          return {
            linkedAccountId: transfer.linkedAccountId,
            type: transfer.type,
            amount: formatAmount(transfer.amount),
            status: transfer.state
          }
        })
      }
    })
}
