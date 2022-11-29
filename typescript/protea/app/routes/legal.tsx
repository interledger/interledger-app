import { AnchorRouter, Layouts, Router } from '~/components'
import { route } from 'routes-gen'

export const handle = {
  layout: Layouts.LandingLayout
}

export default function Page() {
  return (
    <main className='w-full overflow-hidden'>
      <section className='relative mx-auto mt-20 grid w-full grid-cols-4 content-start gap-4 gap-y-2 overflow-x-visible px-4 sm:max-w-lg sm:grid-cols-8 sm:px-0 lg:max-w-3xl lg:grid-cols-12 xl:max-w-[59rem]'>
        <div className='col-span-full'>
          <h1 className='flex justify-center font-display text-2xl font-medium lg:text-4xl'>
            Legal agreements for Fynbos
          </h1>
        </div>
        <div className='col-span-full lg:col-span-10 lg:col-start-2 lg:mt-7'>
          <span className='flex justify-center text-center lg:text-2xl'>
            These legal agreements apply to users in Asia Pacific, certain
            European and Latin American countries, Middle East and Africa.
          </span>
        </div>
      </section>
      <section className='relative mx-auto mt-6 grid w-full grid-cols-4 content-start gap-4 gap-y-2 overflow-x-visible px-4 sm:max-w-lg sm:grid-cols-8 sm:px-0 lg:max-w-3xl lg:grid-cols-12 xl:max-w-[59rem]'>
        <div className='col-span-full'>
          <h2 className='flex font-sans text-xl font-medium text-medium'>
            For all users
          </h2>
        </div>
        <div className='col-span-full mt-2'>
          <ul className='flex list-inside list-disc flex-col space-y-4'>
            <li>
              <Router
                className='text-primary'
                to={route('/legal/terms-of-use')}
              >
                Terms of Service
              </Router>
            </li>
            <li>
              <Router
                className='text-primary'
                to={route('/legal/privacy-policy')}
              >
                Privacy Policy
              </Router>
            </li>
            <li>
              <Router
                className='text-primary'
                to={route('/legal/electronic-disclosures')}
              >
                Electronic Disclosures
              </Router>
            </li>
          </ul>
        </div>
      </section>
      <section className='relative mx-auto mt-6 mb-20 grid w-full grid-cols-4 content-start gap-4 gap-y-2 overflow-x-visible px-4 sm:max-w-lg sm:grid-cols-8 sm:px-0 lg:max-w-3xl lg:grid-cols-12 xl:max-w-[59rem]'>
        <div className='col-span-full'>
          <h2 className='flex font-sans text-xl font-medium text-medium'>
            Machnet
          </h2>
          <p className='mt-3 text-sm'>
            The banking services of Fynbos are powered by Machnet. Machnet is a
            financial technology company and not a bank. Banking services are
            provided by Machnet's partner banks who are Member FDIC. Machnet
            provides the Bank services through its banking software provider,
            Synapse.
          </p>
        </div>
        <div className='col-span-full mt-2'>
          <ul className='flex list-inside list-disc flex-col space-y-4'>
            <li>
              <AnchorRouter
                className='text-primary'
                to={
                  'https://machnetservices.com/fynbos-technologies-llc-termsofservice/'
                }
              >
                Terms of Service
              </AnchorRouter>
            </li>
            <li>
              <AnchorRouter
                className='text-primary'
                to={
                  'https://machnetservices.com/fynbos-technologies-llc-privacypolicy/'
                }
              >
                Privacy policy
              </AnchorRouter>
            </li>
          </ul>
        </div>
      </section>
    </main>
  )
}
