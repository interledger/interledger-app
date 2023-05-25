import type { LoaderArgs, MetaFunction } from '@remix-run/node'
import { json } from '@remix-run/node'
import { useLoaderData } from '@remix-run/react'
import { route } from 'routes-gen'
import {
  Card,
  Chip,
  ChipColor,
  Icon,
  Layouts,
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
      {/*TODO Handle other identity types when we get there*/}
      {linkedIdentities.length > 0 && (
        <Card>
          {linkedIdentities.map((identity) => (
            <Router
              key={identity.id}
              className='mt-2 flex items-center justify-between rounded-xl bg-nav p-3 text-medium hover:bg-nav-hover'
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
            Link another identity
          </Router>
        </Card>
      )}
      {linkedIdentities.length == 0 && (
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
                Link a Twitter identity
              </Router>
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
