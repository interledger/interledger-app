import type { PartialMessage } from '@bufbuild/protobuf'
import { redirect } from '@remix-run/node'
import { route } from 'routes-gen'
import type {
  Amount,
  Features,
  GetPublicWalletDetailsResponse,
  KYCStatusResponse,
  LinkedAccount,
  ListTransactionsResponse,
  PaginationRequest,
  PublicWalletInfo,
  WalletInfo
} from '~/generated/connect/backend/v1/backend_pb'
import { connectClient } from '~/lib/connect.server'
import { isConnectError } from '~/lib/error.server'

export const PAYMENT_POINTER_BASE = process.env.PAYMENT_POINTER_BASE

export const formatAmount = (amount?: Amount): string => {
  if (typeof amount == 'undefined') return '$ 0.00'
  const symbol = amount.asset == 'USD' ? '$' : amount.asset
  return `${symbol} ${(Number(amount.amount) / 100).toFixed(amount.assetScale)}`
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

  if (isConnectError(response)) throw response.errorResponse

  return response.id
}

export async function getPublicWalletDetails(
  request: Request,
  id: string
): Promise<GetPublicWalletDetailsResponse> {
  let response = await connectClient.getPublicWalletDetails(request, {
    id
  })

  if (isConnectError(response)) throw response.errorResponse

  return response
}

export async function getPublicWalletInfo(
  request: Request,
  walletAddress: string
): Promise<PublicWalletInfo> {
  let response = await connectClient.getPublicWalletInfo(request, {
    walletAddress
  })

  if (isConnectError(response)) throw response.errorResponse

  return response
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
  input: PartialMessage<PaginationRequest>
): Promise<ListTransactionsResponse> {
  const response = await connectClient.listTransactionsWithPending(
    request,
    input
  )

  if (isConnectError(response)) throw response.errorResponse
  return response
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
  const response = await connectClient.lookupTransaction(request, { id })

  if (isConnectError(response)) throw response.errorResponse

  return {
    id: response.id,
    walletUrl: response.type.includes('outgoing')
      ? response.destination
      : response.source,
    type: response.type,
    title: response.title,
    status: response.state,
    refundState: response.refundState,
    reference: response.reference,
    accountTitle: response.accountTitle,
    icon: response.destinationIdentityType,
    subTotal: response.subtotal,
    fees: response.fees,
    total: response.formattedAmount,
    date: response.formattedDate,
    time: response.formattedTime,
    transfers: response.transfers.map((transfer) => {
      return {
        linkedAccountId: transfer.linkedAccountId,
        type: transfer.type,
        amount: formatAmount(transfer.amount),
        status: transfer.state
      }
    })
  }
}
