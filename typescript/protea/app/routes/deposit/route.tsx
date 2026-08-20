import {
  href,
  redirect,
  useLoaderData,
  type ActionFunctionArgs,
  type LoaderFunctionArgs,
  type MetaFunction
} from 'react-router'
import type { ApplicationProps } from '~/components'
import { Alert, AlertBody, Icon, Layouts } from '~/components'
import { envBool } from '~/env.server'
import { isConnectError } from '~/lib/error.server'
import { grpc } from '~/lib/grpc.server'
import { KRATOS_URL } from '~/lib/kratos/kratos-client.server'
import { mergeMeta } from '~/lib/meta'
import styles from '~/styles/flags.css?url'
import { ilwDepositAction, xagoTestAccountDepositAction } from './action.server'
import { GatehubDepositPage } from './gatehub'
import { IlwDepositPage } from './ilw'
import { gatehubDepositLoader, ilwDepositLoader } from './loader.server'

export async function loader(args: LoaderFunctionArgs) {
  if (!envBool('DEPOSIT_ENABLED', true)) return null

  const session = await fetch(`${KRATOS_URL}/sessions/whoami`, {
    headers: args.request.headers
  })
  if (session.status === 401) {
    return redirect('/login')
  }
  const providerResponse = await grpc.getOnOffRampProvider(args.request, {})
  if (isConnectError(providerResponse)) throw providerResponse.errorResponse
  if (providerResponse.provider == 'gatehub') {
    return gatehubDepositLoader(args)
  } else return ilwDepositLoader(args)
}

export const handle: ApplicationProps = {
  layout: Layouts.Focus,
  scaffold: {
    header: {
      title: 'Deposit',
      back: '/'
    }
  }
}

export const meta: MetaFunction = mergeMeta(() => [
  {
    title: 'Deposit'
  }
])

export function links() {
  return [{ rel: 'stylesheet', href: styles }]
}

export default function Page() {
  const loaderData = useLoaderData<typeof loader>()

  if (!loaderData) {
    return (
      <Alert role='status'>
        <Icon>error</Icon>
        <AlertBody>Deposits are temporarily unavailable.</AlertBody>
      </Alert>
    )
  }

  if (loaderData.provider == 'gatehub') {
    return <GatehubDepositPage />
  }

  return <IlwDepositPage />
}

export async function action(args: ActionFunctionArgs) {
  if (!envBool('DEPOSIT_ENABLED', true)) throw redirect(href('/deposit'))

  const formName = (await args.request.clone().formData()).get(
    'formName'
  ) as string

  if (formName === 'xago-test-account-deposit') {
    return xagoTestAccountDepositAction(args)
  }
  return ilwDepositAction(args)
}
