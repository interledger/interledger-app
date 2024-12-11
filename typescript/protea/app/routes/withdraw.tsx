import { Code } from '@bufbuild/connect'
import type { PlainMessage } from '@bufbuild/protobuf/dist/types/message'
import type {
  ActionFunctionArgs,
  LoaderFunctionArgs,
  MetaFunction
} from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import {
  Form,
  useActionData,
  useLoaderData,
  useSearchParams,
  useSubmit
} from '@remix-run/react'
import {
  useCallback,
  useEffect,
  useState,
  type ChangeEventHandler
} from 'react'
import { route } from 'routes-gen'
import type { ApplicationProps, SelectOptions } from '~/components'
import {
  Button,
  Card,
  CardContent,
  CardIcon,
  Icon,
  Layouts,
  Router,
  Select,
  TextButton,
  TextField
} from '~/components'
import {
  getBalancesForTransfer,
  getLinkedAccountsForWithdraw,
  type FormattedLinkedAccount
} from '~/data/accounts.server'
import type {
  Balance,
  Amount as RpcAmount
} from '~/generated/connect/backend/v1/backend_pb'
import { jsonWithCSRF, validateCSRFToken } from '~/lib/csrf.server'
import { isConnectError } from '~/lib/error.server'
import { grpc } from '~/lib/grpc.server'
import { mergeMeta } from '~/lib/meta'
import { useScaffoldStore } from '~/lib/useScaffoldStore'
import { PaySelect } from '~/routes/pay_.$paymentId/PaySelect'
import styles from '~/styles/flags.css'

export async function loader(args: LoaderFunctionArgs) {
  const providerResponse = await grpc.getOnOffRampProvider(args.request, {})
  if (isConnectError(providerResponse)) throw providerResponse.error

  if (providerResponse.provider == 'gatehub') {
    return gatehubWithdrawalLoader(args)
  } else return fynbosWithdrawalLoader(args)
}

async function gatehubWithdrawalLoader({ request }: LoaderFunctionArgs) {
  const widgetResponse = await grpc.getGatehubWithdrawalWidget(request, {})
  if (isConnectError(widgetResponse)) throw widgetResponse.error

  return jsonWithCSRF(request, {
    provider: 'gatehub',
    gatehubWidgetUrl: widgetResponse.widgetUrl
  })
}

async function fynbosWithdrawalLoader({ request }: LoaderFunctionArgs) {
  const url = new URL(request.url)

  const balanceResponse = await grpc.getBalances(request, {})
  if (isConnectError(balanceResponse)) throw balanceResponse.error

  const balances = await getBalancesForTransfer(request)

  const linkedAccount = url.searchParams.get('linkedAccount')
  let balanceAccount: Balance | undefined
  let balance: FormattedLinkedAccount | undefined
  if (linkedAccount) {
    balanceAccount = balanceResponse.balances.find(
      (acc) => acc.linkedAccount == linkedAccount
    )
  } else {
    balanceAccount = balanceResponse.balances[0]
  }

  balance = balances.find((acc) => acc.id == balanceAccount?.linkedAccount)
  if (!balanceAccount) throw json({}, { status: 404 })

  const linkedAccounts = await getLinkedAccountsForWithdraw(
    request,
    balanceAccount.linkedAccount
  )

  return jsonWithCSRF(request, {
    provider: 'fynbos',
    balanceAccount,
    balance,
    balances,
    linkedAccounts
  })
}

export const handle: ApplicationProps = {
  layout: Layouts.Focus,
  scaffold: {
    header: {
      title: 'Withdraw',
      back: '/'
    }
  }
}

export const meta: MetaFunction = mergeMeta(() => [
  {
    title: 'Withdraw'
  }
])

export function links() {
  return [{ rel: 'stylesheet', href: styles }]
}

export default function Page() {
  const { provider } = useLoaderData<typeof loader>()

  if (provider == 'gatehub') {
    return <GatehubWithdrawalPage />
  } else return <FynbosWithdrawalPage />
}

function GatehubWithdrawalPage() {
  const submit = useSubmit()
  const { gatehubWidgetUrl } = useLoaderData<typeof gatehubWithdrawalLoader>()

  useEffect(() => {
    if (window) {
      console.log('registering message event handler')
      let url = new URL(gatehubWidgetUrl)
      window.addEventListener('message', (event) => {
        console.log('received message')
        console.log('origin', event.origin)
        console.log('data', event.data)

        if (
          event.origin == url.origin &&
          event.data.type == 'WithdrawalCompleted'
        ) {
          let formData = new FormData()
          formData.append('provider', 'gatehub')
          formData.append('withdrawalId', event.data.uuid)

          submit(formData, {
            action: '/withdraw',
            method: 'post'
          })
        }
      })
    }
  })

  return (
    <iframe
      title='Withdraw'
      src={gatehubWidgetUrl}
      sandbox='allow-top-navigation allow-forms allow-same-origin allow-popups allow-scripts'
      scrolling='no'
      frameBorder='0'
      className='h-[750px]'
    />
  )
}

function FynbosWithdrawalPage() {
  const { linkedAccounts } = useLoaderData<typeof fynbosWithdrawalLoader>()

  const [setLoading] = useScaffoldStore((state) => [state.setLoading])

  useEffect(() => {
    // This ensures that loading is false when this route is unmounted.
    return () => {
      setLoading(false)
    }
  }, [setLoading])

  if (linkedAccounts.length === 0)
    return (
      <Card>
        <CardContent>
          <div className='flex items-start space-x-4'>
            <CardIcon>
              <Icon>account_balance</Icon>
            </CardIcon>
            <div className='flex flex-col space-y-4'>
              <p className='text-sm text-medium'>
                To withdraw from your balance, first connect a bank account.
              </p>
              <Router
                prefetch='render'
                className='text-sm font-medium text-primary'
                to={route('/accounts')}
              >
                Go to accounts page
              </Router>
            </div>
          </div>
        </CardContent>
      </Card>
    )

  return <Amount />
}

const formatAmount = (amount?: PlainMessage<RpcAmount>): string => {
  if (typeof amount == 'undefined') return ''
  if (amount.amount == 0n) return ''

  const floatAmount = Number(amount.amount) / 100
  const formattedAmount = floatAmount.toFixed(amount.assetScale)
  return formattedAmount.replace('.00', '')
}

const Amount = () => {
  const { balance, balances, balanceAccount, linkedAccounts, csrfToken } =
    useLoaderData<typeof fynbosWithdrawalLoader>()
  const [, setSearchParams] = useSearchParams()
  const actionData = useActionData<typeof action>()

  const [amount, setAmount] = useState<string>('')

  const [bank, setBank] = useState<SelectOptions>(linkedAccounts[0])

  const _onChangeWithdrawAmount = useCallback<
    ChangeEventHandler<HTMLInputElement>
  >((event) => {
    setAmount(event.target.value)
  }, [])

  const _maxWithdrawAmount = useCallback(() => {
    setAmount(formatAmount(balanceAccount.balance))
  }, [balanceAccount.balance])

  const _onChangeLinkedAccount = useCallback(
    (linkedAccount: FormattedLinkedAccount) => {
      setSearchParams(
        (prev) => {
          prev.set('linkedAccount', linkedAccount.id)
          return prev
        },
        { replace: true }
      )
    },
    [setSearchParams]
  )

  return (
    <>
      <Form
        id='account-withdraw'
        action={route('/withdraw')}
        method='post'
        className='hidden'
      />
      <input
        form='account-withdraw'
        value={csrfToken}
        name='csrfToken'
        type='hidden'
      />
      <input
        form='account-withdraw'
        value={bank.id}
        name='toLinkedAccount'
        type='hidden'
      />
      <input
        form='account-withdraw'
        value={balanceAccount.linkedAccount}
        name='fromLinkedAccount'
        type='hidden'
      />
      <Card>
        <PaySelect
          id='withdrawAmount'
          label='Withdraw amount'
          name='withdrawAmount'
          form='account-withdraw'
          value={amount}
          onChange={_onChangeWithdrawAmount}
          linkedAccount={balance}
          linkedAccountOptions={balances || []}
          onChangeLinkedAccount={_onChangeLinkedAccount}
          placeholder='0.00'
          prefixIcon={<div className={`flag:${balanceAccount.countryCode}`} />}
          type='number'
          min='0'
          step='0.01'
          aria-invalid={
            Boolean(actionData?.errors?.withdrawAmount) || undefined
          }
          aria-describedby={
            actionData?.errors?.withdrawAmount ? 'amount-error' : undefined
          }
          errorMessage={actionData?.errors?.withdrawAmount || undefined}
        />
        <CardContent className='mt-2 flex flex-col gap-y-4'>
          <span>
            You have{' '}
            <TextButton onClick={_maxWithdrawAmount}>
              {balanceAccount.formattedBalance}
            </TextButton>{' '}
            available in your balance.
          </span>
          <div className='flex flex-col gap-y-1'>
            <div className='flex w-full justify-between'>
              <span className='text-weak'>Fees</span>
              <span className='text-medium'>0.00</span>
            </div>
            <span className='text-xs text-weak'>
              For a limited time, the Interledger Wallet will absorb all fees.
            </span>
          </div>
        </CardContent>
      </Card>
      <Card>
        <Select
          id='bank'
          label='Withdraw to'
          value={bank}
          options={linkedAccounts}
          onChange={setBank}
        />
      </Card>
      <Card>
        <TextField
          id='note'
          label='Withdraw note'
          name='note'
          form='account-withdraw'
          type='text'
          aria-invalid={Boolean(actionData?.errors?.note) || undefined}
          aria-describedby={
            actionData?.errors?.note ? 'reference-error' : undefined
          }
          errorMessage={actionData?.errors?.note}
        />
      </Card>
      <Button type='submit' form='account-withdraw'>
        Continue
      </Button>
    </>
  )
}

function stringToBigInt(amount: string) {
  if (amount == '') return BigInt(0)
  const dotIndex = amount.lastIndexOf('.')
  if (dotIndex > -1) {
    const amounts = amount.split('.')
    return BigInt(amounts[0] + amounts[1].slice(0, 2).padEnd(2, '0'))
  }
  return BigInt(parseFloat(amount) * 100)
}

export async function action({ request }: ActionFunctionArgs) {
  const form = await request.formData()
  const withdrawAmount = String(form.get('withdrawAmount') || '')
  const toLinkedAccount = form.get('toLinkedAccount') as string
  const fromLinkedAccount = form.get('fromLinkedAccount') as string
  const note = form.get('note') as string

  await validateCSRFToken(request, form)

  if ((form.get('provider') as string) == 'gatehub') {
    return createGatehubWithdrawal(request, form)
  }

  // TODO This needs a mapping
  const errors = {
    form: '',
    withdrawAmount: '',
    toLinkedAccount: '',
    note: ''
  }

  const withdrawResponse = await grpc.withdrawBalance(request, {
    fromLinkedAccount: fromLinkedAccount,
    toLinkedAccount: toLinkedAccount,
    amount: {
      amount: stringToBigInt(withdrawAmount),
      asset: 'ZAR',
      assetScale: 2
    },
    note
  })
  if (isConnectError(withdrawResponse)) {
    if (withdrawResponse.code == Code.InvalidArgument) {
      return withdrawResponse.error({ errors })
    }
    if (
      withdrawResponse.code == Code.FailedPrecondition &&
      withdrawResponse.violations.findIndex(
        (violation) =>
          violation.type === 'Payment' &&
          violation.subject === 'insufficientFunds'
      ) > -1
    ) {
      return withdrawResponse.error({
        errors: {
          ...errors,
          withdrawAmount: 'You have insufficient funds available.'
        }
      })
    }
    errors.form = 'Failed to create withdrawal.'
    return withdrawResponse.error(
      { errors },
      {},
      {
        action: 'Contact support'
      }
    )
  }

  return redirect(
    route('/withdraw/:paymentId', {
      paymentId: withdrawResponse.id
    })
  )
}

async function createGatehubWithdrawal(request: Request, formData: FormData) {
  const withdrawResponse = await grpc.createGatehubWithdrawal(request, {
    externalTransactionId: formData.get('withdrawalId') as string
  })
  if (isConnectError(withdrawResponse)) {
    throw withdrawResponse.error
  }

  return redirect(
    route('/transactions/:transactionId', {
      transactionId: withdrawResponse.transactionId
    })
  )
}
