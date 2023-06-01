import type { MetaFunction } from '@remix-run/node'
import { route } from 'routes-gen'
import { Layouts, Router } from '~/components'

export const handle = {
  layout: Layouts.Marketing
}

export const meta: MetaFunction = () => {
  return {
    title: 'Legal'
  }
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
            Some parts of these legal agreements only apply to registered Fynbos
            wallet users, which at the time of publishing, is limited to users
            in the United States of America. All other terms apply to all users
            of Fynbos services.
          </span>
        </div>
      </section>
      <section className='relative mx-auto mb-20 mt-6 grid w-full grid-cols-4 content-start gap-4 gap-y-2 overflow-x-visible px-4 sm:max-w-lg sm:grid-cols-8 sm:px-0 lg:max-w-3xl lg:grid-cols-12 xl:max-w-[59rem]'>
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
                to={route('/legal/terms-of-service')}
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
                to={route('/legal/wallet-license')}
              >
                Wallet license
              </Router>
            </li>
            <li>
              <Router
                className='text-primary'
                to={route('/legal/accessibility-statement')}
              >
                Accessibility Statement
              </Router>
            </li>
          </ul>
        </div>
      </section>
      <section className='relative mx-auto mb-20 mt-6 grid w-full grid-cols-4 content-start gap-4 gap-y-2 overflow-x-visible px-4 sm:max-w-lg sm:grid-cols-8 sm:px-0 lg:max-w-3xl lg:grid-cols-12 xl:max-w-[59rem]'>
        <div className='col-span-full'>
          <h2 className='flex font-sans text-xl font-medium text-medium'>
            Additional terms for US users
          </h2>
          <p className='mt-6 text-sm'>
            Fynbos Technologies LLC (USA) is registered as an agent of Golden
            Money Transfer Inc. a licensed money services business with FINCEN
            MSB Registration Number: 31000231163732. All payments originated
            within the USA are processed under the appropriate state licenses.
          </p>
        </div>
        <div className='col-span-full mt-2'>
          <ul className='flex list-inside list-disc flex-col space-y-4'>
            <li>
              <Router
                className='text-primary'
                to={route('/legal/us/terms-of-use')}
              >
                Terms and Conditions
              </Router>
            </li>
            <li>
              <Router className='text-primary' to={route('/legal/us/licences')}>
                Licences
              </Router>
            </li>
            <li>
              <Router
                className='text-primary'
                to={route('/legal/us/compliance')}
              >
                Compliance Statement
              </Router>
            </li>
            <li>
              <Router
                className='text-primary'
                to={route('/legal/us/privacy-policy')}
              >
                Privacy Policy
              </Router>
            </li>
            <li>
              <Router
                className='text-primary'
                to={route('/legal/us/e-sign-agreement')}
              >
                eSign Agreement
              </Router>
            </li>
          </ul>
        </div>
      </section>
    </main>
  )
}
