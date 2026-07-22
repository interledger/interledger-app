import { useEffect, useRef } from 'react'
import { href, redirect, useLoaderData, useNavigate } from 'react-router'
import type { ApplicationProps } from '~/components'
import { Button, Card, CardContent, Layouts } from '~/components'
import { usePlaidLinkFlow } from '~/components/Plaid/usePlaidLinkFlow'
import { getFeatures, getWalletInfo } from '~/data/wallet.server'
import { envBool } from '~/env.server'
import { jsonWithCSRF } from '~/lib/csrf.server'
import { isConnectError } from '~/lib/error.server'
import { grpc } from '~/lib/grpc.server'
import { mergeMeta } from '~/lib/meta'
import type { Route } from './+types/connect_.bank'

export async function loader({ request }: Route.LoaderArgs) {
  // Gate 0: Plaid must be the active bank-link path; otherwise this page is dead.
  if (!envBool('PLAID_ENABLED')) throw redirect(href('/'))

  // Gate 1: bank linking must be enabled (same flag as the Home card).
  const features = await getFeatures(request)
  if (!features.banksEnabled) throw redirect(href('/'))

  // Gate 2: Plaid is US-only. Non-US users are never eligible — send Home.
  const walletInfo = await getWalletInfo(request)
  if (walletInfo.country !== 'US') throw redirect(href('/'))

  // Gate 3: the US ("USD") balance is provisioned asynchronously after KYC
  // approval. When it doesn't exist yet the user is eligible but not activated:
  // render an informative "still setting up" state instead of silently bouncing
  // them Home, and do NOT auto-launch Plaid / mint a link token (see Page).
  const balancesResponse = await grpc.getBalances(request, {})
  if (isConnectError(balancesResponse)) throw redirect(href('/'))
  const activated = balancesResponse.balances.some(
    (bal) => bal.countryCode === 'US'
  )

  return jsonWithCSRF(request, { activated })
}

export const handle: ApplicationProps = {
  layout: Layouts.Focus,
  scaffold: {
    header: {
      back: href('/'),
      title: 'Connect bank account'
    }
  }
}

export const meta = mergeMeta(() => [
  {
    title: 'Connect bank account'
  }
])

export default function Page() {
  const { csrfToken, activated } = useLoaderData<typeof loader>()
  const navigate = useNavigate()
  const { connect, busy, scriptError } = usePlaidLinkFlow({
    csrfToken,
    // User dismissed the Plaid overlay — leave the dedicated page.
    onCancel: () => navigate(href('/'))
  })

  // Auto-launch the Plaid flow once on mount — but only when the wallet is
  // activated. The ref guard prevents a second token mint under React
  // StrictMode's double-invoke in development.
  const launched = useRef(false)
  useEffect(() => {
    if (!activated) return
    if (launched.current) return
    launched.current = true
    connect()
  }, [connect, activated])

  // Eligible (US, Plaid + banks on) but the USD balance isn't provisioned yet.
  if (!activated) {
    return (
      <Card>
        <CardContent>
          <div className='flex flex-col gap-4'>
            <p>
              Your account is still being set up. Please try again shortly.
            </p>
            <Button type='button' onClick={() => navigate(href('/'))}>
              Back to home
            </Button>
          </div>
        </CardContent>
      </Card>
    )
  }

  return (
    <Card>
      <CardContent>
        {scriptError ? (
          <div className='flex flex-col gap-4'>
            <p>We couldn’t start the bank connection. Please try again.</p>
            <Button type='button' onClick={connect}>
              Try again
            </Button>
          </div>
        ) : (
          <p>{busy ? 'Connecting you to your bank…' : 'Opening your bank…'}</p>
        )}
      </CardContent>
    </Card>
  )
}
