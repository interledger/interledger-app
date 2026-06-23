import { href } from 'react-router'

import { redirectWithSnackbar } from '~/lib/snackbar.server'
import { ErrorHandler, ErrorMapper, UserFacingError } from '~/lib/error-handling/bff-error'
import type { ServerResponse } from '~/lib/error-handling/types'

import { getUserSession } from '~/lib/kratos/session.server'
import plaid, { isPlaidError } from '~/lib/plaid.server'
import type { Route } from './+types/plaid'

/* ─── action ─────────────────────────────────────────────────────────── */

interface ActionLinkTokenResult {
  intent: 'create_link_token'
  ok: true
  linkToken: string
  expiration: string
}

interface ActionExchangeAndLinkResult {
  intent: 'exchange_and_link'
  ok: true
  linkedAccountId: string
  alreadyLinked: boolean
}

export type ActionDataPayload =
  | ActionLinkTokenResult
  | ActionExchangeAndLinkResult

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

    case 'exchange_and_link': {
      const publicToken = String(form.get('public_token') || '')
      const accountId = String(form.get('account_id') || '')
      if (!publicToken || !accountId) {
        return ErrorHandler(request, UserFacingError('public_token and account_id are required', 400)) as any
      }
      const linked = await plaid.linkToFiant(request, {
        public_token: publicToken,
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

    default:
      return ErrorHandler(request, UserFacingError(`unknown intent: ${intent}`, 400)) as any
  }
}
