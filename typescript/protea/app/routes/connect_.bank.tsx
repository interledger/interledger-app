import { useEffect, useRef } from 'react'
import { href, redirect, useNavigate } from 'react-router'
import type { ApplicationProps } from '~/components'
import { Button, Card, CardContent, Layouts } from '~/components'
import { usePlaidLinkFlow } from '~/components/Plaid/usePlaidLinkFlow'
import { getFeatures } from '~/data/wallet.server'
import { isConnectError } from '~/lib/error.server'
import { grpc } from '~/lib/grpc.server'
import { mergeMeta } from '~/lib/meta'
import type { Route } from './+types/connect_.bank'

export async function loader({ request }: Route.LoaderArgs) {
  // Gate 1: bank linking must be enabled (same flag as the Home card).
  const features = await getFeatures(request)
  if (!features.banksEnabled) throw redirect(href('/'))

  // Gate 2: US-only — user must hold a US balance (mirrors connect_.bank_.us).
  const balancesResponse = await grpc.getBalances(request, {})
  if (
    isConnectError(balancesResponse) ||
    !balancesResponse.balances.some((bal) => bal.countryCode === 'US')
  )
    throw redirect(href('/'))

  return null
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
  const navigate = useNavigate()
  const { connect, busy, scriptError } = usePlaidLinkFlow({
    // User dismissed the Plaid overlay — leave the dedicated page.
    onCancel: () => navigate(href('/'))
  })

  // Auto-launch the Plaid flow once on mount. The ref guard prevents a second
  // token mint under React StrictMode's double-invoke in development.
  const launched = useRef(false)
  useEffect(() => {
    if (launched.current) return
    launched.current = true
    connect()
  }, [connect])

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
