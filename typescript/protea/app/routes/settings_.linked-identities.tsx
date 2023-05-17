import type { LoaderArgs, MetaFunction } from '@remix-run/node'
import { json } from '@remix-run/node'
import { useLoaderData } from '@remix-run/react'
import { route } from 'routes-gen'
import { Card, Icon, Layouts, Router, Snackbar } from '~/components'
import {
  getKycStatus,
  getLinkedAccounts,
  getLinkedIdentities
} from '~/lib/wallet.server'
import { KycStatus } from '~/routes/_index/route'
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
      {linkedIdentities && linkedIdentities.length > 0 && (
        <Card>
          <Router
            className='mt-6 text-sm font-medium text-primary'
            to={route('/link-account')}
          >
            Link another account
          </Router>
        </Card>
      )}
      <Card>
        <div>
          <p>Twitter</p>
        </div>
        {linkedIdentities.map((identity) => (
          <Router
            key={identity.id}
            className='mt-2 flex items-center justify-between rounded-xl bg-nav p-3 text-medium hover:bg-nav-hover'
            to={route('/')}
          >
            {identity.identifier} - ({identity.state})
          </Router>
        ))}
        <Router
          className='text-sm font-medium text-primary'
          to={route('/twitter')}
        >
          Link twitter identity
        </Router>
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
