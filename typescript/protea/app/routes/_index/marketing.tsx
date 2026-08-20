import {
  Alert,
  AlertBody,
  FiantLogo,
  GatehubLogo,
  Icon,
  InterledgerWalletLogo,
  XagoLogo
} from '~/components'

export function MarketingPage() {
  return (
    <div className='flex grow flex-col items-center justify-center space-y-24 p-4 text-center'>
      <InterledgerWalletLogo className='max-w-lg' />

      <section className='w-full max-w-4xl' aria-label='Service notice'>
        <Alert role='status'>
          <Icon>error</Icon>
          <AlertBody>
            New account registrations and deposits are temporarily unavailable.
          </AlertBody>
        </Alert>
      </section>

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
