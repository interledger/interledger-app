// /connect/plaid/{country} — Phase 2 of the Plaid POC.
//
// Renders the user's Plaid-authorised accounts so they can pick which one to
// register with Fiant as a deposit source. The actual "register" action lands
// in F11; this file owns the loader, the page chrome, and the per-account UI.
//
// Country gate: `params.country` must match the user's wallet country
// (case-insensitive). Phase-2 sandbox is US-only — non-US wallets get 404.

import { useEffect } from 'react'
import {
  Form,
  href,
  redirect,
  useActionData,
  useLoaderData,
  useNavigation
} from 'react-router'

import type { ApplicationProps } from '~/components'
import {
  Button,
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  GridColumn,
  Layouts,
  WalletGrid
} from '~/components'
import { getWalletInfo } from '~/data/wallet.server'
import { mergeMeta } from '~/lib/meta'
import plaid, { isPlaidError } from '~/lib/plaid.server'
import { redirectWithSnackbar } from '~/lib/snackbar.server'
import type { SnackbarAction } from '~/lib/useScaffoldStore'
import { useScaffoldStore } from '~/lib/useScaffoldStore'

import type { Route } from './+types/connect_.plaid_.$country'

/* ─── handle / meta ──────────────────────────────────────────────────── */

export const handle: ApplicationProps = {
  layout: Layouts.Focus,
  scaffold: {
    header: {
      back: href('/'),
      title: 'Link a Plaid bank account'
    }
  }
}

export const meta = mergeMeta(() => [
  { title: 'Link a Plaid bank account' }
])

/* ─── types ──────────────────────────────────────────────────────────── */

interface PlaidAccount {
  account_id: string
  name: string
  official_name?: string
  mask: string | null
  type: string
  subtype: string | null
}

interface AccountsResponse {
  accounts: PlaidAccount[]
}

interface LoaderData {
  country: string
  accounts: PlaidAccount[]
  alreadyLinked: string[]
  /** Surfaced by the F11 action on snackbar-worthy failures. */
  loaderError?: string
}

/* ─── loader ─────────────────────────────────────────────────────────── */

export async function loader({ request, params }: Route.LoaderArgs): Promise<LoaderData> {
  // 1) Resolve the user's wallet and gate the URL country against it.
  const wallet = await getWalletInfo(request)
  const walletCountry = (wallet.country || '').toLowerCase()
  const urlCountry = (params.country || '').toLowerCase()

  // Phase-2 sandbox: US-only. Future locales add to this allow-list.
  if (walletCountry !== 'us' || urlCountry !== walletCountry) {
    throw redirect(href('/'))
  }

  // 2) Parallel fetch: Plaid accounts + already-registered set.
  const [accountsRes, registeredRes] = await Promise.all([
    plaid.getAccounts(request),
    plaid.getRegistered(request)
  ])

  // If Plaid isn't linked yet (Phase 1 incomplete) the user shouldn't have
  // gotten here from F9's guard, but a manual URL visit would 404 the
  // backend's accessToken lookup. Treat both states as "send them back home
  // with a snackbar" — the guard on Home is the right place to recover from.
  if (isPlaidError(accountsRes) || isPlaidError(registeredRes)) {
    throw redirect(href('/'))
  }

  const accounts = ((accountsRes as AccountsResponse).accounts ?? []).map(
    (a): PlaidAccount => ({
      account_id: a.account_id,
      name: a.name,
      official_name: a.official_name,
      mask: a.mask ?? null,
      type: a.type,
      subtype: a.subtype ?? null
    })
  )

  return {
    country: urlCountry,
    accounts,
    alreadyLinked: registeredRes.plaid_account_ids ?? []
  }
}

/* ─── action ─────────────────────────────────────────────────────────── */

/**
 * Action error payload. The snackbar is delivered INLINE (not via the
 * server-flash cookie) because in this app's single-fetch + Redis-session
 * setup the root loader's `getSnackbar` does not reliably read a flash that
 * the same action just wrote — the user otherwise sees the error toast only
 * after a refresh. The component pushes `actionData.snackbar` via a
 * `useEffect` into `useScaffoldStore` on every new action response.
 *
 * Success path stays on `redirectWithSnackbar` — the redirect is followed by
 * a fresh GET cycle so the flash is read on the next request's root loader.
 */
export interface ActionError {
  ok: false
  error: string
  snackbar: {
    id: string
    message: string
    action?: SnackbarAction
    icon: string
  }
}

function actionError(
  error: string,
  snackbar: { message: string; action?: SnackbarAction; icon?: string }
): ActionError {
  return {
    ok: false,
    error,
    snackbar: {
      id: `link-to-fiant-error-${Date.now()}`,
      message: snackbar.message,
      action: snackbar.action,
      icon: snackbar.icon ?? 'close'
    }
  }
}

export async function action({ request, params }: Route.ActionArgs) {
  const form = await request.formData()
  const intent = String(form.get('intent') || '')

  if (intent !== 'link-to-fiant') {
    return actionError(`unknown intent: ${intent}`, { message: 'Unknown action.' })
  }

  const accountID = String(form.get('account_id') || '')
  if (!accountID) {
    return actionError('account_id is required', {
      message: 'Missing Plaid account id.'
    })
  }

  const accountName = String(form.get('account_name') || '')
  const accountMask = String(form.get('account_mask') || '')

  const result = await plaid.linkToFiant(request, {
    account_id: accountID,
    account_name: accountName || undefined,
    account_mask: accountMask || undefined
  })

  if (isPlaidError(result)) {
    // Status-aware error UX. 401 surfaces via the snackbar but the user can
    // re-link from /plaid; 4xx from Plaid (e.g. INVALID_ACCESS_TOKEN) gets a
    // "Set up Plaid" action; everything else is a plain error toast.
    const isReLinkable =
      result.status === 401 ||
      /INVALID_ACCESS_TOKEN|ITEM_LOGIN_REQUIRED|INVALID_PROCESSOR/i.test(
        result.errorCode || ''
      )
    return actionError(result.message, {
      message: `Couldn't link this account: ${result.message}`,
      action: isReLinkable ? 'Set up Plaid' : undefined
    })
  }

  // Success path — soft-success copy for the idempotent dedupe hit.
  const niceName = (accountName || 'Bank account').trim()
  const niceMask = accountMask ? ` ••••${accountMask}` : ''
  const message = result.already_linked
    ? `${niceName}${niceMask} was already linked to Fiant.`
    : `Linked ${niceName}${niceMask} to Fiant.`

  return redirectWithSnackbar(
    request,
    `/connect/plaid/${params.country}`,
    { message, icon: result.already_linked ? 'info' : 'check' }
  )
}

/* ─── component ──────────────────────────────────────────────────────── */

export default function ConnectPlaidCountry() {
  const { accounts, alreadyLinked } = useLoaderData<typeof loader>()
  const actionData = useActionData<typeof action>() as ActionError | undefined
  const navigation = useNavigation()
  const setLoading = useScaffoldStore((s) => s.setLoading)
  const pushSnackbar = useScaffoldStore((s) => s.pushSnackbar)
  const isSubmitting = navigation.state === 'submitting'

  useEffect(() => {
    setLoading(isSubmitting)
    return () => setLoading(false)
  }, [isSubmitting, setLoading])

  // Action errors deliver their snackbar inline (see action() comment). Push
  // it into the store whenever a new error response comes back; the snackbar
  // `id` is unique per response so re-clicks always surface a fresh toast.
  useEffect(() => {
    if (actionData?.ok === false && actionData.snackbar) {
      pushSnackbar(actionData.snackbar)
    }
  }, [actionData, pushSnackbar])

  const alreadyLinkedSet = new Set(alreadyLinked)

  if (accounts.length === 0) {
    return (
      <WalletGrid>
        <GridColumn className='col-span-full'>
          <Card>
            <CardHeader>
              <CardTitle>No Plaid accounts found</CardTitle>
            </CardHeader>
            <CardContent>
              <p className='text-sm text-medium'>
                No accounts are authorised on your linked Plaid item. Re-run
                the Plaid Link flow and pick the accounts you want to expose.
              </p>
            </CardContent>
          </Card>
        </GridColumn>
      </WalletGrid>
    )
  }

  const allLinked = accounts.every((a) => alreadyLinkedSet.has(a.account_id))

  return (
    <WalletGrid>
      <GridColumn className='col-span-full'>
        <Card>
          <CardContent>
            <p className='text-sm text-medium'>
              Pick a Plaid-authorised account to register with Fiant as a
              deposit source. You can link multiple accounts one at a time.
            </p>
          </CardContent>
        </Card>

        {allLinked && (
          <Card>
            <CardContent>
              <p className='text-sm text-medium'>
                All Plaid accounts are already linked to Fiant.
              </p>
            </CardContent>
          </Card>
        )}

        {accounts.map((acc) => {
          const isLinked = alreadyLinkedSet.has(acc.account_id)
          const subtitle = [
            acc.mask ? `••••${acc.mask}` : null,
            acc.subtype ? acc.subtype : acc.type
          ]
            .filter(Boolean)
            .join(' · ')

          return (
            <Card key={acc.account_id}>
              <CardHeader>
                <CardTitle>{acc.official_name || acc.name}</CardTitle>
              </CardHeader>
              <CardContent>
                <p className='text-sm text-medium'>{subtitle}</p>
                <Form method='post' className='mt-4'>
                  <input type='hidden' name='intent' value='link-to-fiant' />
                  <input type='hidden' name='account_id' value={acc.account_id} />
                  <input type='hidden' name='account_name' value={acc.name} />
                  <input
                    type='hidden'
                    name='account_mask'
                    value={acc.mask ?? ''}
                  />
                  <Button
                    type='submit'
                    disabled={isLinked || isSubmitting}
                    shrink
                  >
                    {isLinked ? 'Linked' : 'Use for deposits'}
                  </Button>
                </Form>
              </CardContent>
            </Card>
          )
        })}
      </GridColumn>
    </WalletGrid>
  )
}
