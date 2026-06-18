import {
  redirect,
  useLoaderData,
  type ActionFunctionArgs,
  type LoaderFunctionArgs,
  type MetaFunction
} from 'react-router'
import type { ApplicationProps } from '~/components'
import { Layouts } from '~/components'
import { isConnectError } from '~/lib/error.server'
import { grpc } from '~/lib/grpc.server'
import { KRATOS_URL } from '~/lib/kratos/kratos-client.server'
import { mergeMeta } from '~/lib/meta'
import styles from '~/styles/flags.css?url'
import { ilpDepositAction, xagoTestAccountDepositAction } from './action.server'
import { GatehubDepositPage } from './gatehub'
import { IlpDepositPage } from './ilp'
import { gatehubDepositLoader, ilpDepositLoader } from './loader.server'

export async function loader(args: LoaderFunctionArgs) {
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
  } else return ilpDepositLoader(args)
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
  const { provider } = useLoaderData<typeof loader>()

  if (provider == 'gatehub') {
    return <GatehubDepositPage />
  }

  return <IlpDepositPage />
}

export async function action(args: ActionFunctionArgs) {
  const formName = (await args.request.clone().formData()).get(
    'formName'
  ) as string

  if (formName === 'xago-test-account-deposit') {
    return xagoTestAccountDepositAction(args)
  }
  return ilpDepositAction(args)
}
