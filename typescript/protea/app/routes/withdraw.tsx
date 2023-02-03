import type { ChangeEventHandler } from 'react'
import { useCallback, useState } from 'react'
import type { ActionArgs, LoaderArgs } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { useFetcher, useLoaderData } from '@remix-run/react'
import type { SelectOptions } from '~/components'
import { Button, Icon, Layouts, Select, TextField } from '~/components'
import { flowType, requireFlow, updateFlow } from '~/lib/flows.server'
import { route } from 'routes-gen'
import { getLinkedAccounts, getWalletBalance } from '~/lib/wallet.server'
import { v4 as uuidv4 } from 'uuid'

export async function loader({ request }: LoaderArgs) {
  const { canWithdraw, linkedAccounts } = await getLinkedAccounts(request)
  if (!canWithdraw)
    return redirect(route('/linked-account/:type', { type: 'bank' }))

  const flow = await requireFlow(request, flowType.Withdraw)
  return json({
    flow,
    balance: await getWalletBalance(request),
    linkedAccounts: linkedAccounts.filter((account) => account.type == 'bank')
  })
}

export const handle = {
  layout: Layouts.FocusLayout
}

export default function Page() {
  const { flow, balance, linkedAccounts } = useLoaderData<typeof loader>()
  const fetcher = useFetcher()

  const [linkedAccount, setLinkedAccount] = useState<{
    id: string
    name: string
  }>(linkedAccounts[0])

  const _onChangeLinkedAccount = useCallback((event: SelectOptions) => {
    setLinkedAccount(event)
  }, [])

  const _onChangeInput = useCallback<ChangeEventHandler<HTMLInputElement>>(
    (event) => {
      let amount = event.target.value
      fetcher.submit({ amount: amount }, { method: 'post', replace: false })
    },
    [fetcher]
  )

  return (
    <>
      <div className='flex w-full flex-col rounded-2xl bg-page p-4 pb-8'>
        <h1 className='mb-6 font-display text-2xl font-medium'>Withdraw</h1>
        <span>Enter the amount to withdraw.</span>
        <fetcher.Form
          id='amount-form'
          action='/withdraw'
          method='post'
          className='hidden'
        />
        <TextField
          id='amount'
          form='amount-form'
          label='Amount'
          name='amount'
          defaultValue={flow?.data.amount}
          onChange={_onChangeInput}
          prefixIcon={<Icon className='text-medium'>attach_money</Icon>}
          type='number'
          min='0'
          step='0.01'
          className='mt-6'
          aria-invalid={Boolean(fetcher.data?.errors.amount) || undefined}
          aria-describedby={
            fetcher.data?.errors.amount ? 'amount-error' : undefined
          }
          errorMessage={fetcher.data?.errors.amount || undefined}
          required
        />

        <div className='mt-4 flex items-center justify-between rounded-xl bg-container p-4 text-medium'>
          <span className='text-sm'>Available cash balance</span>
          <span className='text-sm font-medium'>{balance}</span>
        </div>

        <Select
          id='linkedAccount'
          label='Withdraw to'
          className='mt-12'
          value={linkedAccount}
          options={linkedAccounts}
          onChange={_onChangeLinkedAccount}
          aria-invalid={
            Boolean(fetcher.data?.errors.linkedAccount) || undefined
          }
          aria-describedby={
            fetcher.data?.errors.linkedAccount
              ? 'linkedAccount-error'
              : undefined
          }
          errorMessage={fetcher.data?.errors.linkedAccount}
        />
        <input
          form='amount-form'
          value={linkedAccount.id}
          name='toLinkedAccountId'
          type='hidden'
        />

        <div className='mt-4 flex w-full justify-between'>
          <span className='text-sm'>Total fees</span>
          <span className='text-sm font-medium text-strong'>
            Free<sup>*</sup>
          </span>
        </div>
        <div className='mt-4 flex w-full justify-between'>
          <span className='text-sm'>You receive</span>
          <span className='text-sm text-2xl font-medium text-strong'>
            {flow?.data.displayReceiveAmount || '$ 0.00'}
          </span>
        </div>
        <div className='mt-8'>
          <Button form='amount-form' type='submit' name='route-to' value='next'>
            Continue
          </Button>
        </div>
      </div>
      <div className='mt-6 flex w-full space-x-2'>
        <span className='text-xs text-medium'>*</span>
        <span className='text-xs text-medium'>
          For a limited time, Fynbos will absorb the fees associated with making
          a withdrawal.
        </span>
      </div>
    </>
  )
}

export async function action({ request }: ActionArgs) {
  const form = await request.formData()
  const amount = form.get('amount') as string
  const toLinkedAccountId = form.get('toLinkedAccountId') as string
  const amountToSubmit = String(Math.floor(parseFloat(amount) * 100))
  const routeTo = form.get('route-to')

  const fieldErrors = {
    form: '',
    amount: '',
    toLinkedAccountId: '',
    note: ''
  }

  if (amountToSubmit == 'NaN') {
    fieldErrors.amount = 'Amount is required.'
    return json({ errors: { ...fieldErrors } }, { status: 400 })
  }

  // TODO: get version of balance that isn't formatted.
  const balance = parseInt(
    (await getWalletBalance(request)).replace('$ ', '').replace('.', '')
  )

  let receiveAmount = amountToSubmit,
    fee = 0

  const data = {
    errors: { ...fieldErrors },
    amount: amount,
    fee: fee,
    toLinkedAccountId,
    displayFee: formatMoney(fee),
    receiveAmount,
    displayReceiveAmount: formatMoney(
      parseFloat(receiveAmount as string) / 100
    ),
    idempotencyKey: uuidv4()
  }

  await updateFlow(request, flowType.Withdraw, data)

  if (balance < parseInt(amountToSubmit)) {
    data.errors.amount = "Amount can't exceed your balance."
    return json(data, { status: 400 })
  }

  // TODO: should always return data, because using fetcher means redirecting from here doesn't add the route to the stack which breaks the back button.
  if (routeTo == 'next') {
    return redirect(route('/withdraw/confirm'))
  } else {
    return json(data)
  }
}

const formatMoney = (value: number): string => {
  return `$ ${value.toFixed(2)}`
}
