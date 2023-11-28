import { Code } from '@bufbuild/connect'
import type {
  ActionFunctionArgs,
  LoaderFunctionArgs,
  MetaFunction
} from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { Form, useActionData, useLoaderData, useParams } from '@remix-run/react'
import { useEffect, useState } from 'react'
import { route } from 'routes-gen'
import type { ApplicationProps, SelectOptions } from '~/components'
import {
  Button,
  Card,
  CardContent,
  Layouts,
  Select,
  TextField
} from '~/components'
import { jsonWithCSRF, validateCSRFToken } from '~/lib/csrf.server'
import { isConnectError } from '~/lib/error.server'
import { grpc } from '~/lib/grpc.server'
import { mergeMeta } from '~/lib/meta'
import { useScaffoldStore } from '~/lib/useScaffoldStore'
import { PayTextField } from '~/routes/pay_.$paymentId/PayTextField'
import styles from '~/styles/flags.css'

export async function loader({ request, params }: LoaderFunctionArgs) {
  const balanceResponse = await grpc.getXagoBalances(request, {})
  if (isConnectError(balanceResponse)) throw balanceResponse.error

  const balanceAccount = balanceResponse.balances.find(
    (balance) => balance.linkedAccount == params.accountId
  )
  if (!balanceAccount) throw json({}, { status: 404 })

  // TODO Get linked accounts that can be withdrawn to
  const linkedAccountsResponse = await grpc.getLinkedAccounts(request, {})
  if (isConnectError(linkedAccountsResponse)) throw linkedAccountsResponse.error
  const linkedAccounts = linkedAccountsResponse.linkedAccounts.filter(
    (account) => account.type.includes('bank')
  )
  // return jsonWithCSRF(request, {
  //   formattedAvailableBalance: 'R 2000.00',
  //   linkedAccounts: linkedAccountsResponse.linkedAccounts
  // })

  return jsonWithCSRF(request, {
    formattedAvailableBalance: balanceAccount.formattedAvailableBalance,
    linkedAccounts
  })
}

export const handle: ApplicationProps = {
  layout: Layouts.Focus,
  scaffold: {
    header: {
      title: 'Withdraw',
      back: '/accounts'
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
  const { formattedAvailableBalance, linkedAccounts, csrfToken } =
    useLoaderData<typeof loader>()
  const actionData = useActionData<typeof action>()
  const params = useParams()

  const [bank, setBank] = useState<SelectOptions>(linkedAccounts[0])
  const [setLoading] = useScaffoldStore((state) => [state.setLoading])

  useEffect(() => {
    // This ensures that loading is false when this route is unmounted.
    return () => {
      setLoading(false)
    }
  }, [setLoading])

  return (
    <>
      <Form
        id='account-withdraw'
        action={route('/accounts/:accountId/withdraw', {
          accountId: params.accountId as string
        })}
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
      <Card>
        <CardContent className='flex justify-between'>
          <span className='text-weak'>Current balance</span>
          <span className='font-medium'>{formattedAvailableBalance}</span>
        </CardContent>
      </Card>
      <Card>
        <PayTextField
          id='withdrawAmount'
          label='Withdraw amount'
          name='withdrawAmount'
          form='account-withdraw'
          placeholder='0.00'
          prefixIcon={
            // TODO Need to get the country code from the linked account
            <div className={`flag:ZA`} />
          }
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

export async function action({ request, params }: ActionFunctionArgs) {
  const form = await request.formData()
  const withdrawAmount = String(form.get('withdrawAmount') || '')
  const toLinkedAccount = form.get('toLinkedAccount') as string
  const note = form.get('note') as string

  await validateCSRFToken(request, form)

  const errors = {
    form: '',
    withdrawAmount: '',
    toLinkedAccount: '',
    note: ''
  }

  const withdrawResponse = await grpc.withdrawXagoBalance(request, {
    fromLinkedAccount: params.accountId as string,
    toLinkedAccount: toLinkedAccount,
    amount: {
      amount: stringToBigInt(withdrawAmount),
      asset: '',
      assetScale: 2
    },
    note
  })
  if (isConnectError(withdrawResponse)) {
    if (withdrawResponse.code == Code.InvalidArgument) {
      return withdrawResponse.error({ errors })
    } else {
      errors.form = 'Failed to create withdrawal.'
      return withdrawResponse.error(
        { errors },
        {},
        { action: 'Contact support' }
      )
    }
  }

  return redirect(
    route('/accounts/:accountId/withdraw/:paymentId', {
      accountId: params.accountId as string,
      paymentId: withdrawResponse.id
    })
  )
}
