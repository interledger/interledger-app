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
  useSearchParams
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
  Alert,
  AlertBody,
  AlertContent,
  AlertTitle,
  Button,
  Card,
  CardContent,
  CardHeader,
  CardIcon,
  CardTitle,
  Icon,
  Layouts,
  Router,
  Select,
  TextButton,
  TextField
} from '~/components'
import {
  getBalancesForTransfer,
  getLinkedAccountsForDeposit,
  type FormattedLinkedAccount
} from '~/data/accounts.server'
import type {
  Balance,
  Amount as RpcAmount,
  XagoDepositDetails
} from '~/generated/connect/backend/v1/backend_pb'
import { jsonWithCSRF, validateCSRFToken } from '~/lib/csrf.server'
import { isConnectError } from '~/lib/error.server'
import { grpc } from '~/lib/grpc.server'
import { mergeMeta } from '~/lib/meta'
import { useScaffoldStore } from '~/lib/useScaffoldStore'
import { PaySelect } from '~/routes/pay_.$paymentId/PaySelect'
import styles from '~/styles/flags.css'

export async function loader({ request }: LoaderFunctionArgs) {
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
  if (!balanceAccount) throw json({}, { status: 404 })

  const linkedAccounts = await getLinkedAccountsForDeposit(
    request,
    balanceAccount.linkedAccount
  )

  if (linkedAccounts.length == 0) {
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
    balanceAccount,
    balance,
    balances,
    linkedAccounts,
    depositDetails
  })
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
  const { depositDetails } = useLoaderData<typeof loader>()

  const [setLoading] = useScaffoldStore((state) => [state.setLoading])

  useEffect(() => {
    // This ensures that loading is false when this route is unmounted.
    return () => {
      setLoading(false)
    }
  }, [setLoading])

  if (depositDetails) return <DepositDetails />

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
    useLoaderData<typeof loader>()
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

  // TODO Need to make this dynamic based on jurisdiction capabilities
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
              <span className='text-medium'>R 0.00</span>
            </div>
            <span className='text-xs text-weak'>
              For a limited time, Fynbos will absorb all fees.
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

function DepositDetails() {
  const { depositDetails } = useLoaderData<typeof loader>()

  return (
    <>
      <Card>
        <CardHeader>
          <CardTitle>EFT details</CardTitle>
        </CardHeader>
        <CardContent>
          <span className='mt-4'>Arrives 1-2 business days.</span>
          <div className='mt-4 flex w-full flex-col justify-between space-y-1'>
            <span className='text-weak'>Bank</span>
            <span className='text-medium'>{depositDetails?.bankName}</span>
          </div>
          <div className='mt-4 flex w-full flex-col space-y-1'>
            <span className='text-weak'>Branch code</span>
            <span className='text-medium'>{depositDetails?.branchCode}</span>
          </div>
          <div className='mt-4 flex w-full flex-col space-y-1'>
            <span className='text-weak'>Account number</span>
            <span className='text-medium'>{depositDetails?.accountNumber}</span>
          </div>
          <div className='mt-4 flex w-full flex-col space-y-1'>
            <span className='text-weak'>Reference</span>
            <span className='text-medium'>
              {depositDetails?.depositReference}
            </span>
          </div>
        </CardContent>
      </Card>
      <Alert>
        <Icon>error</Icon>
        <AlertContent className='items-start'>
          <AlertTitle>Important</AlertTitle>
          <AlertBody>
            Use the reference above when depositing for secure and faster
            processing.
          </AlertBody>
        </AlertContent>
      </Alert>
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
