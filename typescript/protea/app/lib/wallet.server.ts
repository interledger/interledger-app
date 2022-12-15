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
  PaginationRequest,
  PaymentPointer
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

type LinkedAccountsResponse = {
  linkedAccounts: Array<{
    id: string
    name: string
    type: string
    icon: string
  }>
  canTopUp: boolean
  canWithdraw: boolean
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

  const linkedAccounts = response.response.linkedAccounts.map(
    (linkedAccount) => {
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
  )

  return {
    linkedAccounts,
    canTopUp: linkedAccounts.filter(({ type }) => type == 'card').length > 0,
    canWithdraw: linkedAccounts.filter(({ type }) => type == 'bank').length > 0
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

export async function getTransactions(
  request: Request,
  input: PaginationRequest = { page: 1, pageSize: 20 }
): Promise<Transaction[]> {
  const cookie = String(request.headers.get('cookie'))
  return openPaymentsClient
    .listTransactions(input, {
      meta: {
        cookies: cookie || ''
      }
    })
    .then((response) =>
      response.response.transactions.map((trx) => ({
        id: trx.id,
        type: trx.type,
        icon: trx.type == 'outgoing' ? 'north_east' : 'south_west',
        title: trx.type == 'outgoing' ? trx.destination : trx.source,
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
  status: string
  paymentPointer: string
  subTotal: string
  fees: string
  total: string
  date: string
  note: string
}

export async function getTransaction(
  request: Request,
  type: string,
  id: string
): Promise<DetailedTransaction> {
  const cookie = String(request.headers.get('cookie'))
  if (type == 'outgoing') {
    return openPaymentsClient
      .lookupOutgoingPayment(
        { id },
        {
          meta: {
            cookies: cookie || ''
          }
        }
      )
      .then((response) => {
        const status =
          response.response.sentAmount?.amount ==
          response.response.sendAmount?.amount
            ? 'Sent'
            : 'Pending'

        return {
          id: response.response.id.split('/').at(-1) as string,
          status,
          paymentPointer: response.response.toPaymentPointer,
          subTotal: formatAmount(response.response.sendAmount),
          fees: '$ 0.00',
          total: formatAmount(response.response.sendAmount),
          date: DateTime.fromSeconds(
            parseInt(response.response.updatedAt?.seconds ?? '')
          ).toLocaleString(DateTime.DATETIME_FULL),
          note: response.response.description
        }
      })
      .catch((error) => {
        const status = StatusError(error)
        if (isGrpcError(status)) {
          throw json({}, httpMapping(status.code))
        }
        throw json({}, { status: 404, statusText: "Can't find transaction." })
      })
  } else if (type == 'incoming') {
    return openPaymentsClient
      .lookupIncomingPayment(
        { id },
        {
          meta: {
            cookies: cookie || ''
          }
        }
      )
      .then((response) => {
        return {
          id: response.response.id.split('/').at(-1) as string,
          status: 'Received',
          paymentPointer: response.response.fromPaymentPointer,
          subTotal: formatAmount(response.response.receivedAmount),
          fees: '$ 0.00',
          total: formatAmount(response.response.receivedAmount),
          date: DateTime.fromSeconds(
            parseInt(response.response.updatedAt?.seconds ?? '')
          ).toLocaleString(DateTime.DATETIME_FULL),
          note: response.response.externalRef
        }
      })
      .catch((error) => {
        const status = StatusError(error)
        if (isGrpcError(status)) {
          throw json({}, httpMapping(status.code))
        }
        throw json({}, { status: 404, statusText: "Can't find transaction." })
      })
  } else
    throw json({}, { status: 400, statusText: 'Invalid transaction type.' })
}
