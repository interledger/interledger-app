import type { LoaderArgs, MetaFunction } from '@remix-run/node'
import { json } from '@remix-run/node'
import { useLoaderData } from '@remix-run/react'
import { route } from 'routes-gen'
import { Card, Layouts, Router, Snackbar } from '~/components'
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

  // TODO we need to add a fallback if we can't open a popup window
  const openPopup = () => {
    window.open('/connect/twitter', 'twitter_connect', 'width=700,height=560')
  }

  return (
    <>
      <Card>
        <div>
          <p>Twitter</p>
        </div>
        {linkedIdentities.map((identity) => (
          <Router
            key={identity.id}
            className='mt-2 flex items-center justify-between rounded-xl bg-nav p-3 text-medium hover:bg-nav-hover'
            to={route('/settings/linked-identities/:identityId', {
              identityId: identity.id
            })}
          >
            {identity.identifier} - ({identity.state})
          </Router>
        ))}
        <a
          className='text-sm font-medium text-primary rounded focus-visible:outline focus-visible:outline-2 focus-visible:outline-focus'
          onClick={openPopup}
        >
          Link twitter identity
        </a>
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
    </>
  )
}
