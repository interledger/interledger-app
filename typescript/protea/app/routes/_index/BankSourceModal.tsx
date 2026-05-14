import type { Dispatch, SetStateAction } from 'react'
import { useState } from 'react'
import { href, useNavigate } from 'react-router'

import {
  CardContent,
  CardHeader,
  OutlineButton,
  TextButton
} from '~/components'
import { Dialog } from '~/components/Dialog'
import { useScaffoldStore } from '~/lib/useScaffoldStore'

type Props = {
  open: boolean
  setOpen: Dispatch<SetStateAction<boolean>>
  /** Uppercase ISO-3166-1 alpha-2 code from the user's wallet (e.g. 'US' / 'ZA'). */
  country: string
}

/** Subset of GET /api/plaid/state we care about for the guard. */
interface PlaidStateLite {
  linked: boolean
}

/**
 * BankSourceModal lets the user pick how they want to link a bank account:
 *
 *   - Enter bank manually → existing `/connect/bank/{country}` route. Routing
 *     forks on `country` because the two manual flows (US ACH vs ZA) live in
 *     separate route files today; expand the switch as new countries land.
 *
 *   - Link via Plaid → Phase 2 of the Plaid POC. Click triggers an in-browser
 *     guard fetch to `/api/plaid/state` (Traefik routes `/api/plaid/*` straight
 *     to the Go backend; Kratos cookie is same-origin). If the user already
 *     completed Phase 1 (`linked: true`) we navigate to
 *     `/connect/plaid/{country}` (F10 lands that route). Otherwise we push a
 *     snackbar with a "Set up Plaid" action that links to `/plaid`.
 */
export function BankSourceModal({ open, setOpen, country }: Props) {
  const navigate = useNavigate()
  const pushSnackbar = useScaffoldStore((s) => s.pushSnackbar)
  const [checkingPlaid, setCheckingPlaid] = useState(false)

  const manualPath =
    country === 'US' ? href('/connect/bank/us') : href('/connect/bank/za')

  const close = () => setOpen(false)

  const goManual = () => {
    close()
    navigate(manualPath)
  }

  const goPlaid = async () => {
    if (checkingPlaid) return
    setCheckingPlaid(true)
    try {
      const res = await fetch('/api/plaid/state', {
        credentials: 'same-origin',
        headers: { Accept: 'application/json' }
      })
      if (!res.ok) {
        // Likely 401 (cookie expired) or 5xx. Surface as a snackbar without
        // the "Set up Plaid" action — that CTA wouldn't fix an auth/server bug.
        pushSnackbar({
          id: `plaid-state-error-${Date.now()}`,
          message: `Couldn't check Plaid status (HTTP ${res.status}). Please try again.`,
          icon: 'close'
        })
        return
      }
      const state: PlaidStateLite = await res.json()
      close()
      if (state.linked) {
        navigate(`/connect/plaid/${country.toLowerCase()}`)
      } else {
        pushSnackbar({
          id: `plaid-not-linked-${Date.now()}`,
          message: 'Set up Plaid first to link a bank account this way.',
          action: 'Set up Plaid',
          icon: 'info'
        })
      }
    } catch (err) {
      pushSnackbar({
        id: `plaid-state-error-${Date.now()}`,
        message: `Couldn't reach the Plaid service: ${
          err instanceof Error ? err.message : 'unknown error'
        }`,
        icon: 'close'
      })
    } finally {
      setCheckingPlaid(false)
    }
  }

  return (
    <Dialog open={open} setOpen={setOpen}>
      <CardHeader>
        <h1 className='text-xl font-medium'>Connect a bank account</h1>
      </CardHeader>
      <CardContent>
        <p className='mb-4 text-sm text-medium'>
          Pick how you want to link a bank account to your wallet.
        </p>
        <div className='flex flex-col gap-3'>
          <OutlineButton onClick={goManual} disabled={checkingPlaid}>
            Enter bank manually
          </OutlineButton>
          <OutlineButton onClick={goPlaid} disabled={checkingPlaid}>
            {checkingPlaid ? 'Checking…' : 'Link via Plaid'}
          </OutlineButton>
          <TextButton onClick={close} disabled={checkingPlaid}>
            Cancel
          </TextButton>
        </div>
      </CardContent>
    </Dialog>
  )
}
