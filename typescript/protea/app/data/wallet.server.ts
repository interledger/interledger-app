import type { PartialMessage } from '@bufbuild/protobuf'
import type { PlainMessage } from '@bufbuild/protobuf/dist/types/message'
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
import { isConnectError } from '~/lib/error.server'
import { grpc } from '~/lib/grpc.server'

export const PAYMENT_POINTER_BASE = process.env.PAYMENT_POINTER_BASE

export const formatAmount = (amount?: Amount): string => {
  if (typeof amount == 'undefined') return '$ 0.00'
  const symbol = amount.asset == 'USD' ? '$' : amount.asset
  return `${symbol} ${(Number(amount.amount) / 100).toFixed(amount.assetScale)}`
}

export async function getKycStatus(
  request: Request
): Promise<KYCStatusResponse> {
  const response = await grpc.kYCStatus(request, {})

  if (isConnectError(response)) throw response.errorResponse

  return response
}

export async function getFeatures(request: Request): Promise<Features> {
  const response = await grpc.listFeatures(request, {})

  if (isConnectError(response)) throw response.errorResponse

  return response
}

export async function getWalletId(request: Request): Promise<string> {
  let response = await grpc.getCurrentWallet(request, {})

  if (isConnectError(response)) throw response.errorResponse

  return response.id
}

export async function getPublicWalletDetails(
  request: Request,
  id: string
): Promise<GetPublicWalletDetailsResponse> {
  let response = await grpc.getPublicWalletDetails(request, {
    id
  })

  if (isConnectError(response)) throw response.errorResponse

  return response
}

export async function getPublicWalletInfo(
  request: Request,
  walletAddress: string
): Promise<PublicWalletInfo> {
  let response = await grpc.getPublicWalletInfo(request, {
    walletAddress
  })

  if (isConnectError(response)) throw response.errorResponse

  return response
}

export async function getWalletInfo(request: Request): Promise<WalletInfo> {
  let response = await grpc.getWalletInfo(request, {})

  if (isConnectError(response)) {
    throw response.errorResponse
  } else if (!response.hasWalletAddress) {
    throw redirect(route('/wallet-address'))
  }

  return response
}

export type FormattedLinkedAccount = {
  name: string
  type: string
  icon: string
} & PlainMessage<LinkedAccount>

type LinkedAccountsResponse = {
  bankAccounts: Array<FormattedLinkedAccount>
  cardAccounts: Array<FormattedLinkedAccount>
  balanceAccounts: Array<FormattedLinkedAccount>
}

export async function getLinkedAccount(
  request: Request,
  id: string
): Promise<FormattedLinkedAccount> {
  const response = await grpc.getLinkedAccount(request, {
    id
  })

  if (isConnectError(response)) throw response.errorResponse

  return formatLinkedAccount(response)
}

export async function getLinkedAccounts(
  request: Request
): Promise<LinkedAccountsResponse> {
  const [xagoBalances, accounts] = await Promise.all([
    grpc.getXagoBalances(request, {}),
    grpc.getLinkedAccounts(request, {})
  ])

  if (isConnectError(xagoBalances)) throw xagoBalances.errorResponse
  if (isConnectError(accounts)) throw accounts.errorResponse

  const linkedAccounts = accounts.linkedAccounts.map(formatLinkedAccount)
  const balanceAccounts = linkedAccounts
    .filter(({ type }) => type == 'wallet')
    .map((acc) => {
      let balance = xagoBalances.balances.find(
        (balance) => balance.linkedAccount == acc.id
      )
      if (balance) {
        acc.title = formatAmount(balance.available)
      }

      return acc
    })

  return {
    bankAccounts: linkedAccounts.filter(({ type }) => type == 'bank'),
    cardAccounts: linkedAccounts.filter(({ type }) => type == 'card'),
    balanceAccounts
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
    case 'balance':
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
  const response = await grpc.listTransactionsWithPending(request, input)

  if (isConnectError(response)) throw response.errorResponse
  return response
}
