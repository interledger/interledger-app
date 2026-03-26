import type { Route } from './+types/settings.profile-public'
import { data } from 'react-router';
import { useLoaderData } from 'react-router';
import { href } from 'react-router'
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

export async function loader({ request }: Route.LoaderArgs) {
  const walletInfo = await getWalletInfo(request)
  const wallet = await getPublicWalletDetails(request, walletInfo.walletID)

  return data({
    name: wallet.publicName,
    walletInfo
  })
}

export const handle: ApplicationProps = {
  layout: Layouts.Wallet,
  scaffold: {
    header: {
      back: href('/settings'),
      title: 'Public information'
    },
    isNested: true
  }
}

export const meta = mergeMeta(() => [
  {
    title: 'Public information'
  }
])

export default function Page() {
  const { name, walletInfo } = useLoaderData()

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
          to={href('/settings/profile-public/name')}
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
