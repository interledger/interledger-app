import type { LoaderArgs, MetaFunction } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { Outlet, useLoaderData, useLocation } from '@remix-run/react'
import { useState } from 'react'
import { route } from 'routes-gen'
import type { ApplicationProps } from '~/components'
import {
  Card,
  CardColumn,
  CardLink,
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
          <CardColumn>
            {kycStatus != KycStatus.Unknown && (
              <CardLink
                end
                preventScrollReset
                prefetch='intent'
                className='items-center justify-between'
                to={route('/settings/profile-personal')}
              >
                <div className='flex space-x-3'>
                  <Icon>account_circle</Icon>
                  <span>Personal information</span>
                </div>
                <Icon>navigate_next</Icon>
              </CardLink>
            )}
            <CardLink
              end
              preventScrollReset
              prefetch='intent'
              className='items-center justify-between'
              to={route('/settings/profile-public')}
            >
              <div className='flex space-x-3'>
                <Icon>contact_page</Icon>
                <span>Public information</span>
              </div>
              <Icon>navigate_next</Icon>
            </CardLink>
            <CardLink
              end
              preventScrollReset
              prefetch='intent'
              className='items-center justify-between'
              to={route('/settings/profile-contact')}
            >
              <div className='flex space-x-3'>
                <Icon>call</Icon>
                <span>Contact information</span>
              </div>
              <Icon>navigate_next</Icon>
            </CardLink>
          </CardColumn>
        </Card>
        <Card>
          <CardTitle>Account</CardTitle>
          <CardColumn>
            <CardLink
              end
              preventScrollReset
              prefetch='intent'
              className='items-center justify-between'
              to={route('/settings/keys')}
            >
              <div className='flex space-x-3'>
                <Icon>key</Icon>
                <span>Keys</span>
              </div>
              <Icon>navigate_next</Icon>
            </CardLink>
          </CardColumn>
        </Card>
        <Card>
          <CardTitle>Security</CardTitle>
          <CardColumn>
            <CardLink
              end
              preventScrollReset
              prefetch='intent'
              className='items-center justify-between'
              to={route('/login/challenge')}
            >
              <div className='flex space-x-3'>
                <Icon>password</Icon>
                <span>Password</span>
              </div>
              <Icon>navigate_next</Icon>
            </CardLink>
            <CardLink
              end
              preventScrollReset
              prefetch='intent'
              className='items-center justify-between'
              to={route('/logout')}
            >
              <div className='flex space-x-3'>
                <Icon>logout</Icon>
                <span>Log out</span>
              </div>
              {/*<Icon>open_in_new</Icon>*/}
            </CardLink>
          </CardColumn>
        </Card>
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
