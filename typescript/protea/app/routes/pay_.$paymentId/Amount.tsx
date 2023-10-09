import { useFetcher, useLoaderData } from '@remix-run/react'
import type { ChangeEventHandler } from 'react'
import { useCallback, useEffect, useState } from 'react'
import { route } from 'routes-gen'
import {
  Button,
  Card,
  CardContent,
  Icon,
  SelectRouter,
  Switch,
  TextField
} from '~/components'
import { PayStep, usePayStore } from '~/lib/usePayStore'

import type { PlainMessage } from '@bufbuild/protobuf/dist/types/message'
import type { FormattedLinkedAccount } from '~/data/wallet.server'
import type { Amount as RpcAmount } from '~/generated/connect/backend/v1/backend_pb'
import { PayTextField } from '~/routes/pay_.$paymentId/PayTextField'
import { PaySelect } from './PaySelect'
import { PaymentDetailsCard } from './PaymentDetailsCard'
import type { loader, updatePaymentAction } from './route'
//
// function reducer(state: AmountState, action: PayAction): AmountState {
//   switch (action.type) {
//     case 'focussed':
//       console.log('Changing focussed')
//       return { ...state, focussed: action.payload.focussed }
//     case 'localSend':
//       console.log('Changing send amount')
//       // The payment was updated
//       // We should update the receive amount
//       // Check that the send amount is the same as what we have
//       return { ...state, send: action.payload.send }
//     case 'networkSend':
//       console.log('Changing receive amount')
//       // The payment was updated
//       // We should update the send amount
//       // Check that the receive amount is the same as what we have
//       return { ...state, receive: action.payload.receive }
//
//     default:
//       console.log('Changing default amount')
//       // This is the initial state or if the user has inputted something
//       return state
//   }
// }

export const Amount = () => {
  const { account, sendAccounts, payment, csrfToken } =
    useLoaderData<typeof loader>()

  const [localAccount, setLocalAccount] = useState<
    FormattedLinkedAccount | undefined
  >(account)

  const [focussed, setFocussed] = useState<'send' | 'receive' | 'none'>('none')

  const [send, setSend] = useState('')
  const [receive, setReceive] = useState('')

  const _onChangeLinkedAccount = useCallback(
    (event: FormattedLinkedAccount) => {
      setLocalAccount(event)
    },
    []
  )

  const updatePaymentFetcher = useFetcher<typeof updatePaymentAction>()

  const [setStep] = usePayStore((state) => [state.setStep])

  const _onChangeAmount = useCallback<ChangeEventHandler<HTMLInputElement>>(
    (event) => {
      let amount = event.target.value
      setSend(amount)
      updatePaymentFetcher.submit(
        {
          formName: 'updatePayment',
          amount,
          csrfToken
        },
        { method: 'post' }
      )
    },
    [csrfToken, updatePaymentFetcher]
  )

  const _onChangeReceiveAmount = useCallback<
    ChangeEventHandler<HTMLInputElement>
  >(
    (event) => {
      let receiveAmount = event.target.value
      setReceive(receiveAmount)
      updatePaymentFetcher.submit(
        {
          formName: 'updatePayment',
          receiveAmount,
          csrfToken
        },
        { method: 'post' }
      )
    },
    [csrfToken, updatePaymentFetcher]
  )

  const _onChangeSwitch = useCallback<{
    (formName: string, publish: boolean): void
  }>(
    (formName, paymentProtection) => {
      updatePaymentFetcher.submit(
        {
          formName,
          paymentProtection: paymentProtection.toString(),
          csrfToken
        },
        { method: 'post' }
      )
    },
    [csrfToken, updatePaymentFetcher]
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

  useEffect(() => {
    if (updatePaymentFetcher.data?.payment) {
      if (focussed == 'send') {
        const amount = formatAmount(
          updatePaymentFetcher.data?.payment?.receiverAmount
        )
        setReceive(amount)
      } else if (focussed == 'receive') {
        const amount = formatAmount(
          updatePaymentFetcher.data?.payment?.senderAmount
        )
        setSend(amount)
      }
    }
  }, [focussed, updatePaymentFetcher.data?.payment])

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
      <input
        type='hidden'
        name='accountId'
        value={localAccount?.id}
        form='amount-form'
      />
      <PaymentDetailsCard />
      <Card>
        <PaySelect
          id='amount'
          label='Amount to send'
          name='amount'
          form='amount-form'
          onFocus={() => setFocussed('send')}
          value={send}
          onChange={_onChangeAmount}
          linkedAccount={localAccount}
          linkedAccountOptions={sendAccounts || []}
          onChangeLinkedAccount={_onChangeLinkedAccount}
          selectButton={
            <SelectRouter to={route('/accounts')}>
              <span>Connect new account</span> <Icon>add</Icon>
            </SelectRouter>
          }
          prefixIcon={
            <div className={`flag:${localAccount?.sendCurrencyCountryCode}`} />
          }
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
        <CardContent className='mt-2 flex flex-col gap-y-4'>
          <div className='flex flex-col gap-y-1'>
            <div className='flex w-full justify-between'>
              <span className='text-weak'>Fees</span>
              <span className='text-medium'>$ 0.00</span>
            </div>
            <span className='text-xs text-weak'>
              For a limited time, Fynbos will absorb all fees.
            </span>
          </div>
          <div className='flex flex-col gap-y-1'>
            <div className='flex w-full justify-between'>
              <span className='text-weak'>Payment protection (+3%)</span>
              <span className='text-medium'>
                {payment.paymentProtectionAmount}
              </span>
            </div>
            <div className='flex w-full gap-x-2'>
              <Switch
                checked={payment.hasPaymentProtection}
                disabled={false}
                onChange={() =>
                  _onChangeSwitch(
                    'updatePayment',
                    !payment.hasPaymentProtection
                  )
                }
              />
              <span className='text-xs text-weak'>
                Add payment protection to safeguard against unexpected
                circumstances.
                {/*<Router className='text-primary' to='/payment-protection'>*/}
                {/*  Find out more.*/}
                {/*</Router>*/}
              </span>
            </div>
          </div>
          {/*<div className='flex flex-col gap-y-1'>*/}
          {/*  <div className='flex w-full justify-between'>*/}
          {/*    <span className='text-weak'>Exchange rate</span>*/}
          {/*    <span className='text-medium'>$ 0.00</span>*/}
          {/*  </div>*/}
          {/*  <span className='text-xs text-weak'>1 USD = 0,94 Euro</span>*/}
          {/*</div>*/}
          <div className='flex w-full justify-between'>
            <span className='font-medium text-medium'>
              Total amount to debit
            </span>
            <span className='font-medium text-error'>
              {payment.totalSendAmount}
            </span>
          </div>
        </CardContent>
      </Card>
      <Card>
        <PayTextField
          id='receiveAmount'
          label='Recipient gets'
          name='receiveAmount'
          form='amount-form'
          onFocus={() => setFocussed('receive')}
          value={receive}
          onChange={_onChangeReceiveAmount}
          prefixIcon={
            <div className={`flag:${payment.receiverAmount?.country}`} />
          }
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
        <TextField
          id='note'
          label='Payment note (optional)'
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

const formatAmount = (amount?: PlainMessage<RpcAmount>): string => {
  if (typeof amount == 'undefined') return ''

  // const floatAmount = Number(amount.amount) / 100
  // const formattedAmount = floatAmount.toFixed(amount.assetScale)
  // return formattedAmount
  return `${(Number(amount.amount) / 100).toFixed(amount.assetScale)}`
}
