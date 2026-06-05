import type { PlainMessage } from '@bufbuild/protobuf'
import type { LinkedAccount } from '~/generated/connect/backend/v1/backend_pb'
import { isConnectError } from '~/lib/error.server'
import { grpc } from '~/lib/grpc.server'

export type FormattedLinkedAccount = {
  name: string
  type: string
  icon: string
  enabled?: boolean
} & PlainMessage<LinkedAccount>

type LinkedAccountsResponse = {
  interacAccounts: Array<FormattedLinkedAccount>
  bankAccounts: Array<FormattedLinkedAccount>
  cardAccounts: Array<FormattedLinkedAccount>
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
  const accounts = await grpc.getLinkedAccounts(request, {})

  if (isConnectError(accounts)) throw accounts.errorResponse

  const linkedAccounts = accounts.linkedAccounts.map((account) =>
    formatLinkedAccount(account)
  )

  return {
    bankAccounts: linkedAccounts.filter(({ type }) => type == 'bank'),
    cardAccounts: linkedAccounts.filter(({ type }) => type == 'card'),
    interacAccounts: linkedAccounts.filter(({ type }) => type == 'interac')
  }
}

export async function getLinkedAccountsForPayment(
  request: Request,
  id: string
): Promise<FormattedLinkedAccount[]> {
  const accounts = await grpc.getLinkedAccountsForPayment(request, {
    paymentId: id
  })

  if (isConnectError(accounts)) throw accounts.errorResponse

  // TODO Format, filter and sort this in the backend.
  return (
    accounts.linkedAccounts
      .map((account) => formatLinkedAccount(account.details!, account.enabled))
      .sort((a, b) => {
        // bubble the balances to the top
        if (a.type != b.type && b.type == 'balance') return -1
        return 0
      })
      // Temporary: restrict P2P PAYMENT source to balance accounts.
      .filter((acc) => acc.canSend && acc.type == 'balance')
  )
}

export async function getLinkedAccountsForWithdraw(
  request: Request,
  linkedAccountId: string
): Promise<FormattedLinkedAccount[]> {
  const accounts = await grpc.getLinkedAccountsForWithdraw(request, {
    linkedAccountId
  })

  if (isConnectError(accounts)) throw accounts.errorResponse

  return accounts.linkedAccounts
    .map((account) => formatLinkedAccount(account.details!, account.enabled))
    .sort((a, b) => {
      // bubble the balances to the top
      if (a.type != b.type && b.type == 'balance') return -1
      return 0
    })
}

export async function getLinkedAccountsForDeposit(
  request: Request,
  linkedAccountId: string
): Promise<FormattedLinkedAccount[]> {
  const accounts = await grpc.getLinkedAccountsForDeposit(request, {
    linkedAccountId
  })

  if (isConnectError(accounts)) throw accounts.errorResponse

  return accounts.linkedAccounts
    .map((account) => formatLinkedAccount(account.details!, account.enabled))
    .sort((a, b) => {
      // bubble the balances to the top
      if (a.type != b.type && b.type == 'balance') return -1
      return 0
    })
}

export async function getBalancesForTransfer(
  request: Request
): Promise<FormattedLinkedAccount[]> {
  const accounts = await grpc.getLinkedAccounts(request, {})

  if (isConnectError(accounts)) throw accounts.errorResponse

  const linkedAccounts = accounts.linkedAccounts.map((account) =>
    formatLinkedAccount(account, true)
  )

  return linkedAccounts.filter(({ type }) => type == 'balance')
}

const formatLinkedAccount = (
  linkedAccount: LinkedAccount,
  enabled?: boolean
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
    case 'bank_account':
    case 'bankAccount':
      type = 'bank'
      name = linkedAccount.title
      icon = 'account_balance'
      break
    case 'interac':
      type = 'interac'
      name = linkedAccount.title
      icon = 'account_balance'
      break
    case 'balance':
    case 'wallet':
      type = 'balance'
      name = linkedAccount.title
      icon = 'wallet'
      break
  }
  return {
    ...linkedAccount,
    name,
    type,
    icon,
    enabled
  }
}
