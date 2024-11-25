import type {
  ActionFunctionArgs,
  LoaderFunctionArgs,
  MetaFunction
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
  fynbosDepositLoader
} from './fynbos'
import { GatehubDepositPage, gatehubDepositLoader } from './gatehub'
import { jsonWithCSRF } from '~/lib/csrf.server'
import PtiDepositPage from './pti'

export async function loader(args: LoaderFunctionArgs) {
  const providerResponse = await grpc.getOnOffRampProvider(args.request, {})
  if (isConnectError(providerResponse)) throw providerResponse.error

  if (providerResponse.provider == 'gatehub') {
    return gatehubDepositLoader(args)
  } else if (providerResponse.provider == 'chimoney') {
    return chimoneyDepositLoader(args)
  } else if (providerResponse.provider == 'pti') {
    return jsonWithCSRF(args.request, {
      provider: 'pti',
      ptiWidget: providerResponse.ptiWidget
    })
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
  } else if (provider == 'pti') {
    return <PtiDepositPage />
  } else return <FynbosDepositPage />
}

export async function action(args: ActionFunctionArgs) {
  const formName = (await args.request.clone().formData()).get(
    'formName'
  ) as string

  if (formName == 'chimoney-amount') {
    return chimoneyAmountAction(args)
  } else if (formName == 'chimoney-successfull-deposit') {
    return chimoneySuccessfullDepositAction(args)
  }

  return fynbosDepositAction(args)
}
