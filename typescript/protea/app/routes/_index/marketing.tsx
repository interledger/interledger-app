import {
  ChimoneyLogo,
  FiantLogo,
  GatehubLogo,
  InterledgerLogo,
  XagoLogo
} from '~/components'

export function MarketingPage() {
  return (
    <div className='flex grow flex-col items-center justify-center space-y-24 p-4 text-center'>
      <InterledgerLogo className='max-w-lg' />

      <p className='max-w-4xl text-3xl text-strong'>
        Unlock the potential of Open Payments and Web Monetization through the
        Interledger Wallet and help drive the evolution of digital finances.
      </p>

      <div className='justfy-center hidden space-x-16 md:flex'>
        <FiantLogo className='w-28' />
        <GatehubLogo className='w-28' />
        <ChimoneyLogo className='w-28' />
        <XagoLogo className='w-28' />
      </div>
    </div>
  )
}
