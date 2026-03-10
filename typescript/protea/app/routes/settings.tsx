import type { Route } from './+types/settings'
import { data, redirect } from 'react-router';
import { Outlet, useLoaderData, useLocation } from 'react-router';
import { href } from 'react-router'
import type { ApplicationProps } from '~/components'
import {
  Card,
  CardHeader,
  CardLink,
  CardTitle,
  GridColumn,
  Icon,
  Layouts,
  WalletGrid
} from '~/components'
import { getKycStatus } from '~/data/wallet.server'
import { mergeMeta } from '~/lib/meta'
import { KycStatus } from '~/lib/types'

export async function loader({ request }: Route.LoaderArgs) {
  const url = new URL(request.url)

  const flowId = url.searchParams.get('flow')
  if (flowId) return redirect(`${href('/recovery/password')}?flow=${flowId}`)

  const { kycStatus } = await getKycStatus(request)

  return data({
    kycStatus
  })
}

export const handle: ApplicationProps = {
  layout: Layouts.Wallet,
  scaffold: {
    header: {
      title: 'Settings'
    }
  }
}

export const meta = mergeMeta(() => [
  {
    title: 'Settings'
  }
])

export default function Page() {
  const { kycStatus } = useLoaderData()
  const location = useLocation()
  const pathSegments = location.pathname.split('/').filter(Boolean)

  return (
    <WalletGrid>
      <GridColumn
        hideOnMobile={pathSegments[pathSegments.length - 1] !== 'settings'}
        className='col-span-full lg:col-span-6'
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
              to={href('/settings/profile-personal')}
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
            to={href('/settings/profile-public')}
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
            to={href('/settings/profile-contact')}
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
            to={href('/settings/grants')}
          >
            <div className='mr-auto flex space-x-3'>
              <Icon>request_quote</Icon>
              <span>Grants</span>
            </div>
            <Icon>navigate_next</Icon>
          </CardLink>
          <CardLink
            end
            preventScrollReset
            prefetch='intent'
            to={href('/settings/keys')}
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
            to={href('/totp/two-factor-authentication')}
          >
            <div className='mr-auto flex space-x-3'>
              <Icon>scan</Icon>
              <span>Two Factor Auth</span>
            </div>
            <Icon>navigate_next</Icon>
          </CardLink>
          <CardLink
            end
            preventScrollReset
            prefetch='intent'
            to={href('/login/challenge')}
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
            to={href('/logout')}
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
    </WalletGrid>
  )
}
