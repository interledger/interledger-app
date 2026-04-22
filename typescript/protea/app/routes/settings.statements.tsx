import type { Route } from './+types/settings.statements'
import { redirect } from 'react-router'
import { href } from 'react-router'
import type { ApplicationProps } from '~/components'
import {
  ButtonRouter,
  Card,
  CardContent,
  Layouts
} from '~/components'
import { isConnectError } from '~/lib/error.server'
import { grpc } from '~/lib/grpc.server'
import { mergeMeta } from '~/lib/meta'

export async function loader({ request }: Route.LoaderArgs) {
  const providerResponse = await grpc.getOnOffRampProvider(request, {})
  if (isConnectError(providerResponse)) throw providerResponse.errorResponse

  if (providerResponse.provider !== 'gatehub') {
    return redirect(href('/settings'))
  }

  return {}
}

export const handle: ApplicationProps = {
  layout: Layouts.Wallet,
  scaffold: {
    header: {
      back: href('/settings'),
      title: 'Statements'
    },
    isNested: true
  }
}

export const meta = mergeMeta(() => [{ title: 'Statements' }])

export default function Page() {
  return (
    <>
      <Card>
        <CardContent>
          <p>Download your account statement as a PDF.</p>
        </CardContent>
      </Card>

      <ButtonRouter to={href('/api/statements/accountStatement')} reloadDocument>
        Download Statement
      </ButtonRouter>
    </>
  )
}
