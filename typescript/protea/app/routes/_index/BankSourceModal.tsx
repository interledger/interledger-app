import type { Dispatch, SetStateAction } from 'react'
import { href, useNavigate } from 'react-router'

import {
  CardContent,
  CardHeader,
  OutlineButton,
  TextButton
} from '~/components'
import { Dialog } from '~/components/Dialog'

type Props = {
  open: boolean
  setOpen: Dispatch<SetStateAction<boolean>>
  /** Uppercase ISO-3166-1 alpha-2 code from the user's wallet (e.g. 'US' / 'ZA'). */
  country: string
}

/**
 * BankSourceModal lets the user pick how they want to link a bank account:
 *
 *   - Enter bank manually → existing `/connect/bank/{country}` route. Routing
 *     forks on `country` because the two manual flows (US ACH vs ZA) live in
 *     separate route files today; expand the switch as new countries land.
 *
 *   - Link via Plaid → Phase 2 of the Plaid POC. F8 wires the modal shell + a
 *     direct redirect to `/plaid` (so the user can complete the Plaid setup if
 *     they haven't yet). F9 will replace the direct navigate with a guard that
 *     hits `/api/plaid/state` and either redirects to `/connect/plaid/{country}`
 *     (when linked) or surfaces a "set up Plaid first" snackbar that points at
 *     `/plaid`.
 */
export function BankSourceModal({ open, setOpen, country }: Props) {
  const navigate = useNavigate()
  const manualPath =
    country === 'US' ? href('/connect/bank/us') : href('/connect/bank/za')

  const close = () => setOpen(false)
  const goManual = () => {
    close()
    navigate(manualPath)
  }
  // F9 replaces this with the /api/plaid/state guard.
  const goPlaid = () => {
    close()
    navigate(href('/plaid'))
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
          <OutlineButton onClick={goManual}>Enter bank manually</OutlineButton>
          <OutlineButton onClick={goPlaid}>Link via Plaid</OutlineButton>
          <TextButton onClick={close}>Cancel</TextButton>
        </div>
      </CardContent>
    </Dialog>
  )
}
