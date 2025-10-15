import {
  redirect,
  type ActionFunctionArgs,
  type LoaderFunctionArgs,
  type MetaFunction
} from '@remix-run/node'
import { useLoaderData } from '@remix-run/react'
import type { ApplicationProps } from '~/components'
import { Layouts } from '~/components'
import { isConnectError } from '~/lib/error.server'
import { grpc } from '~/lib/grpc.server'
import { mergeMeta } from '~/lib/meta'
import styles from '~/styles/flags.css'
import {
  ChimoneyDepositPage,
  chimoneyAmountAction,
  chimoneyDepositLoader,
  chimoneySuccessfullDepositAction
} from './chimoney'
import {
  FynbosDepositPage,
  fynbosDepositAction,
  fynbosDepositLoader,
  xagoTestAccountDepositAction
} from './fynbos'
import { GatehubDepositPage, gatehubDepositLoader } from './gatehub'
import { KRATOS_URL } from '~/lib/kratos.server'
import { PTIDepositAction, ptiDepositLoader, PTIDepositPage } from './pti'

export async function loader(args: LoaderFunctionArgs) {
  const session = await fetch(`${KRATOS_URL}/sessions/whoami`, {
      headers: args.request.headers
    })
    if (session.status === 401) {
      return redirect('/login')
    }

  const providerResponse = await grpc.getOnOffRampProvider(args.request, {})
  if (isConnectError(providerResponse)) throw providerResponse.error
  if (providerResponse.provider == 'gatehub') {
    return gatehubDepositLoader(args)
  } else if (providerResponse.provider == 'chimoney') {
    return chimoneyDepositLoader(args)
  } else if (providerResponse.provider == 'pti') {
    return ptiDepositLoader(args)
  } else return fynbosDepositLoader(args)
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
  } else if (provider == 'chimoney') {
    return <ChimoneyDepositPage />
  } else if  (provider == 'pti') {
    return <PTIDepositPage />
  } else return <FynbosDepositPage />
}

export async function action(args: ActionFunctionArgs) {
  const formName = (await args.request.clone().formData()).get(
    'formName'
  ) as string

  if (formName === 'chimoney-amount') {
    return chimoneyAmountAction(args)
  } else if (formName === 'chimoney-successfull-deposit') {
    return chimoneySuccessfullDepositAction(args)
  } else if (formName === 'xago-test-account-deposit') {
    return xagoTestAccountDepositAction(args)
  } else if (formName === 'pti-deposit') {
    return PTIDepositAction(args)
  }

  return fynbosDepositAction(args)
}
