import type { LoaderArgs, MetaFunction } from '@remix-run/node'
import { json } from '@remix-run/node'
import { useLoaderData } from '@remix-run/react'
import { useState } from 'react'
import { route } from 'routes-gen'
import type { ApplicationProps } from '~/components'
import {
  AnchorRouter,
  Card,
  CardColumn,
  CardLink,
  Icon,
  Layouts,
  Snackbar
} from '~/components'
import { Label } from '~/components/Label'
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
  layout: Layouts.Wallet,
  scaffold: {
    header: {
      back: route('/settings'),
      title: 'Public information'
    },
    isNested: true
  }
}

export const meta: MetaFunction = () => {
  return {
    title: 'Public information'
  }
}

export default function Page() {
  const { name, paymentPointer, snackbar } = useLoaderData<typeof loader>()
  const [showSnackbar, setSnackbar] = useState<boolean>(snackbar.show ?? false)

  return (
    <>
      <Card>
        <p>
          The following information will appear on your public{' '}
          <AnchorRouter className='text-primary' to={paymentPointer.url}>
            Fynbos.me
          </AnchorRouter>{' '}
          page.
        </p>
      </Card>
      <Card>
        <CardColumn>
          <Label className='-mb-5'>Public name</Label>
          <CardLink
            prefetch='intent'
            className='items-center justify-between'
            to={route('/settings/profile-public/name')}
          >
            <div className='flex space-x-3'>
              <Icon>flag</Icon>
              <span>{name}</span>
            </div>
            <Icon>navigate_next</Icon>
          </CardLink>
        </CardColumn>
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
