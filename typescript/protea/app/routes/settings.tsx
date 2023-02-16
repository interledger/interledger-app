import { useState } from 'react'
import type { LoaderArgs , MetaFunction} from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { useLoaderData } from '@remix-run/react'
import { route } from 'routes-gen'
import {
  HomeShapes,
  Icon,
  Layouts,
  Router,
  Snackbar,
  WalletGrid
} from '~/components'
import { getSnackbar } from '~/lib/snackbar.server'
import { getKycStatus } from '~/lib/wallet.server'
import { KycStatus } from '~/routes/index'

export async function loader({ request }: LoaderArgs) {
  const url = new URL(request.url)

  const flowId = url.searchParams.get('flow')
  if (flowId) return redirect(`${route('/recovery/password')}?flow=${flowId}`)

  const { kycStatus } = await getKycStatus(request)

  const snackbar = await getSnackbar(request)

  return json({
    snackbar,
    kycStatus
  })
}

export const handle = {
  layout: Layouts.WalletLayout
}

export const meta: MetaFunction = () => {
  return {
    title: 'Settings'
  }
}

export default function Page() {
  const { snackbar, kycStatus } = useLoaderData<typeof loader>()
  const [showSnackbar, setSnackbar] = useState<boolean>(snackbar.show ?? false)
  return (
    <WalletGrid>
      <div className='col-span-full flex flex-col rounded-2xl bg-page p-4 pb-8 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <div className='mt-2'>
          <HomeShapes />
        </div>
        <h1 className='mt-6 font-display text-2xl font-medium'>Settings</h1>
        <h2 className='mt-6 text-sm font-medium'>Profile</h2>
        <Router
          to={route('/settings/personal-details')}
          className='mt-2 flex items-center justify-between rounded-xl bg-container p-3 text-medium hover:bg-container-hover'
        >
          <div className='flex space-x-3'>
            <Icon>face</Icon>
            <span>Personal details</span>
          </div>
          <Icon>navigate_next</Icon>
        </Router>
        <h2 className='mt-6 text-sm font-medium'>Account</h2>
        <Router
          to={route('/settings/linked-accounts')}
          className='mt-2 flex items-center justify-between rounded-xl bg-container p-3 text-medium hover:bg-container-hover'
        >
          <div className='flex space-x-3'>
            <Icon>add_card</Icon>
            <span>Linked accounts</span>
          </div>
          <Icon>navigate_next</Icon>
        </Router>
        {kycStatus == KycStatus.Verified && (
          <Router
            to={route('/settings/account-limits')}
            className='mt-2 flex items-center justify-between rounded-xl bg-container p-3 text-medium hover:bg-container-hover'
          >
            <div className='flex space-x-3'>
              <Icon>tune</Icon>
              <span>Account limits</span>
            </div>
            <Icon>navigate_next</Icon>
          </Router>
        )}
        <h2 className='mt-6 text-sm font-medium'>Security</h2>
        <Router
          to={route('/login/challenge')}
          className='mt-2 flex items-center justify-between rounded-xl bg-container p-3 text-medium hover:bg-container-hover'
        >
          <div className='flex space-x-3'>
            <Icon>password</Icon>
            <span>Password</span>
          </div>
          <Icon>navigate_next</Icon>
        </Router>
        <Router
          to={route('/legal')}
          className='mt-2 flex items-center justify-between rounded-xl bg-container p-3 text-medium hover:bg-container-hover'
        >
          <div className='flex space-x-3'>
            <Icon>policy</Icon>
            <span>Legal &amp; privacy</span>
          </div>
          <Icon>navigate_next</Icon>
        </Router>
      </div>
      <Snackbar
        message={snackbar.message}
        action={snackbar.action}
        icon={snackbar.icon}
        show={showSnackbar}
        id='cookie-snackbar'
        dismissAfter={3000}
        offset
        onClose={() => setSnackbar(false)}
      />
    </WalletGrid>
  )
}
