import {
  data,
  redirect,
  type ActionFunctionArgs,
  type LoaderFunctionArgs,
  type MetaFunction,
} from 'react-router';
import { Code } from '@bufbuild/connect'
import { useLoaderData } from 'react-router';
import type { ApplicationProps } from '~/components'
import { Layouts } from '~/components'
import { isConnectError } from '~/lib/error.server'
import { grpc } from '~/lib/grpc.server'
import { mergeMeta } from '~/lib/meta'
import { href } from 'react-router'
import styles from '~/styles/flags.css?url'
import { ChimoneyDepositPage } from './chimoney'
import { FynbosDepositPage } from './fynbos'
import { GatehubDepositPage } from './gatehub'
import { jsonWithCSRF, validateCSRFToken } from '~/lib/csrf.server'
import { redirectWithSnackbar } from '~/lib/snackbar.server'
import {
  getBalancesForTransfer,
  getLinkedAccountsForDeposit
} from '~/data/accounts.server'
import type {
  Balance,
  XagoDepositDetails
} from '~/generated/connect/backend/v1/backend_pb'
import type { FormattedLinkedAccount } from '~/data/accounts.server'
import { KRATOS_URL } from '~/lib/kratos.server'

import { getKycStatus } from '~/data/wallet.server'
import { KycStatus } from '~/lib/types'
import { ActionData, LoaderData } from '~/lib/types.server';

function stringToBigInt(amount: string) {
  if (amount == '') return BigInt(0)
  const dotIndex = amount.lastIndexOf('.')
  if (dotIndex > -1) {
    const amounts = amount.split('.')
    return BigInt(amounts[0] + amounts[1].slice(0, 2).padEnd(2, '0'))
  }
  return BigInt(parseFloat(amount) * 100)
}

/**
 * Loaders:
 */
async function gatehubDepositLoader({ request }: LoaderFunctionArgs) {
  const widgetResponse = await grpc.getGatehubDepositWidget(request, {})
  if (isConnectError(widgetResponse)) throw widgetResponse.error

  return jsonWithCSRF(request, {
    provider: 'gatehub',
    gatehubWidgetUrl: widgetResponse.widgetUrl
  })
}

async function chimoneyDepositLoader({ request }: LoaderFunctionArgs) {
  return jsonWithCSRF(request, { provider: 'chimoney' })
}

async function fynbosDepositLoader({ request }: LoaderFunctionArgs) {
  const providerResponse = await grpc.getOnOffRampProvider(request, {})
  if (isConnectError(providerResponse)) throw providerResponse.error
  const url = new URL(request.url)
  let balanceAccount: Balance | undefined
  let balance: FormattedLinkedAccount | undefined
  let depositDetails: XagoDepositDetails | undefined

  const balanceResponse = await grpc.getBalances(request, {})
  if (isConnectError(balanceResponse)) throw balanceResponse.error

  const balances = await getBalancesForTransfer(request)

  const linkedAccount = url.searchParams.get('linkedAccount')

  if (linkedAccount) {
    balanceAccount = balanceResponse.balances.find(
      (acc) => acc.linkedAccount == linkedAccount
    )
  } else {
    balanceAccount = balanceResponse.balances[0]
  }
  balance = balances.find((acc) => acc.id == balanceAccount?.linkedAccount)
  if (
    !balanceAccount ||
    typeof balanceAccount == 'undefined' ||
    !balance ||
    typeof balance == 'undefined'
  )
    throw data({}, { status: 404 })

  const linkedAccounts = await getLinkedAccountsForDeposit(
    request,
    balanceAccount.linkedAccount
  )

  if (linkedAccounts.length == 0 && balanceAccount.countryCode == 'ZA') {
    let details = await grpc.getXagoDepositDetails(request, {
      linkedAccount: balanceAccount.linkedAccount
    })
    if (isConnectError(details)) throw details.errorResponse

    let ret = details.details.filter(
      (d) => d.currency == balanceAccount?.currency
    )
    depositDetails = ret[0]
  }

  return jsonWithCSRF(request, {
    provider: providerResponse.provider,
    balanceAccount,
    balance,
    balances,
    linkedAccounts,
    depositDetails
  })
}

export type GatehubDepositLoaderData  = LoaderData<typeof gatehubDepositLoader>
export type ChimoneyDepositLoaderData = LoaderData<typeof chimoneyDepositLoader>
export type FynbosDepositLoaderData   = LoaderData<typeof fynbosDepositLoader>

export async function loader(args: LoaderFunctionArgs) {
  const session = await fetch(`${KRATOS_URL}/sessions/whoami`, {
    headers: args.request.headers
  })
  if (session.status === 401) {
    return redirect('/login')
  }
  const { kycStatus } = await getKycStatus(args.request)
  if (kycStatus != KycStatus.Approved)
    return redirect(href('/personal-details'))

  const providerResponse = await grpc.getOnOffRampProvider(args.request, {})
  if (isConnectError(providerResponse)) throw providerResponse.error
  if (providerResponse.provider == 'gatehub') {
    return gatehubDepositLoader(args)
  } else if (providerResponse.provider == 'chimoney') {
    return chimoneyDepositLoader(args)
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
  } else return <FynbosDepositPage />
}

/**
 * Actions:
 */
async function chimoneyAmountAction({ request }: ActionFunctionArgs) {
  const form = await request.formData()
  const depositAmount = String(form.get('depositAmount') || '')

  const response = await grpc.getChimoneyDepositLink(request, {
    amount: stringToBigInt(depositAmount),
    asset: 'CAD',
    assetScale: 2
  })
  if (isConnectError(response)) throw response.error

  return data({
    chimoneyWidget: response.link
  })
}

async function chimoneySuccessfullDepositAction({
  request
}: ActionFunctionArgs) {
  const form = await request.formData()
  const issueId = String(form.get('issueId') || '')

  const response = await grpc.createChimoneyDeposit(request, { issueId })
  if (isConnectError(response)) throw response.error

  return redirect(href('/'))
}

async function fynbosDepositAction({ request }: ActionFunctionArgs) {
  const form = await request.formData()
  const depositAmount = String(form.get('depositAmount') || '')
  const toLinkedAccount = form.get('toLinkedAccount') as string
  const fromLinkedAccount = form.get('fromLinkedAccount') as string
  const provider = form.get('provider') as string
  const note = form.get('note') as string

  await validateCSRFToken(request, form)

  const errors = {
    form: '',
    depositAmount: '',
    toLinkedAccount: '',
    note: ''
  }

  const cc = provider === 'pit' ? 'USD' : 'ZAR'
  const depositResponse = await grpc.depositBalance(request, {
    fromLinkedAccount: fromLinkedAccount,
    toLinkedAccount: toLinkedAccount,
    amount: {
      amount: stringToBigInt(depositAmount),
      asset: cc,
      assetScale: 2
    },
    note
  })
  if (isConnectError(depositResponse)) {
    if (depositResponse.code == Code.InvalidArgument) {
      return depositResponse.error({
        errors: {
          ...errors,
          depositAmount: depositResponse?.fieldViolations?.find((v: { field: string }) => v.field === 'amount')?.description ?? ''
        }
      })
    }
    if (
      depositResponse.code == Code.FailedPrecondition &&
      depositResponse.violations.findIndex(
        (violation) =>
          violation.type === 'Payment' &&
          violation.subject === 'insufficientFunds'
      ) > -1
    ) {
      return depositResponse.error({
        errors: {
          ...errors,
          depositAmount: 'You have insufficient funds available.'
        }
      })
    }
    errors.form = 'Failed to create deposital.'
    return depositResponse.error(
      { errors },
      {},
      {
        action: 'Contact support'
      }
    )
  }

  return redirect(
    href('/deposit/:paymentId', {
      paymentId: depositResponse.id
    })
  )
}

async function xagoTestAccountDepositAction({
  request
}: ActionFunctionArgs) {
  const form = await request.formData()

  await validateCSRFToken(request, form)

  const response = await grpc.depositTestXago(request, {})

  if (isConnectError(response)) {
    return response.errorResponse
  }

  return redirectWithSnackbar(request, href('/'), {
    message: 'Test deposit added successfully.',
    icon: 'close'
  })
}

export type ChimoneyAmountActionData             = ActionData<typeof chimoneyAmountAction>
export type ChimoneySuccessfullDepositActionData = ActionData<typeof chimoneySuccessfullDepositAction>
export type FynbosDepositActionData              = ActionData<typeof fynbosDepositAction>
export type XagoTestAccountDepositActionData     = ActionData<typeof xagoTestAccountDepositAction>

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
  }
  return fynbosDepositAction(args)
}
