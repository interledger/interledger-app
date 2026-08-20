import { useRouteLoaderData } from 'react-router'
import {
  Alert,
  AlertBody,
  FiantLogo,
  GatehubLogo,
  Icon,
  InterledgerWalletLogo,
  XagoLogo
} from '~/components'
import type { RootLoaderData } from '~/root'

export function MarketingPage() {
  const { signupEnabled, depositEnabled } = useRouteLoaderData(
    'root'
  ) as RootLoaderData
  let notice: string | null = null
  if (!signupEnabled && !depositEnabled) {
    notice =
      'New account registrations and deposits are temporarily unavailable.'
  } else if (!signupEnabled) {
    notice = 'New account registrations are temporarily unavailable.'
  } else if (!depositEnabled) {
    notice = 'Deposits are temporarily unavailable.'
  }

  return (
    <div className='flex grow flex-col items-center justify-center space-y-24 p-4 text-center'>
      <InterledgerWalletLogo className='max-w-lg' />

      {notice && (
        <section className='w-full max-w-4xl' aria-label='Service notice'>
          <Alert role='status'>
            <Icon>error</Icon>
            <AlertBody>{notice}</AlertBody>
          </Alert>
        </section>
      )}

      <p className='max-w-4xl text-3xl text-strong'>
        Unlock the potential of Open Payments and Web Monetization through the
        Interledger Wallet and help drive the evolution of digital finances.
      </p>

      <div className='justfy-center hidden space-x-16 md:flex'>
        <FiantLogo className='w-28' />
        <GatehubLogo className='w-28' />
        <XagoLogo className='w-28' />
      </div>
    </div>
  )
}
