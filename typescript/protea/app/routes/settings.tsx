import type { LoaderArgs, MetaFunction } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { Outlet, useLoaderData, useLocation } from '@remix-run/react'
import {useEffect, useState} from 'react'
import { route } from 'routes-gen'
import type { ApplicationProps } from '~/components'
import {
  Card,
  CardHeader,
  CardLink,
  CardTitle,
  GridColumn,
  Icon,
  Layouts,
  Snackbar,
  WalletGrid
} from '~/components'
import { getSnackbar } from '~/lib/snackbar.server'
import { getKycStatus } from '~/lib/wallet.server'
import { KycStatus } from '~/routes/_index/route'

export async function loader({ request }: LoaderArgs) {
  const url = new URL(request.url)

  const flowId = url.searchParams.get('flow')

  const { kycStatus } = await getKycStatus(request)

  const snackbar = await getSnackbar(request)

  return json({
    snackbar,
    kycStatus,
    flowId
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
  const { snackbar, kycStatus, flowId } = useLoaderData<typeof loader>()
  const [showSnackbar, setSnackbar] = useState<boolean>(snackbar.show ?? false)
  const location = useLocation()
  const pathSegments = location.pathname.split('/').filter(Boolean)

  useEffect(() => {
    // change location to password reset
    if (flowId) {
      window.location.href = route('/recovery/password')
    }
  }, [flowId])

  return (
    <WalletGrid>
      <GridColumn
        hideOnMobile={pathSegments[pathSegments.length - 1] !== 'settings'}
        className='col-span-full lg:col-span-5'
      >
        <Card>
          <CardHeader>
            <CardTitle>Profile</CardTitle>
          </CardHeader>
          {kycStatus != KycStatus.Unknown && (
            <CardLink
              end
              preventScrollReset
              prefetch='intent'
              to={route('/settings/profile-personal')}
            >
              <div className='mr-auto flex space-x-3'>
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
            to={route('/settings/profile-public')}
          >
            <div className='mr-auto flex space-x-3'>
              <Icon>contact_page</Icon>
              <span>Public information</span>
            </div>
            <Icon>navigate_next</Icon>
          </CardLink>
          <CardLink
            end
            preventScrollReset
            prefetch='intent'
            to={route('/settings/profile-contact')}
          >
            <div className='mr-auto flex space-x-3'>
              <Icon>call</Icon>
              <span>Contact information</span>
            </div>
            <Icon>navigate_next</Icon>
          </CardLink>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle>Account</CardTitle>
          </CardHeader>
          <CardLink
            end
            preventScrollReset
            prefetch='intent'
            to={route('/settings/keys')}
          >
            <div className='mr-auto flex space-x-3'>
              <Icon>key</Icon>
              <span>Keys</span>
            </div>
            <Icon>navigate_next</Icon>
          </CardLink>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle>Security</CardTitle>
          </CardHeader>
          <CardLink
            end
            preventScrollReset
            prefetch='intent'
            to={route('/login/challenge')}
          >
            <div className='mr-auto flex space-x-3'>
              <Icon>password</Icon>
              <span>Password</span>
            </div>
            <Icon>navigate_next</Icon>
          </CardLink>
          <CardLink
            end
            preventScrollReset
            prefetch='intent'
            to={route('/logout')}
          >
            <div className='mr-auto flex space-x-3'>
              <Icon>logout</Icon>
              <span>Log out</span>
            </div>
            {/*<Icon>open_in_new</Icon>*/}
          </CardLink>
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
