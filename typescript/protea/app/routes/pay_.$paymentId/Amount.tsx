import { useFetcher, useLoaderData } from '@remix-run/react'
import type { ChangeEventHandler } from 'react'
import { useCallback, useEffect, useState } from 'react'
import { route } from 'routes-gen'
import type { SelectOptions } from '~/components'
import {
  Button,
  Card,
  CardContent,
  Icon,
  Select,
  SelectRouter,
  TextField
} from '~/components'
import { PayStep, usePayStore } from '~/lib/usePayStore'

import { PaymentDetailsCard } from './PaymentDetailsCard'
import type { loader, updatePaymentAction } from './route'

export const Amount = () => {
  const { account, sendAccounts, payment, csrfToken } =
    useLoaderData<typeof loader>()

  const [localAccount, setLocalAccount] = useState<SelectOptions>(
    account as SelectOptions
  )

  const _onChangeLinkedAccount = useCallback((event: SelectOptions) => {
    setLocalAccount(event)
  }, [])

  const updatePaymentFetcher = useFetcher<typeof updatePaymentAction>()

  const [amount, setAmount, setStep] = usePayStore((state) => [
    state.amount,
    state.setAmount,
    state.setStep
  ])

  const _onChangeAmount = useCallback<ChangeEventHandler<HTMLInputElement>>(
    (event) => {
      let amount = event.target.value
      setAmount(amount)
      updatePaymentFetcher.submit(
        {
          formName: 'updatePayment',
          amount,
          csrfToken
        },
        { method: 'post' }
      )
    },
    [csrfToken, setAmount, updatePaymentFetcher]
  )

  useEffect(() => {
    if (
      updatePaymentFetcher.data?.intent == 'submit' &&
      updatePaymentFetcher.state == 'idle'
    ) {
      setStep(PayStep.CONFIRM)
    }
  }, [
    updatePaymentFetcher,
    updatePaymentFetcher.data?.payment?.id,
    updatePaymentFetcher.data?.intent,
    setStep
  ])

  return (
    <>
      <updatePaymentFetcher.Form
        id='amount-form'
        action={route('/pay/:paymentId', { paymentId: payment.id })}
        method='post'
        className='hidden'
      />
      <input
        type='hidden'
        name='formName'
        value='updatePayment'
        form='amount-form'
      />
      <PaymentDetailsCard />
      <Card>
        <TextField
          id='amount'
          label='Amount'
          name='amount'
          form='amount-form'
          value={amount}
          onChange={_onChangeAmount}
          prefix='$'
          type='number'
          min='0'
          step='0.01'
          aria-invalid={
            Boolean(updatePaymentFetcher.data?.errors?.amount) || undefined
          }
          aria-describedby={
            updatePaymentFetcher.data?.errors?.amount
              ? 'amount-error'
              : undefined
          }
          errorMessage={updatePaymentFetcher.data?.errors?.amount || undefined}
          required
        />
      </Card>
      <Card>
        <CardContent>
          <span>Select an account to pay from:</span>
        </CardContent>
        <Select
          id='linkedAccount'
          label='Connected accounts'
          className='mt-2'
          value={localAccount as SelectOptions}
          options={sendAccounts || []}
          onChange={_onChangeLinkedAccount}
          selectButton={
            <SelectRouter to={route('/accounts')}>
              <span>Connect new account</span> <Icon>add</Icon>
            </SelectRouter>
          }
          aria-invalid={
            Boolean(updatePaymentFetcher.data?.errors?.linkedAccount) ||
            undefined
          }
          aria-describedby={
            updatePaymentFetcher.data?.errors?.linkedAccount
              ? 'linkedAccount-error'
              : undefined
          }
          errorMessage={updatePaymentFetcher.data?.errors?.linkedAccount}
        />
        <input
          type='hidden'
          name='accountId'
          value={localAccount?.id}
          form='amount-form'
        />
        <TextField
          id='note'
          label='Note'
          name='note'
          form='amount-form'
          type='text'
          defaultValue={payment.note}
          className='mt-4'
          aria-invalid={
            Boolean(updatePaymentFetcher.data?.errors?.note) || undefined
          }
          aria-describedby={
            updatePaymentFetcher.data?.errors?.note
              ? 'reference-error'
              : undefined
          }
          errorMessage={updatePaymentFetcher.data?.errors?.note}
        />
      </Card>

      <Button type='submit' form='amount-form' name='intent' value='submit'>
        Continue
      </Button>
    </>
  )
}
