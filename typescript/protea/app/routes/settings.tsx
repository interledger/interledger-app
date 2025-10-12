import type { LoaderFunctionArgs, MetaFunction } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { Outlet, useLoaderData, useLocation } from '@remix-run/react'
import { useMemo } from 'react'
import { route } from 'routes-gen'
import type { ApplicationProps } from '~/components'
import {
  Card,
  CardButton,
  CardHeader,
  CardLink,
  CardTitle,
  GridColumn,
  Icon,
  Layouts,
  WalletGrid
} from '~/components'
import { getKycStatus } from '~/data/wallet.server'
import { useTotpChallenge } from '~/lib/hooks/useTotpChallenge'
import { mergeMeta } from '~/lib/meta'
import { KycStatus } from '~/routes/_index/route'

export async function loader({ request }: LoaderFunctionArgs) {
  const url = new URL(request.url)

  const flowId = url.searchParams.get('flow')
  if (flowId) return redirect(`${route('/recovery/password')}`)

  const { kycStatus } = await getKycStatus(request)

  return json({
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

export const meta: MetaFunction = mergeMeta(() => [
  {
    title: 'Settings'
  }
])

export default function Page() {
  const { kycStatus } = useLoaderData<typeof loader>()
  const location = useLocation()
  const pathSegments = location.pathname.split('/').filter(Boolean)

  const callbacks = useMemo(
    () => ({
      onSuccess: () => {
        console.log('✅ TOTP Challenge completed successfully')
      },
      onError: (error) => {
        console.error('❌ TOTP Challenge error:', error)
      }
    }),
    []
  )

  // TOTP Challenge hook for inline verification (no redirects)
  const { popTotp, TotpPopup } = useTotpChallenge(callbacks)

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
            to={route('/settings/grants')}
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
            to={route('/totp/two-factor-authentication')}
          >
            <div className='mr-auto flex space-x-3'>
              <Icon>scan</Icon>
              <span>Two Factor Auth</span>
            </div>
            <Icon>navigate_next</Icon>
          </CardLink>
          <CardButton onClick={popTotp}>Test TOTP Challenge</CardButton>
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

      {/* TOTP Challenge Popup (rendered inline, no redirect) */}
      <TotpPopup />
    </WalletGrid>
  )
}
