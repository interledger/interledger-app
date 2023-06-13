import { json, redirect } from '@remix-run/node'
import { DateTime } from 'luxon'
import { route } from 'routes-gen'
import type {
  Amount,
  GetPublicWalletDetailsResponse,
  Transaction as GrpcTransaction,
  Identity,
  IndividualKYCResponse,
  KYCStatusResponse,
  LinkedAccount,
  ListContactsRequest,
  ListContactsResponse,
  PaginationRequest,
  PaymentPointer
} from '~/generated/protobuf-ts/backend/v1/backend'
import {
  StatusError,
  grpcClient,
  httpMapping,
  isGrpcError,
  openPaymentsClient
} from '~/lib/proto.server'

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

export async function getKycDetails(
  request: Request
): Promise<IndividualKYCResponse> {
  const response = await grpcClient
    .getIndividualKYC(
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

export async function getWalletId(request: Request): Promise<string> {
  const cookie = String(request.headers.get('cookie'))
  let response = await grpcClient
    .getCurrentWallet(
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

  return response.response.id
}

export async function getPublicWalletDetails(
  request: Request,
  id: string
): Promise<GetPublicWalletDetailsResponse> {
  const cookie = String(request.headers.get('cookie'))
  let response = await grpcClient
    .getPublicWalletDetails(
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
    throw redirect(route('/wallet-address'))
  }

  return response.response.pointers[0]
}

type FormattedLinkedAccount = {
  id: string
  name: string
  nickname: string
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
    case 'card':
    case 'sendCard':
      type = 'card'
      name = `**** ${linkedAccount.mask}`
      icon = 'credit_card'
      break
    case 'bankAccount':
      type = 'bank'
      name = `**** ${linkedAccount.mask}`
      icon = 'account_balance'
      break
    case 'wallet':
      type = 'wallet'
      name = 'Cash balance'
      icon = 'wallet'
      break
  }
  return {
    ...linkedAccount,
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

function transactionIcon(type: string): string {
  switch (type) {
    case 'open_payments_outgoing':
      return 'north_east'
    case 'open_payments_incoming':
      return 'south_west'
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
    default:
      return ''
  }
}

type getTransactionsWithPendingResponse = {
  nextPageToken: string
  transactions: Transaction[]
}

export async function getTransactionsWithPending(
  request: Request,
  input: PaginationRequest
): Promise<getTransactionsWithPendingResponse> {
  const cookie = String(request.headers.get('cookie'))
  return grpcClient
    .listTransactionsWithPending(input, {
      meta: {
        cookies: cookie || ''
      }
    })
    .then((response) => ({
      nextPageToken: response.response.nextPageToken,
      transactions: response.response.transactions.map((trx) => ({
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
    }))
    .catch((error) => {
      const status = StatusError(error)
      if (isGrpcError(status)) {
        throw json({}, httpMapping(status.code))
      }
      return { nextPageToken: '', transactions: [] }
    })
}

type getWalletContactsResponse = {
  nextPageToken: string
  contacts: ListContactsResponse['contacts']
}

export async function getWalletContacts(
  request: Request,
  input: ListContactsRequest
): Promise<getWalletContactsResponse> {
  const cookie = String(request.headers.get('cookie'))
  return grpcClient
    .listContacts(input, {
      meta: {
        cookies: cookie || ''
      }
    })
    .then((response) => ({
      nextPageToken: response.response.nextPageToken,
      contacts: response.response.contacts
    }))
    .catch((error) => {
      const status = StatusError(error)
      if (isGrpcError(status)) {
        throw json({}, httpMapping(status.code))
      }
      return { nextPageToken: '', contacts: [] }
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

export async function createCard(
  request: Request,
  tokenID: string
): Promise<void> {
  const cookie = String(request.headers.get('cookie'))
  const response = await grpcClient
    .createCard(
      { tokenID },
      {
        meta: {
          cookies: cookie || ''
        },
        timeout: 60 * 1000
      }
    )
    .then((v) => v)
    .catch(StatusError)
  if (isGrpcError(response)) {
    throw json({}, httpMapping(response.code))
  }
}

export async function getLinkedIdentities(
  request: Request
): Promise<Record<string, Identity[]>> {
  const cookie = String(request.headers.get('cookie'))
  const response = await grpcClient
    .listIdentities(
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

  return response.response.identities.reduce(
    (acc, identity) => {
      if (!acc[identity.platform]) {
        acc[identity.platform] = []
      }
      acc[identity.platform].push(identity)
      return acc
    },

    {} as Record<string, Identity[]>
  )
}

export async function getPublicLinkedIdentities(
  request: Request,
  walletId: string
): Promise<Array<Identity>> {
  const response = await grpcClient
    .listPublicIdentities({
      walletId
    })
    .then((v) => v)
    .catch(StatusError)
  if (isGrpcError(response)) {
    throw json({}, httpMapping(response.code))
  }

  return response.response.identities
}

export async function getIdentity(
  request: Request,
  id: string
): Promise<Identity> {
  const cookie = String(request.headers.get('cookie'))
  const response = await grpcClient
    .getIdentity(
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
  // TODO: remove, this should be handled by the backend
  if (!response.response.identity) {
    throw json({}, { status: 404 })
  }

  return response.response.identity
}

export async function getPublicIdentity(
  request: Request,
  signatureHash: string
): Promise<Identity> {
  const response = await grpcClient
    .getIdentityBySignatureHash(
      {
        signatureHash
      },
      {}
    )
    .then((v) => v)
    .catch(StatusError)
  if (isGrpcError(response)) {
    throw json({}, httpMapping(response.code))
  }
  // TODO: remove, this should be handled by the backend
  if (!response.response.identity) {
    throw json({}, { status: 404 })
  }

  return response.response.identity
}

export async function deleteTwitterIdentity(
  request: Request,
  id: string
): Promise<void> {
  const cookie = String(request.headers.get('cookie'))
  const response = await grpcClient
    .deleteIdentity(
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
}

export async function verifyTwitterIdentity(
  request: Request,
  id: string
): Promise<void> {
  const cookie = String(request.headers.get('cookie'))
  const response = await grpcClient
    .verifyTwitter(
      {
        identityId: id
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
}

export async function setTwitterIdentityPublic(
  request: Request,
  id: string,
  publicVal: boolean
): Promise<void> {
  const cookie = String(request.headers.get('cookie'))
  const response = await grpcClient
    .setIdentityPublic(
      {
        id,
        public: publicVal
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
}
