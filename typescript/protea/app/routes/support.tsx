import type { Route } from './+types/support'
import type { ApplicationProps } from '~/components'
import { RootLoaderData } from '~/root'
import { useRouteLoaderData } from 'react-router'
import {
  AnchorRouter,
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  GridColumn,
  Icon,
  Layouts,
  WalletGrid
} from '~/components'
import { jsonWithCSRF } from '~/lib/csrf.server'
import { getUserSession } from '~/lib/kratos/session.server'
import { mergeMeta } from '~/lib/meta'

export async function loader({ request }: Route.LoaderArgs) {
  await getUserSession(request)
  return jsonWithCSRF(request, {})
}

export const handle: ApplicationProps = {
  layout: Layouts.Wallet,
  scaffold: {
    header: {
      title: 'Support'
    }
  }
}

export const meta = mergeMeta(() => [
  {
    title: 'Support'
  }
])

export default function Page() {
  const { env } = useRouteLoaderData('root') as RootLoaderData
  const supportEmail = env.supportEmail

  return (
    <WalletGrid>
      <GridColumn className='col-span-full lg:col-span-6'>
        <Card>
          <CardHeader>
            <CardTitle>Support details</CardTitle>
          </CardHeader>
          <CardContent>
            <p>
              Please share all relevant details so we can assist you effectively
              and efficiently.
            </p>
          </CardContent>
          <CardContent>
            <div className='mt-4 flex items-center space-x-2 text-medium'>
              <Icon>mail</Icon>
              <AnchorRouter
                to={`mailto:${supportEmail}`}
                className='text-sm text-primary'
              >
                {supportEmail}
              </AnchorRouter>
            </div>
          </CardContent>
        </Card>
      </GridColumn>
    </WalletGrid>
  )
}
