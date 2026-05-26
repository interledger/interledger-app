// /plaid — POC route for Plaid Link integration.
//
// Loader fetches the current Plaid link state from the backend so the page
// can show "Connect a bank" vs the linked-bank panel without a flash.
//
// Action dispatches by `intent`:
//   create_link_token  → mint a Plaid Link token (button click)
//   exchange           → exchange the public_token returned by Plaid Link
//   fetch_product      → call /plaid/{accounts|auth|balance|identity|transactions}
//   disconnect         → tear down on Plaid + backend stores
//
// UI components arrive in F5a–F5c; this file currently renders a minimal
// status surface so the route is navigable and the loader/action wiring can
// be exercised. Snackbars wired in F6.

import { useEffect, useState } from 'react'
import {
  Form,
  href,
  redirect,
  useActionData,
  useLoaderData,
  useNavigation
} from 'react-router'

import { jsonWithSnackbar, redirectWithSnackbar } from '~/lib/snackbar.server'
import { ErrorHandler, ErrorMapper, UserFacingError } from '~/lib/error-handling/bff-error'
import type { ServerResponse } from '~/lib/error-handling/types'

import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  Layouts,
  WalletGrid,
  GridColumn,
  OutlineButton,
  PlaidLinkButton,
  EndpointButton,
  ProductCard,
  DebugPanel,
  TextButton
} from '~/components'
import type { ApplicationProps } from '~/components'
import { getUserSession } from '~/lib/kratos/session.server'
import { mergeMeta } from '~/lib/meta'
import plaid, { isPlaidError, type PlaidError, type PlaidState } from '~/lib/plaid.server'
import { usePlaidStore, type PlaidProduct } from '~/lib/usePlaidStore'
import type { Route } from './+types/plaid'

export const handle: ApplicationProps = {
  layout: Layouts.Wallet,
  scaffold: {
    header: { title: 'Plaid (POC)' }
  }
}

export const meta = mergeMeta(() => [{ title: 'Plaid (POC)' }])

interface LoaderData {
  state: PlaidState
  /** Kratos user id (rendered in the debug panel). */
  userId: string
  /** Surfaced on action errors so the component can render them. */
  error?: { message: string; status: number; errorCode: string }
}

export async function loader({ request }: Route.LoaderArgs): Promise<ServerResponse<LoaderData>> {
  const session = await getUserSession(request)
  const userId = session?.identity?.id ?? ''
  const state = await plaid.getState(request)
  if (isPlaidError(state)) {
    const userFacingError = ErrorMapper.plaid.toUserFacingError(state)
    return ErrorHandler(request, userFacingError, {
      cb: () => ({
        success: false as const,
        error: UserFacingError('Plaid state not available, please try again or contact support if the issue persists.')
      })
    }) as any
  }
  return { success: true, data: { state, userId } }
}

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

interface ActionError {
  ok: false
  intent: string
  message: string
  status: number
  errorCode: string
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
      // Already linked → stay on page, inline snackbar via hook
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
      // New link → redirect to /accounts so user sees the row immediately
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
      // Force a fresh load so the loader observes the unlinked state.
      return redirectWithSnackbar(request, href('/plaid'), { message: 'Bank disconnected' })
    }

    default:
      return ErrorHandler(request, UserFacingError(`unknown intent: ${intent}`, 400)) as any
  }
}

/* ─── component ──────────────────────────────────────────────────────── */

export default function PlaidRoute() {
  const loaderData = useLoaderData<typeof loader>()
  const actionResponse = useActionData<typeof action>()
  const actionData = actionResponse?.success ? actionResponse.data : undefined
  const navigation = useNavigation()
  const isSubmitting = navigation.state === 'submitting'

  if (loaderData.success === false) {
    return (
      <WalletGrid>
        <GridColumn className='col-span-full'>
          <Card>
            <CardHeader><CardTitle>Status</CardTitle></CardHeader>
            <CardContent>{loaderData.error?.message}</CardContent>
          </Card>
        </GridColumn>
      </WalletGrid>
    )
  }

  const { state, userId } = loaderData.data

  const {
    itemId,
    institutionName,
    linkedAt,
    activeProduct,
    setLinked,
    clearLinked,
    setLastResponse
  } = usePlaidStore()

  const [showDebug, setShowDebug] = useState(false)

  // Mirror canonical backend state into the Zustand store.
  useEffect(() => {
    if (state.linked && state.item_id) {
      setLinked({
        itemId: state.item_id,
        institutionName: state.institution_name ?? null,
        linkedAt: state.linked_at ?? null
      })
    } else {
      clearLinked()
    }
  }, [state.linked, state.item_id, state.institution_name, state.linked_at, setLinked, clearLinked])

  // Save successful product fetch responses into the Zustand store.
  useEffect(() => {
    if (actionData?.intent === 'fetch_product' && actionData.ok) {
      setLastResponse(actionData.product, actionData.response)
    }
  }, [actionData, setLastResponse])

  return (
    <WalletGrid>
      <GridColumn className='col-span-full'>
        <Card>
          <CardHeader>
            <CardTitle>Status</CardTitle>
          </CardHeader>
          <CardContent>
            {state.linked ? (
              <div className='flex flex-col gap-2'>
                <p>
                  Linked to <strong>{institutionName || 'an institution'}</strong>
                </p>
                <p>Item: <code className='break-all'>{itemId}</code></p>
                {linkedAt && <p>Linked at: <code>{linkedAt}</code></p>}
                <Form method='post' className='mt-2'>
                  <input type='hidden' name='intent' value='disconnect' />
                  <OutlineButton type='submit' disabled={isSubmitting}>
                    Disconnect
                  </OutlineButton>
                </Form>
              </div>
            ) : (
              <div className='flex flex-col gap-3'>
                <p>No bank linked yet.</p>
                <PlaidLinkButton />
              </div>
            )}
          </CardContent>
        </Card>

        {state.linked && (
          <Card>
            <CardHeader>
              <CardTitle>Endpoints</CardTitle>
            </CardHeader>
            <CardContent>
              <div className='grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-5'>
                {PRODUCT_KEYS.map((p) => (
                  <EndpointButton key={p} product={p} />
                ))}
              </div>
            </CardContent>
          </Card>
        )}

        {activeProduct && (
          <ProductCard key={activeProduct} product={activeProduct} />
        )}

        <div className='mt-6 flex justify-end'>
          <TextButton
            onClick={() => setShowDebug((v) => !v)}
            aria-expanded={showDebug}
          >
            {showDebug ? 'Hide debug' : 'Show debug'}
          </TextButton>
        </div>

        {showDebug && (
          <DebugPanel userId={userId} state={state} actionData={actionData} />
        )}
      </GridColumn>
    </WalletGrid>
  )
}
