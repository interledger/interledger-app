import type { LoaderArgs, MetaFunction } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { useLoaderData } from '@remix-run/react'
import { useState } from 'react'
import { route } from 'routes-gen'
import type { ApplicationProps } from '~/components'
import { Card, Icon, Layouts, Router, Snackbar, WalletGrid } from '~/components'
import { getSnackbar } from '~/lib/snackbar.server'
import { getKycStatus } from '~/lib/wallet.server'
import { KycStatus } from '~/routes/_index/route'

export async function loader({ request }: LoaderArgs) {
  const url = new URL(request.url)

  const flowId = url.searchParams.get('flow')
  if (flowId) return redirect(`${route('/recovery/password')}?flow=${flowId}`)

  const { kycStatus } = await getKycStatus(request)

  const snackbar = await getSnackbar(request)

  return json({
    chip: 'Hello',
    snackbar,
    kycStatus
  })
}

export const handle: ApplicationProps = {
  title: 'Settings',
  layout: Layouts.Wallet,
  scaffold: {
    header: {
      title: 'Settings',
      actions: [{ type: 'shapes' }]
    }
  }
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
      <Card className='col-span-full sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <h2 className='text-sm font-medium'>Profile</h2>
        {kycStatus != KycStatus.Unknown && (
          <Router
            to={route('/settings/profile-personal')}
            className='mt-2 flex items-center justify-between rounded-xl bg-nav p-3 text-medium hover:bg-nav-hover'
          >
            <div className='flex space-x-3'>
              <Icon>account_circle</Icon>
              <span>Personal information</span>
            </div>
            <Icon>navigate_next</Icon>
          </Router>
        )}
        <Router
          to={route('/settings/profile-public')}
          className='mt-2 flex items-center justify-between rounded-xl bg-nav p-3 text-medium hover:bg-nav-hover'
        >
          <div className='flex space-x-3'>
            <Icon>contact_page</Icon>
            <span>Public information</span>
          </div>
          <Icon>navigate_next</Icon>
        </Router>
        <Router
          to={route('/settings/profile-contact')}
          className='mt-2 flex items-center justify-between rounded-xl bg-nav p-3 text-medium hover:bg-nav-hover'
        >
          <div className='flex space-x-3'>
            <Icon>call</Icon>
            <span>Contact information</span>
          </div>
          <Icon>navigate_next</Icon>
        </Router>
        <h2 className='mt-6 text-sm font-medium'>Account</h2>
        <Router
          to={route('/settings/linked-accounts')}
          className='mt-2 flex items-center justify-between rounded-xl bg-nav p-3 text-medium hover:bg-nav-hover'
        >
          <div className='flex space-x-3'>
            <Icon>add_card</Icon>
            <span>Linked accounts</span>
          </div>
          <Icon>navigate_next</Icon>
        </Router>
        <Router
          to={route('/settings/linked-identities')}
          className='mt-2 flex items-center justify-between rounded-xl bg-nav p-3 text-medium hover:bg-nav-hover'
        >
          <div className='flex space-x-3'>
            <Icon>add_card</Icon>
            <span>Linked identities</span>
          </div>
          <Icon>navigate_next</Icon>
        </Router>
        <Router
          to={route('/connections')}
          className='mt-2 flex items-center justify-between rounded-xl bg-nav p-3 text-medium hover:bg-nav-hover'
        >
          <div className='flex space-x-3'>
            <Icon>sync</Icon>
            <span>Connections</span>
          </div>
          <Icon>navigate_next</Icon>
        </Router>
        <h2 className='mt-6 text-sm font-medium'>Security</h2>
        <Router
          to={route('/login/challenge')}
          className='mt-2 flex items-center justify-between rounded-xl bg-nav p-3 text-medium hover:bg-nav-hover'
        >
          <div className='flex space-x-3'>
            <Icon>password</Icon>
            <span>Password</span>
          </div>
          <Icon>navigate_next</Icon>
        </Router>
        <Router
          to={route('/legal')}
          className='mt-2 flex items-center justify-between rounded-xl bg-nav p-3 text-medium hover:bg-nav-hover'
        >
          <div className='flex space-x-3'>
            <Icon>policy</Icon>
            <span>Legal &amp; privacy</span>
          </div>
          <Icon>navigate_next</Icon>
        </Router>
        <Router
          to={route('/logout')}
          className='mt-6 flex items-center space-x-3 rounded-xl p-3 text-primary'
        >
          <Icon>logout</Icon>
          <span>Log out</span>
        </Router>
      </Card>
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
