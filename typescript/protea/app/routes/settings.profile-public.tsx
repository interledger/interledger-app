import type { LoaderFunctionArgs, MetaFunction } from '@remix-run/node'
import { json } from '@remix-run/node'
import { useLoaderData } from '@remix-run/react'
import { route } from 'routes-gen'
import type { ApplicationProps } from '~/components'
import {
  AnchorRouter,
  Card,
  CardContent,
  CardLink,
  Icon,
  Layouts
} from '~/components'
import { Label } from '~/components/Label'
import { getPublicWalletDetails, getWalletInfo } from '~/data/wallet.server'
import { mergeMeta } from '~/lib/meta'

export async function loader({ request }: LoaderFunctionArgs) {
  const walletInfo = await getWalletInfo(request)
  const wallet = await getPublicWalletDetails(request, walletInfo.walletID)

  return json({
    name: wallet.publicName,
    walletInfo
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

export const meta: MetaFunction = mergeMeta(() => [
  {
    title: 'Public information'
  }
])

export default function Page() {
  const { name, walletInfo } = useLoaderData<typeof loader>()

  return (
    <>
      <Card>
        <CardContent>
          <p>
            The following information will appear on your public{' '}
            <AnchorRouter className='text-primary' to={walletInfo.url}>
              ilp.link
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
            <Icon>account_circle</Icon>
            <span>{name}</span>
          </div>
          <Icon>navigate_next</Icon>
        </CardLink>
      </Card>
    </>
  )
}
