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
  WalletGrid
} from '~/components'
import { getSnackbar } from '~/lib/snackbar.server'
import { getKycStatus, getLinkedIdentities } from '~/lib/wallet.server'

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
      actions: [{ type: 'shapes' }]
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
  const { snackbar, linkedIdentities } = useLoaderData<typeof loader>()

  const [showSnackbar, setSnackbar] = useState<boolean>(snackbar.show ?? false)
  const location = useLocation()
  const pathSegments = location.pathname.split('/').filter(Boolean)
  return (
    <WalletGrid>
      <GridColumn
        hideOnMobile={pathSegments[pathSegments.length - 1] !== 'identities'}
        className='col-span-full lg:col-span-6'
      >
        {linkedIdentities.twitter && (
          <Card>
            <CardHeader>
              <CardTitle>Twitter</CardTitle>
            </CardHeader>
            <CardContent>
              {linkedIdentities.twitter.map((identity) => (
                <Router
                  key={identity.id}
                  className='mt-2 flex items-center justify-between rounded-xl bg-nav p-3 text-medium first-of-type:mt-6 hover:bg-nav-hover'
                  to={route('/identities/:identityId', {
                    identityId: identity.id
                  })}
                >
                  <div className='flex space-x-3'>
                    <TwitterIcon className='text-medium' />
                    <span>@{identity.identifier}</span>
                  </div>
                  <div className='flex space-x-3'>
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
                </Router>
              ))}
              <Router
                className='mt-4 rounded text-sm font-medium text-primary focus-visible:outline focus-visible:outline-2 focus-visible:outline-focus'
                to={route('/connect/twitter')}
              >
                Connect another twitter identity
              </Router>
            </CardContent>
          </Card>
        )}
        {!linkedIdentities.twitter && (
          <Card className='space-y-4'>
            <CardContent>
              <div className='-m-2 flex items-center space-x-4'>
                <div className='flex items-center justify-between rounded-full bg-nav p-5 text-medium'>
                  <TwitterIcon />
                </div>
                <div className='flex flex-col space-y-1'>
                  <h1 className='font-medium text-medium'>Twitter</h1>
                  <Router
                    className='text-sm font-medium text-primary'
                    to={route('/connect/twitter')}
                  >
                    Connect a Twitter identity
                  </Router>
                </div>
              </div>
            </CardContent>
          </Card>
        )}
        {!linkedIdentities.github && (
          <Card className='space-y-4'>
            <CardContent>
              <div className='-m-2 flex items-center space-x-4'>
                <div className='flex items-center justify-between rounded-full bg-nav p-5 text-medium'>
                  <GithubIcon />
                </div>
                <div className='flex flex-col space-y-1'>
                  <h3 className='font-medium text-medium'>Github</h3>
                  <p className='text-sm font-medium text-disabled'>
                    Coming soon
                  </p>
                </div>
              </div>
            </CardContent>
          </Card>
        )}
        {!linkedIdentities.linkedIn && (
          <Card className='space-y-4'>
            <CardContent>
              <div className='-m-2 flex items-center space-x-4'>
                <div className='flex items-center justify-between rounded-full bg-nav p-5 text-medium'>
                  <LinkedInIcon />
                </div>
                <div className='flex flex-col space-y-1'>
                  <h3 className='font-medium text-medium'>LinkedIn</h3>
                  <p className='text-sm font-medium text-disabled'>
                    Coming soon
                  </p>
                </div>
              </div>
            </CardContent>
          </Card>
        )}
        {!linkedIdentities.facebook && (
          <Card className='space-y-4'>
            <CardContent>
              <div className='-m-2 flex items-center space-x-4'>
                <div className='flex items-center justify-between rounded-full bg-nav p-5 text-medium'>
                  <FaceBookIcon />
                </div>
                <div className='flex flex-col space-y-1'>
                  <h3 className='font-medium text-medium'>Facebook</h3>
                  <p className='text-sm font-medium text-disabled'>
                    Coming soon
                  </p>
                </div>
              </div>
            </CardContent>
          </Card>
        )}
        {!linkedIdentities.instagram && (
          <Card className='space-y-4'>
            <CardContent>
              <div className='-m-2 flex items-center space-x-4'>
                <div className='flex items-center justify-between rounded-full bg-nav p-5 text-medium'>
                  <InstagramIcon />
                </div>
                <div className='flex flex-col space-y-1'>
                  <h3 className='font-medium text-medium'>Instagram</h3>
                  <p className='text-sm font-medium text-disabled'>
                    Coming soon
                  </p>
                </div>
              </div>
            </CardContent>
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
