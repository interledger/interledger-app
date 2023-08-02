import { json, redirect } from '@remix-run/node'
import { route } from 'routes-gen'
import type {
  Amount,
  Features,
  GetPublicWalletDetailsResponse,
  Identity,
  IndividualKYCResponse,
  KYCStatusResponse,
  LinkedAccount,
  ListContactsRequest,
  ListContactsResponse,
  ListTransactionsResponse,
  PaginationRequest,
  PublicWalletInfo,
  WalletInfo
} from '~/generated/protobuf-ts/backend/v1/backend'
import { Code } from '~/generated/protobuf-ts/google/rpc/code'
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

export async function getFeatures(request: Request): Promise<Features> {
  const response = await grpcClient
    .listFeatures(
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
  const cookie = String(request.headers.get('cookie'))
  let response = await grpcClient
    .getWalletInfo(
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
  } else if (!response.response.hasWalletAddress) {
    throw redirect(route('/wallet-address'))
  }

  return response.response
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
  const cookie = String(request.headers.get('cookie'))
  return grpcClient
    .listTransactionsWithPending(input, {
      meta: {
        cookies: cookie || ''
      }
    })
    .then((response) => ({
      ...response.response
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

export type createCardError =
  | 'Failed precondition: ErrUnsupportedCard'
  | 'Failed precondition: ErrUnsupportedCountry'
  | 'Failed precondition: ErrMaxCardsAdded'
  | 'Already exists: ErrDuplicateCard'
  | 'Internal server error'

export async function createCard(
  request: Request,
  tokenID: string
): Promise<{
  error: createCardError | undefined
  httpMapping: ResponseInit | undefined
  linkedAccountID: string
}> {
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
    return {
      error: response.message as createCardError,
      httpMapping: httpMapping(response.code),
      linkedAccountID: ''
    }
  }

  return {
    error: undefined,
    httpMapping: httpMapping(Code.OK),
    linkedAccountID: response.response.id
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

export async function SetKYCStatusPending(
  request: Request,
): Promise<void> {
  const cookie = String(request.headers.get('cookie'))
  const response = await grpcClient
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
