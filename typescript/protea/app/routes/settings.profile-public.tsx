import type { LoaderArgs } from '@remix-run/node'
import { json } from '@remix-run/node'
import { useLoaderData } from '@remix-run/react'
import { route } from 'routes-gen'
import {
  AnchorRouter,
  Icon,
  Layouts,
  Router,
  Snackbar,
  WalletGrid
} from '~/components'
import {
  getPublicWalletDetails,
  getWalletPaymentPointer
} from '~/lib/wallet.server'
import { getSnackbar } from '~/lib/snackbar.server'
import { useState } from 'react'

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

export const handle = {
  layout: Layouts.WalletLayout
}

export default function Page() {
  const { name, paymentPointer, snackbar } = useLoaderData<typeof loader>()
  const [showSnackbar, setSnackbar] = useState<boolean>(snackbar.show ?? false)

  return (
    <WalletGrid>
      <div className='col-span-full flex flex-col rounded-2xl bg-page p-4 pb-8 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <h1 className='font-display text-2xl font-medium'>
          Public information
        </h1>
        <p className='mt-4'>
          The information below will appear on your{' '}
          <AnchorRouter className='text-primary' to={paymentPointer.url}>
            public payment pointer
          </AnchorRouter>{' '}
          page.
        </p>
        <h2 className='mt-6 text-sm font-medium'>Public name</h2>
        <Router
          to={route('/settings/profile-public/name')}
          className='mt-2 flex items-center justify-between rounded-xl bg-container p-3 text-medium hover:bg-container-hover'
        >
          <div className='flex space-x-3'>
            <Icon>flag</Icon>
            <span>{name}</span>
          </div>
          <Icon>navigate_next</Icon>
        </Router>
      </div>
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
