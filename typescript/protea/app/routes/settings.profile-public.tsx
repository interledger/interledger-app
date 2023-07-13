import type { LoaderArgs, MetaFunction } from '@remix-run/node'
import { json } from '@remix-run/node'
import { useLoaderData } from '@remix-run/react'
import { useState } from 'react'
import { route } from 'routes-gen'
import type { ApplicationProps } from '~/components'
import {
  AnchorRouter,
  Card,
  CardContent,
  CardLink,
  Icon,
  Layouts,
  Snackbar
} from '~/components'
import { Label } from '~/components/Label'
import { getSnackbar } from '~/lib/snackbar.server'
import { getPublicWalletDetails, getWalletInfo } from '~/lib/wallet.server'

export async function loader({ request }: LoaderArgs) {
  const walletInfo = await getWalletInfo(request)
  const wallet = await getPublicWalletDetails(request, walletInfo.walletID)

  const snackbar = await getSnackbar(request)

  return json({
    name: wallet.publicName,
    walletInfo,
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
  const { name, walletInfo, snackbar } = useLoaderData<typeof loader>()
  const [showSnackbar, setSnackbar] = useState<boolean>(snackbar.show ?? false)

  return (
    <>
      <Card>
        <CardContent>
          <p>
            The following information will appear on your public{' '}
            <AnchorRouter className='text-primary' to={walletInfo.url}>
              Fynbos.me
            </AnchorRouter>{' '}
            page.
          </p>
        </CardContent>
      </Card>
      <Card>
        <Label>Public name</Label>
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
