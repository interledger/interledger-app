import type { LoaderArgs, MetaFunction } from '@remix-run/node'
import { json } from '@remix-run/node'
import { Outlet, useLoaderData, useLocation } from '@remix-run/react'
import { useState } from 'react'
import { route } from 'routes-gen'
import type { ApplicationProps } from '~/components'
import {
  Card,
  CardContent,
  CardHeader,
  CardIcon,
  CardLink,
  CardTitle,
  Chip,
  ChipColor,
  Fab,
  FaceBookIcon,
  GithubIcon,
  GridColumn,
  Icon,
  InstagramIcon,
  Layouts,
  LinkedInIcon,
  Router,
  Snackbar,
  TwitterIcon,
  WalletGrid,
  WalletShapes
} from '~/components'
import { getSnackbar } from '~/lib/snackbar.server'
import { getKycStatus, getLinkedIdentities } from '~/lib/wallet.server'
import { KycStatus } from '~/routes/_index/route'

export async function loader({ request }: LoaderArgs) {
  const identities = await getLinkedIdentities(request)

  const kycStatus = await getKycStatus(request)

  const snackbar = await getSnackbar(request)

  return json({
    snackbar,
    linkedIdentities: identities,
    kycStatus: kycStatus.kycStatus
  })
}

export const handle: ApplicationProps = {
  layout: Layouts.Wallet,
  scaffold: {
    header: {
      title: 'Identities',
      actions: <WalletShapes />
    },
    fab: Fab.Pay
  }
}

export const meta: MetaFunction = () => {
  return {
    title: 'Identities'
  }
}

export default function Page() {
  const { snackbar, linkedIdentities, kycStatus } =
    useLoaderData<typeof loader>()

  const [showSnackbar, setSnackbar] = useState<boolean>(snackbar.show ?? false)
  const location = useLocation()
  const pathSegments = location.pathname.split('/').filter(Boolean)
  return (
    <WalletGrid>
      <GridColumn
        hideOnMobile={pathSegments[pathSegments.length - 1] !== 'identities'}
        className='col-span-full lg:col-span-6'
      >
        {kycStatus == KycStatus.Unknown && (
          <Card>
            <CardHeader>
              <CardTitle>Wallet</CardTitle>
              <Chip color={ChipColor.orange}>Reserved</Chip>
            </CardHeader>
            <CardContent>
              <div className='flex items-start space-x-4'>
                <CardIcon>
                  <Icon>account_balance_wallet</Icon>
                </CardIcon>
                <div className='flex flex-col space-y-4'>
                  <p className='text-sm text-medium'>
                    Your wallet is reserved, we just need a few more details to
                    activate it.
                  </p>
                  <Router
                    prefetch='render'
                    className='text-sm font-medium text-primary'
                    to={route('/personal-details')}
                  >
                    Activate wallet
                  </Router>
                </div>
              </div>
            </CardContent>
          </Card>
        )}
        {linkedIdentities.twitter && (
          <Card>
            <CardHeader>
              <CardTitle>Twitter</CardTitle>
            </CardHeader>
            {linkedIdentities.twitter.map((identity) => (
              <CardLink
                key={identity.id}
                className='mt-2 flex items-center justify-between first-of-type:mt-4'
                to={route('/identities/:identityId', {
                  identityId: identity.id
                })}
              >
                <div className='flex space-x-3'>
                  <TwitterIcon className='text-medium' />
                  <span>@{identity.identifier}</span>
                </div>
                <div className='flex items-center space-x-3'>
                  {identity.state == 'verified' && (
                    <Chip color={ChipColor.green}>Verified</Chip>
                  )}
                  {identity.state == 'unverified' && (
                    <Chip color={ChipColor.yellow}>Unverified</Chip>
                  )}
                  {identity.state == 'failed' && (
                    <Chip color={ChipColor.red}>Failed</Chip>
                  )}
                  {identity.state == 'pending' && (
                    <Chip color={ChipColor.orange}>Pending</Chip>
                  )}
                  <Icon>navigate_next</Icon>
                </div>
              </CardLink>
            ))}
            <CardContent>
              <Router
                className='rounded text-sm font-medium text-primary'
                to={route('/connect/twitter')}
              >
                Connect another twitter identity
              </Router>
            </CardContent>
          </Card>
        )}
        {!linkedIdentities.twitter && kycStatus == KycStatus.Approved && (
          <Card>
            <div className='flex items-center space-x-4'>
              <CardIcon>
                <TwitterIcon />
              </CardIcon>
              <div className='flex flex-col space-y-1'>
                <h3 className='font-medium text-medium'>Twitter</h3>
                <Router
                  className='text-sm font-medium text-primary'
                  to={route('/connect/twitter')}
                >
                  Connect a Twitter identity
                </Router>
              </div>
            </div>
          </Card>
        )}
        {!linkedIdentities.github && kycStatus == KycStatus.Approved && (
          <Card>
            <div className='flex items-center space-x-4'>
              <CardIcon>
                <GithubIcon />
              </CardIcon>
              <div className='flex flex-col space-y-1'>
                <h3 className='font-medium text-medium'>Github</h3>
                <p className='text-sm text-disabled'>Coming soon</p>
              </div>
            </div>
          </Card>
        )}
        {!linkedIdentities.linkedIn && kycStatus == KycStatus.Approved && (
          <Card>
            <div className='flex items-center space-x-4'>
              <CardIcon>
                <LinkedInIcon />
              </CardIcon>
              <div className='flex flex-col space-y-1'>
                <h3 className='font-medium text-medium'>LinkedIn</h3>
                <p className='text-sm text-disabled'>Coming soon</p>
              </div>
            </div>
          </Card>
        )}
        {!linkedIdentities.facebook && kycStatus == KycStatus.Approved && (
          <Card>
            <div className='flex items-center space-x-4'>
              <CardIcon>
                <FaceBookIcon />
              </CardIcon>
              <div className='flex flex-col space-y-1'>
                <h3 className='font-medium text-medium'>Facebook</h3>
                <p className='text-sm text-disabled'>Coming soon</p>
              </div>
            </div>
          </Card>
        )}
        {!linkedIdentities.instagram && kycStatus == KycStatus.Approved && (
          <Card>
            <div className='flex items-center space-x-4'>
              <CardIcon>
                <InstagramIcon />
              </CardIcon>
              <div className='flex flex-col space-y-1'>
                <h3 className='font-medium text-medium'>Instagram</h3>
                <p className='text-sm text-disabled'>Coming soon</p>
              </div>
            </div>
          </Card>
        )}
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
        onClose={() => setSnackbar(false)}
      />
    </WalletGrid>
  )
}
