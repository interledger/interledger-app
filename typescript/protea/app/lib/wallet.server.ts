import { json, redirect } from '@remix-run/node'
import { route } from 'routes-gen'
import type {
  Features,
  KYCStatusResponse,
  LinkedAccount,
  ListTransactionsResponse,
  WalletInfo
} from '~/generated/connect/backend/v1/backend_pb'
import type {
  Amount,
  GetPublicWalletDetailsResponse,
  Identity,
  ListContactsRequest,
  ListContactsResponse,
  PaginationRequest,
  PublicWalletInfo
} from '~/generated/protobuf-ts/backend/v1/backend'
import { connectClient } from '~/lib/connect.server'
import { isConnectError } from '~/lib/error.server'
import {
  StatusError,
  grpcClient,
  httpMapping,
  isGrpcError
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
  const response = await connectClient.kYCStatus(request, {})

  if (isConnectError(response)) throw response.errorResponse

  return response
}

export async function getFeatures(request: Request): Promise<Features> {
  const response = await connectClient.listFeatures(request, {})

  if (isConnectError(response)) throw response.errorResponse

  return response
}

export async function getWalletId(request: Request): Promise<string> {
  let response = await connectClient.getCurrentWallet(request, {})

  if (isConnectError(response)) {
    throw response.errorResponse
  }

  return response.id
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

export async function getPublicWalletInfo(
  request: Request,
  walletAddress: string
): Promise<PublicWalletInfo> {
  const cookie = String(request.headers.get('cookie'))
  let response = await grpcClient
    .getPublicWalletInfo(
      {
        walletAddress
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

export async function getWalletInfo(request: Request): Promise<WalletInfo> {
  let response = await connectClient.getWalletInfo(request, {})

  if (isConnectError(response)) {
    throw response.errorResponse
  } else if (!response.hasWalletAddress) {
    throw redirect(route('/wallet-address'))
  }

  return response
}

export type FormattedLinkedAccount = {
  id: string
  name: string
  nickname: string
  type: string
  icon: string
  canSend: boolean
  canReceive: boolean
}

type LinkedAccountsResponse = {
  bankAccounts: Array<FormattedLinkedAccount>
  cardAccounts: Array<FormattedLinkedAccount>
}

export async function getLinkedAccount(
  request: Request,
  id: string
): Promise<FormattedLinkedAccount> {
  const response = await connectClient.getLinkedAccount(request, {
    id
  })

  if (isConnectError(response)) throw response.errorResponse

  return formatLinkedAccount(response)
}

export async function getLinkedAccounts(
  request: Request
): Promise<LinkedAccountsResponse> {
  const response = await connectClient.getLinkedAccounts(request, {})

  if (isConnectError(response)) throw response.errorResponse

  const linkedAccounts = response.linkedAccounts.map(formatLinkedAccount)

  return {
    bankAccounts: linkedAccounts.filter(({ type }) => type == 'bank'),
    cardAccounts: linkedAccounts.filter(({ type }) => type == 'card')
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
      name = linkedAccount.title
      icon = 'credit_card'
      break
    case 'bankAccount':
      type = 'bank'
      name = linkedAccount.title
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

export async function getTransactionsWithPending(
  request: Request,
  input: PaginationRequest
): Promise<ListTransactionsResponse> {
  const response = await connectClient.listTransactionsWithPending(
    request,
    input
  )

  if (isConnectError(response)) throw response.errorResponse
  return response
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
  walletUrl: string
  type: string
  status: string
  icon: string
  title: string
  subTotal: string
  fees: string
  total: string
  date: string
  time: string
  accountTitle: string
  reference: string
  transfers: Array<DetailedTransfer>
  refundState: string
}

export type DetailedTransfer = {
  linkedAccountId: string
  type: string
  amount: string
  status: string
}

export async function getTransaction(
  request: Request,
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
        walletUrl: resp.response.type.includes('outgoing')
          ? resp.response.destination
          : resp.response.source,
        type: resp.response.type,
        title: resp.response.title,
        status: resp.response.state,
        refundState: resp.response.refundState,
        reference: resp.response.reference,
        accountTitle: resp.response.accountTitle,
        icon: resp.response.destinationIdentityType,
        subTotal: resp.response.subtotal,
        fees: resp.response.fees,
        total: resp.response.formattedAmount,
        date: resp.response.formattedDate,
        time: resp.response.formattedTime,
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
): Promise<Record<string, Identity[]>> {
  const response = await grpcClient
    .listPublicIdentities({
      walletId
    })
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
    .verifyIdentity(
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

export async function setKYCStatusPending(request: Request): Promise<void> {
  const cookie = String(request.headers.get('cookie'))
  await grpcClient
    .setKYCStatusPending(
      {},
      {
        meta: {
          cookies: cookie || ''
        }
      }
    )
    .then((v) => v)
    .catch(StatusError)
}
