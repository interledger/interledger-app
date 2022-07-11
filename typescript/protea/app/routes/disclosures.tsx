import { Logo, Router } from '~/components'

export default function Page() {
  return (
    <div className='w-full'>
      {/* Header */}
      <header className='sticky top-0 mx-auto flex h-16 w-full select-none items-center justify-between bg-white p-4 text-medium sm:max-w-lg lg:max-w-3xl xl:max-w-4xl'>
        <div className='flex items-center'>
          <div className='flex items-center justify-start font-display text-2xl font-medium'>
            <Logo className='h-8' />
          </div>
        </div>
      </header>
      {/* Body */}

      <div className='mx-auto grid min-h-[calc(100vh-9rem)] w-full grid-cols-4 content-start gap-4 gap-y-2 overflow-y-auto p-4 pb-24 sm:max-w-lg sm:grid-cols-8 sm:px-0 lg:max-w-3xl lg:grid-cols-12 xl:max-w-4xl'>
        <div className='col-span-full flex flex-col space-y-2 pt-4 pb-8 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
          <span className='pb-8 font-display text-2xl font-medium'>
            Disclosures
          </span>
          <span>
            Fynbos is a financial technology company and is not a bank. Banking
            services provided by Piermont bank; Member FDIC. The Fynbos Visa®
            Debit Card is issued by Piermont bank pursuant to a license from
            Visa U.S.A. Inc. and may be used everywhere Visa debit cards are
            accepted.
          </span>
          <span className='pt-2 font-display text-lg font-medium'>
            SMS fraud alerts
          </span>
          <span>
            Messaging frequency depends on account activity. For more
            information, text HELP to 23618. To cancel fraud text messaging
            services at any time reply STOP to any alert from your mobile
            device.
          </span>
          <span>
            For fraud alerts support, call [insert number] or email [insert
            email]. By giving us your mobile number, you agree that fraud alerts
            text messaging is authorized to notify you of suspected incidents of
            financial or identity fraud. HELP instructions: Text HELP to 23618
            for help or call [insert number] or email [insert email]. Stop
            instructions: Text STOP to 23618 to cancel.
          </span>
          <span>
            Release of Liability: Alerts sent via SMS may not be delivered to
            you if your phone is not in the range of a transmission site, or if
            sufficient network capacity is not available at a particular time.
            Even within coverage, factors beyond the control of wireless
            carriers may interfere with messages delivery for which the carrier
            is not responsible. Carriers do not guarantee that alerts will be
            delivered.
          </span>
          <span className='pt-2 font-display text-lg font-medium'>
            Other disclosures
          </span>
          <span>
            {/* TODO Fix links when making mdx loader */}
            <Router className='text-primary' to={'/privacy-policy'}>
              Privacy Policy
            </Router>
            ,&nbsp;
            <Router className='text-primary' to={'/privacy-policy'}>
              Consent to Electronic Disclosures
            </Router>
            ,&nbsp;
            <Router className='text-primary' to={'/privacy-policy'}>
              Deposit Terms & Conditions
            </Router>
            ,&nbsp;
            <Router className='text-primary' to={'/privacy-policy'}>
              Client Terms of Service
            </Router>
          </span>
          <span className='pt-2 font-display text-lg font-medium'>
            Contact us
          </span>
          <span>
            help@fynbos.dev
            {/* TODO Figure out what needs to go here and make them links */}
            <br />
            +1999999999
          </span>
        </div>
      </div>
    </div>
  )
}
