import type { LoaderFunction } from '@remix-run/node'
import { route } from 'routes-gen'
import { Logo, Router } from '~/components'
import { requireUserSession } from '~/lib/kratos.server'

export const loader: LoaderFunction = async ({ request }) => {
  return requireUserSession(request)
}

export default function Page() {
  return (
    <div className='w-full'>
      <header className='sticky top-0 mx-auto flex h-16 w-full select-none items-center justify-between bg-white p-4 text-medium sm:max-w-lg lg:max-w-3xl xl:max-w-4xl'>
        <div className='flex items-center'>
          <div className='flex items-center justify-start font-display text-2xl font-medium'>
            <Logo className='h-8' />
          </div>
        </div>
      </header>
      <div className='mx-auto grid min-h-[calc(100vh-9rem)] w-full grid-cols-4 content-start gap-4 gap-y-2 overflow-y-auto p-4 pb-24 sm:max-w-lg sm:grid-cols-8 sm:px-0 lg:max-w-3xl lg:grid-cols-12 xl:max-w-4xl'>
        <div className='col-span-full flex flex-col space-y-2 pt-4 pb-8 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
          <span className='font-display text-4xl font-medium'>
            Verify identity
          </span>
          <span className='text-medium'>
            In order for us to provide financial services, we are required to
            have a banking partner.
          </span>
          <span className='text-medium'>
            They require some identifying information from you to get started.
          </span>
          <span className='text-medium'>
            You can find out more at the disclosures page linked in the footer
            of this page.
          </span>
        </div>
        <div className='col-span-full flex justify-end pt-4 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
          <Router
            to={route('/flows/:flowId/unit-onboarding/address', {
              flowId: 'init'
            })}
          >
            <div className='flex h-10 items-center rounded-full bg-container-primary px-6 font-display text-sm font-medium text-medium hover:bg-container-primary-hover active:bg-container-primary-active'>
              Get started
            </div>
          </Router>
        </div>
      </div>
    </div>
  )
}
