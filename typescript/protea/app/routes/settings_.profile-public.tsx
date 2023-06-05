import type { LoaderArgs } from '@remix-run/node'
import { json } from '@remix-run/node'
import { useLoaderData } from '@remix-run/react'
import { useState } from 'react'
import { route } from 'routes-gen'
import type { ApplicationProps } from '~/components'
import {
  AnchorRouter,
  Card,
  Icon,
  Layouts,
  Router,
  Snackbar
} from '~/components'
import { getSnackbar } from '~/lib/snackbar.server'
import {
  getPublicWalletDetails,
  getWalletPaymentPointer
} from '~/lib/wallet.server'

export async function loader({ request }: LoaderArgs) {
  const paymentPointer = await getWalletPaymentPointer(request)
  const wallet = await getPublicWalletDetails(request, paymentPointer.walletID)

  const snackbar = await getSnackbar(request)

  return json({
    name: wallet.publicName,
    paymentPointer,
    snackbar
  })
}

export const handle: ApplicationProps = {
  layout: Layouts.Focus,
  scaffold: {
    header: {
      back: route('/settings'),
      title: 'Public information'
    }
  }
}

export default function Page() {
  const { name, paymentPointer, snackbar } = useLoaderData<typeof loader>()
  const [showSnackbar, setSnackbar] = useState<boolean>(snackbar.show ?? false)

  return (
    <>
      <Card>
        <p>
          The information below will appear on your{' '}
          <AnchorRouter className='text-primary' to={paymentPointer.url}>
            public payment pointer
          </AnchorRouter>{' '}
          page.
        </p>
        <h2 className='mt-6 text-sm font-medium'>Public name</h2>
        <Router
          to={route('/settings/profile-public/name')}
          className='mt-2 flex items-center justify-between rounded-xl bg-nav p-3 text-medium hover:bg-nav-hover'
        >
          <div className='flex space-x-3'>
            <Icon>flag</Icon>
            <span>{name}</span>
          </div>
          <Icon>navigate_next</Icon>
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
