import { href } from 'react-router'

import { redirectWithSnackbar } from '~/lib/snackbar.server'
import { ErrorHandler, ErrorMapper, UserFacingError } from '~/lib/error-handling/bff-error'
import type { ServerResponse } from '~/lib/error-handling/types'

import { getUserSession } from '~/lib/kratos/session.server'
import plaid, { isPlaidError } from '~/lib/plaid.server'
import { type PlaidProduct } from '~/lib/usePlaidStore'
import type { Route } from './+types/plaid'

/* ─── action ─────────────────────────────────────────────────────────── */

const PRODUCT_KEYS: PlaidProduct[] = [
  'accounts',
  'auth',
  'balance',
  'identity',
  'transactions'
]

function isPlaidProduct(value: string): value is PlaidProduct {
  return (PRODUCT_KEYS as string[]).includes(value)
}

interface ActionLinkTokenResult {
  intent: 'create_link_token'
  ok: true
  linkToken: string
  expiration: string
}

interface ActionExchangeResult {
  intent: 'exchange'
  ok: true
  itemId: string
  institutionName: string
}

interface ActionExchangeAndLinkResult {
  intent: 'exchange_and_link'
  ok: true
  linkedAccountId: string
  alreadyLinked: boolean
}

interface ActionProductResult {
  intent: 'fetch_product'
  ok: true
  product: PlaidProduct
  response: unknown
}

interface ActionDisconnectResult {
  intent: 'disconnect'
  ok: true
}

export type ActionDataPayload =
  | ActionLinkTokenResult
  | ActionExchangeResult
  | ActionExchangeAndLinkResult
  | ActionProductResult
  | ActionDisconnectResult

export type ActionData = ServerResponse<ActionDataPayload>

export async function action({ request }: Route.ActionArgs): Promise<ActionData | Response> {
  await getUserSession(request)
  const form = await request.formData()
  const intent = String(form.get('intent') || '')

  switch (intent) {
    case 'create_link_token': {
      const result = await plaid.createLinkToken(request)
      if (isPlaidError(result)) {
        return ErrorHandler(request, ErrorMapper.plaid.toUserFacingError(result)) as any
      }
      return {
        success: true,
        data: {
          intent: 'create_link_token',
          ok: true,
          linkToken: result.link_token,
          expiration: result.expiration
        }
      }
    }

    case 'exchange': {
      const publicToken = String(form.get('public_token') || '')
      if (!publicToken) {
        return ErrorHandler(request, UserFacingError('public_token is required', 400)) as any
      }
      const result = await plaid.exchangePublicToken(
        request,
        publicToken
      )
      if (isPlaidError(result)) {
        return ErrorHandler(request, ErrorMapper.plaid.toUserFacingError(result)) as any
      }
      return {
        success: true,
        data: {
          intent: 'exchange',
          ok: true,
          itemId: result.item_id,
          institutionName: result.institution_name
        }
      }
    }

    case 'exchange_and_link': {
      const publicToken = String(form.get('public_token') || '')
      const accountId = String(form.get('account_id') || '')
      if (!publicToken || !accountId) {
        return ErrorHandler(request, UserFacingError('public_token and account_id are required', 400)) as any
      }
      const exchanged = await plaid.exchangePublicToken(request, publicToken)
      if (isPlaidError(exchanged)) {
        return ErrorHandler(request, ErrorMapper.plaid.toUserFacingError(exchanged)) as any
      }
      const linked = await plaid.linkToFiant(request, {
        account_id: accountId,
        account_name: String(form.get('account_name') || '') || undefined,
        account_mask: String(form.get('account_mask') || '') || undefined
      })
      if (isPlaidError(linked)) {
        return ErrorHandler(request, ErrorMapper.plaid.toUserFacingError(linked)) as any
      }
      if (linked.already_linked) {
        return {
          success: true,
          data: {
            intent: 'exchange_and_link' as const,
            ok: true,
            linkedAccountId: linked.linked_account_id,
            alreadyLinked: true
          }
        }
      }
      return redirectWithSnackbar(request, href('/accounts'), {
        message: 'Bank account linked',
        icon: 'check'
      })
    }

    case 'fetch_product': {
      const product = String(form.get('product') || '')
      if (!isPlaidProduct(product)) {
        return ErrorHandler(request, UserFacingError(`unknown product: ${product}`, 400)) as any
      }
      let response: unknown
      switch (product) {
        case 'accounts':
          response = await plaid.getAccounts(request)
          break
        case 'auth':
          response = await plaid.getAuth(request)
          break
        case 'balance':
          response = await plaid.getBalance(request)
          break
        case 'identity':
          response = await plaid.getIdentity(request)
          break
        case 'transactions':
          response = await plaid.getTransactions(request)
          break
      }

      if (isPlaidError(response)) {
        return ErrorHandler(request, ErrorMapper.plaid.toUserFacingError(response)) as any
      }

      return {
        success: true,
        data: {
          intent: 'fetch_product',
          ok: true,
          product,
          response
        }
      }
    }

    case 'disconnect': {
      const result = await plaid.disconnect(request)
      if (isPlaidError(result)) {
        return ErrorHandler(request, ErrorMapper.plaid.toUserFacingError(result)) as any
      }
      return redirectWithSnackbar(request, href('/accounts'), { message: 'Bank disconnected' })
    }

    default:
      return ErrorHandler(request, UserFacingError(`unknown intent: ${intent}`, 400)) as any
  }
}
