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

import { useEffect } from 'react'
import {
  Form,
  href,
  redirect,
  useActionData,
  useLoaderData,
  useNavigation
} from 'react-router'

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
  ProductCard
} from '~/components'
import type { ApplicationProps } from '~/components'
import { getUserSession } from '~/lib/kratos/session.server'
import { mergeMeta } from '~/lib/meta'
import plaid, { PlaidError, type PlaidState } from '~/lib/plaid.server'
import { usePlaidStore, type PlaidProduct } from '~/lib/usePlaidStore'

import type { Route } from './+types/plaid'

/* ─── handle / meta ──────────────────────────────────────────────────── */

export const handle: ApplicationProps = {
  layout: Layouts.Wallet,
  scaffold: {
    header: { title: 'Plaid (POC)' }
  }
}

export const meta = mergeMeta(() => [{ title: 'Plaid (POC)' }])

/* ─── loader ─────────────────────────────────────────────────────────── */

interface LoaderData {
  state: PlaidState
  /** Kratos user id (rendered in the debug panel). */
  userId: string
  /** Surfaced on action errors so the component can render them. */
  error?: { message: string; status: number; errorCode: string }
}

export async function loader({ request }: Route.LoaderArgs): Promise<LoaderData> {
  const session = await getUserSession(request)
  const userId = session?.identity?.id ?? ''
  const state = await plaid.getState(request)
  return { state, userId }
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

export type ActionData =
  | ActionLinkTokenResult
  | ActionExchangeResult
  | ActionProductResult
  | ActionDisconnectResult
  | ActionError

export async function action({ request }: Route.ActionArgs): Promise<ActionData | Response> {
  await getUserSession(request)
  const form = await request.formData()
  const intent = String(form.get('intent') || '')

  try {
    switch (intent) {
      case 'create_link_token': {
        const { link_token, expiration } = await plaid.createLinkToken(request)
        return {
          intent: 'create_link_token',
          ok: true,
          linkToken: link_token,
          expiration
        }
      }

      case 'exchange': {
        const publicToken = String(form.get('public_token') || '')
        if (!publicToken) {
          return {
            ok: false,
            intent,
            message: 'public_token is required',
            status: 400,
            errorCode: 'BAD_REQUEST'
          }
        }
        const { item_id, institution_name } = await plaid.exchangePublicToken(
          request,
          publicToken
        )
        return {
          intent: 'exchange',
          ok: true,
          itemId: item_id,
          institutionName: institution_name
        }
      }

      case 'fetch_product': {
        const product = String(form.get('product') || '')
        if (!isPlaidProduct(product)) {
          return {
            ok: false,
            intent,
            message: `unknown product: ${product}`,
            status: 400,
            errorCode: 'BAD_REQUEST'
          }
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
        return {
          intent: 'fetch_product',
          ok: true,
          product,
          response
        }
      }

      case 'disconnect': {
        await plaid.disconnect(request)
        // Force a fresh load so the loader observes the unlinked state.
        return redirect(href('/plaid'))
      }

      default:
        return {
          ok: false,
          intent,
          message: `unknown intent: ${intent}`,
          status: 400,
          errorCode: 'BAD_REQUEST'
        }
    }
  } catch (err) {
    if (err instanceof PlaidError) {
      return {
        ok: false,
        intent,
        message: err.message,
        status: err.status,
        errorCode: err.errorCode
      }
    }
    throw err
  }
}

/* ─── component ──────────────────────────────────────────────────────── */

export default function PlaidRoute() {
  const { state, userId } = useLoaderData<typeof loader>()
  const actionData = useActionData<typeof action>()
  const navigation = useNavigation()
  const isSubmitting = navigation.state === 'submitting'

  const {
    itemId,
    institutionName,
    linkedAt,
    linkToken,
    lastError,
    lastResponses,
    setLinked,
    clearLinked,
    setLastResponse
  } = usePlaidStore()

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

        {PRODUCT_KEYS.map((p) => (
          <ProductCard key={p} product={p} />
        ))}

        {linkToken && (
          <Card>
            <CardHeader>
              <CardTitle>Pending link</CardTitle>
            </CardHeader>
            <CardContent>
              <p>link_token issued; F5a will hand it to react-plaid-link.</p>
              <code className='block break-all text-xs'>{linkToken}</code>
            </CardContent>
          </Card>
        )}

        {lastError && (
          <Card>
            <CardHeader>
              <CardTitle>Last error</CardTitle>
            </CardHeader>
            <CardContent>
              <code className='break-all'>{lastError}</code>
            </CardContent>
          </Card>
        )}

        {actionData && (
          <Card>
            <CardHeader>
              <CardTitle>Debug — last action result</CardTitle>
            </CardHeader>
            <CardContent>
              <pre className='overflow-x-auto whitespace-pre-wrap break-all text-xs'>
                {JSON.stringify(actionData, null, 2)}
              </pre>
            </CardContent>
          </Card>
        )}

        <Card>
          <CardHeader>
            <CardTitle>Debug — session</CardTitle>
          </CardHeader>
          <CardContent>
            <p>
              user_id: <code className='break-all'>{userId || '(unauthenticated)'}</code>
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Debug — loader state</CardTitle>
          </CardHeader>
          <CardContent>
            <pre className='overflow-x-auto whitespace-pre-wrap break-all text-xs'>
              {JSON.stringify(state, null, 2)}
            </pre>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Debug — cached product responses</CardTitle>
          </CardHeader>
          <CardContent>
            <pre className='overflow-x-auto whitespace-pre-wrap break-all text-xs'>
              {JSON.stringify(lastResponses, null, 2)}
            </pre>
          </CardContent>
        </Card>
      </GridColumn>
    </WalletGrid>
  )
}
