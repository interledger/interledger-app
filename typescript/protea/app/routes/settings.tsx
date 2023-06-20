import type { LoaderArgs, MetaFunction } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { NavLink, Outlet, useLoaderData, useLocation } from '@remix-run/react'
import clsx from 'clsx'
import { useState } from 'react'
import { route } from 'routes-gen'
import type { ApplicationProps } from '~/components'
import {
  Card,
  CardTitle,
  GridColumn,
  Icon,
  Layouts,
  Router,
  Snackbar,
  WalletGrid
} from '~/components'
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
  const location = useLocation()
  const pathSegments = location.pathname.split('/').filter(Boolean)
  console.log(pathSegments[pathSegments.length - 1])
  return (
    <WalletGrid>
      <GridColumn
        hideOnMobile={pathSegments[pathSegments.length - 1] !== 'settings'}
        className='col-span-full lg:col-span-5'
      >
        <Card>
          <CardTitle>Profile</CardTitle>
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
          <NavLink
            end
            preventScrollReset={true}
            prefetch='intent'
            className='group relative mt-2 flex w-full items-center justify-between rounded-xl py-3 focus-visible:outline-2 focus-visible:outline-focus'
            to={route('/settings/profile-public')}
          >
            {({ isActive }) => (
              <>
                <div className='z-10 flex space-x-3'>
                  <Icon>contact_page</Icon>
                  <span>Public information</span>
                </div>
                <Icon>navigate_next</Icon>
                <div
                  className={clsx(
                    'absolute -inset-x-2 inset-y-0 rounded-xl',
                    isActive ? 'bg-nav-hover' : 'group-hover:bg-nav'
                  )}
                />
              </>
            )}
          </NavLink>
          {/* TODO: do the rest of these */}
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
        </Card>
        <Card>
          <CardTitle>Account</CardTitle>
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
        </Card>
        <Card>
          <CardTitle>Security</CardTitle>
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
        </Card>
        <Router
          to={route('/logout')}
          className='flex items-center space-x-3 rounded-xl p-3 text-primary'
        >
          <Icon>logout</Icon>
          <span>Log out</span>
        </Router>
      </GridColumn>
      <GridColumn className='col-span-full lg:col-span-6 lg:col-start-7'>
        <Outlet />
      </GridColumn>
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
