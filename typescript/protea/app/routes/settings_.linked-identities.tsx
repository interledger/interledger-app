import type { LoaderArgs, MetaFunction } from '@remix-run/node'
import { json } from '@remix-run/node'
import { useLoaderData } from '@remix-run/react'
import { route } from 'routes-gen'
import {
  Card,
  Chip,
  ChipColor,
  FaceBookIcon,
  GithubIcon,
  Icon,
  InstagramIcon,
  Layouts,
  LinkedInIcon,
  Router,
  Snackbar,
  TwitterIcon
} from '~/components'
import { getKycStatus, getLinkedIdentities } from '~/lib/wallet.server'
import { useState } from 'react'
import { getSnackbar } from '~/lib/snackbar.server'

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

export const handle = {
  title: 'Linked identities',
  layout: Layouts.FocusLayout
}

export const meta: MetaFunction = () => {
  return {
    title: 'Settings | Linked identities'
  }
}

export default function Page() {
  const { snackbar, linkedIdentities } = useLoaderData<typeof loader>()

  const [showSnackbar, setSnackbar] = useState<boolean>(snackbar.show ?? false)

  return (
    <>
      {linkedIdentities.twitter && (
        <Card>
          <h1 className='font-display text-lg font-medium'>Twitter</h1>
          {linkedIdentities.twitter.map((identity) => (
            <Router
              key={identity.id}
              className='mt-2 first-of-type:mt-6 flex items-center justify-between rounded-xl bg-nav p-3 text-medium hover:bg-nav-hover'
              to={route('/settings/linked-identities/:identityId', {
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
            className='mt-4 text-sm font-medium text-primary rounded focus-visible:outline focus-visible:outline-2 focus-visible:outline-focus'
            to={route('/connect/twitter')}
          >
            Connect another twitter identity
          </Router>
        </Card>
      )}
      {!linkedIdentities.twitter && (
        <Card className='space-y-4'>
          <div className='flex items-center space-x-4'>
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
        </Card>
      )}
      {!linkedIdentities.github && (
        <Card className='space-y-4'>
          <div className='flex items-center space-x-4'>
            <div className='flex items-center justify-between rounded-full bg-nav p-5 text-medium'>
              <GithubIcon />
            </div>
            <div className='flex flex-col space-y-1'>
              <h1 className='font-medium text-medium'>Github</h1>
              <p className='font-medium text-sm text-disabled'>Coming soon</p>
            </div>
          </div>
        </Card>
      )}
      {!linkedIdentities.linkedIn && (
        <Card className='space-y-4'>
          <div className='flex items-center space-x-4'>
            <div className='flex items-center justify-between rounded-full bg-nav p-5 text-medium'>
              <LinkedInIcon />
            </div>
            <div className='flex flex-col space-y-1'>
              <h1 className='font-medium text-medium'>LinkedIn</h1>
              <p className='font-medium text-sm text-disabled'>Coming soon</p>
            </div>
          </div>
        </Card>
      )}
      {!linkedIdentities.facebook && (
        <Card className='space-y-4'>
          <div className='flex items-center space-x-4'>
            <div className='flex items-center justify-between rounded-full bg-nav p-5 text-medium'>
              <FaceBookIcon />
            </div>
            <div className='flex flex-col space-y-1'>
              <h1 className='font-medium text-medium'>Facebook</h1>
              <p className='font-medium text-sm text-disabled'>Coming soon</p>
            </div>
          </div>
        </Card>
      )}
      {!linkedIdentities.instagram && (
        <Card className='space-y-4'>
          <div className='flex items-center space-x-4'>
            <div className='flex items-center justify-between rounded-full bg-nav p-5 text-medium'>
              <InstagramIcon />
            </div>
            <div className='flex flex-col space-y-1'>
              <h1 className='font-medium text-medium'>Instagram</h1>
              <p className='font-medium text-sm text-disabled'>Coming soon</p>
            </div>
          </div>
        </Card>
      )}
      <Snackbar
        message={snackbar.message}
        action={snackbar.action}
        icon={snackbar.icon}
        show={showSnackbar}
        id='cookie-snackbar'
        dismissAfter={3000}
        onClose={() => setSnackbar(false)}
      />
    </>
  )
}
