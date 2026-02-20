import { useFetcher, useLoaderData } from '@remix-run/react'
import type { ChangeEventHandler } from 'react'
import { useCallback, useEffect, useReducer, useRef } from 'react'
import { route } from 'routes-gen'
import {
  Button,
  Card,
  CardContent,
  Icon,
  SelectRouter,
  TextField
} from '~/components'
import { PayStep, usePayStore } from '~/lib/usePayStore'

import type { PlainMessage } from '@bufbuild/protobuf/dist/types/message'
import type { FormattedLinkedAccount } from '~/data/accounts.server'
import type {
  Payment,
  Amount as RpcAmount
} from '~/generated/connect/backend/v1/backend_pb'
import { useScaffoldStore } from '~/lib/useScaffoldStore'
import { PayTextField } from '~/routes/pay_.$paymentId/PayTextField'
import { PaySelect } from './PaySelect'
import { PaymentDetailsCard } from './PaymentDetailsCard'
import type { loader } from './route'

const DEBOUNCE_WAIT = 150

const formatAmount = (amount?: PlainMessage<RpcAmount>): string => {
  if (typeof amount == 'undefined') return ''
  if (amount.amount == 0n) return ''

  const floatAmount = Number(amount.amount) / 100
  const formattedAmount = floatAmount.toFixed(amount.assetScale)
  return formattedAmount.replace('.00', '')
}

type AmountState = {
  focussed: 'send' | 'receive' | 'none'
  send: string
  receive: string
  linkedAccount: FormattedLinkedAccount
  totalSendAmount: string
}

type Action = {
  type: 'focussed' | 'send' | 'receive' | 'linkedAccount' | 'network'
} & Partial<AmountState>

type InitialState = {
  account: FormattedLinkedAccount
  payment: PlainMessage<Payment>
}

function createInitialState({ account, payment }: InitialState): AmountState {
  const send = formatAmount(payment.senderAmount)
  const receive = formatAmount(payment.receiverAmount)
  return {
    focussed: 'none',
    send,
    receive,
    linkedAccount: account,
    totalSendAmount: payment.totalSendAmount
  }
}

function reducer(state: AmountState, action: Action): AmountState {
  switch (action.type) {
    case 'focussed':
      return { ...state, focussed: action.focussed! }
    case 'send':
      return { ...state, send: action.send! }
    case 'receive':
      return { ...state, receive: action.receive! }
    case 'linkedAccount':
      return { ...state, linkedAccount: action.linkedAccount! }
    case 'network':
      let newSend = state.send
      let newReceive = state.receive
      if (state.focussed == 'send') {
        newReceive = action.receive!
      } else if (state.focussed == 'receive') {
        newSend = action.send!
      } else if (state.focussed == 'none') {
        newSend = action.send!
        newReceive = action.receive!
      }
      return {
        ...state,
        send: newSend,
        receive: newReceive,
        totalSendAmount: action.totalSendAmount!
      }
    default:
      return state
  }
}

export const Amount = () => {
  const { account, sendAccounts, payment, csrfToken } =
    useLoaderData<any>()

  const [localPayment, dispatchPayment] = useReducer(
    reducer,
    { account, payment },
    createInitialState
  )

  const updatePaymentFetcher = useFetcher<any>()

  const [setStep] = usePayStore((state) => [state.setStep])

  const [setLoading] = useScaffoldStore((state) => [state.setLoading])

  const _onChangeLinkedAccount = useCallback(
    (linkedAccount: FormattedLinkedAccount) => {
      dispatchPayment({ type: 'linkedAccount', linkedAccount })
      updatePaymentFetcher.submit(
        {
          formName: 'updatePayment',
          accountId: linkedAccount.id,
          csrfToken
        },
        { method: 'post' }
      )
    },
    [csrfToken, updatePaymentFetcher]
  )

  let onChangeSendAmountDismissRef = useRef<NodeJS.Timeout>()
  const _onChangeSendAmount = useCallback<ChangeEventHandler<HTMLInputElement>>(
    (event) => {
      let send = event.target.value
      dispatchPayment({ type: 'send', send })

      if (onChangeSendAmountDismissRef.current) {
        clearTimeout(onChangeSendAmountDismissRef.current)
      }
      onChangeSendAmountDismissRef.current = setTimeout(() => {
        updatePaymentFetcher.submit(
          {
            formName: 'updatePayment',
            send,
            accountId: localPayment.linkedAccount.id,
            sendCurrency: payment.senderAmount?.asset || '',
            csrfToken
          },
          { method: 'post' }
        )
      }, DEBOUNCE_WAIT)
    },
    [
      csrfToken,
      localPayment.linkedAccount.id,
      updatePaymentFetcher,
      payment.senderAmount?.asset
    ]
  )

  let onChangeReceiveAmountDismissRef = useRef<NodeJS.Timeout>()
  const _onChangeReceiveAmount = useCallback<
    ChangeEventHandler<HTMLInputElement>
  >(
    (event) => {
      let receive = event.target.value
      dispatchPayment({ type: 'receive', receive })

      if (onChangeReceiveAmountDismissRef.current) {
        clearTimeout(onChangeReceiveAmountDismissRef.current)
      }
      onChangeReceiveAmountDismissRef.current = setTimeout(() => {
        updatePaymentFetcher.submit(
          {
            formName: 'updatePayment',
            receive,
            receiveCurrency: payment.receiverAmount?.asset || '',
            accountId: localPayment.linkedAccount.id,
            csrfToken
          },
          { method: 'post' }
        )
      }, DEBOUNCE_WAIT)
    },
    [
      csrfToken,
      localPayment.linkedAccount.id,
      payment.receiverAmount?.asset,
      updatePaymentFetcher
    ]
  )

  useEffect(() => {
    return () => {
      if (onChangeSendAmountDismissRef.current) {
        clearTimeout(onChangeSendAmountDismissRef.current)
      }
      if (onChangeReceiveAmountDismissRef.current) {
        clearTimeout(onChangeReceiveAmountDismissRef.current)
      }
    }
  }, [])

  useEffect(() => {
    if (
      updatePaymentFetcher.data?.intent == 'submit' &&
      updatePaymentFetcher.state == 'idle'
    ) {
      setStep(PayStep.CONFIRM)
    }
    if (updatePaymentFetcher.state !== 'idle') {
      setLoading(true)
    } else {
      setLoading(false)
    }
  }, [
    setLoading,
    setStep,
    updatePaymentFetcher.data?.intent,
    updatePaymentFetcher.state
  ])

  useEffect(() => {
    if (updatePaymentFetcher.data?.payment) {
      dispatchPayment({
        type: 'network',
        send: formatAmount(updatePaymentFetcher.data?.payment?.senderAmount),
        receive: formatAmount(
          updatePaymentFetcher.data?.payment?.receiverAmount
        ),
        totalSendAmount: updatePaymentFetcher.data?.payment?.totalSendAmount
      })
    }
  }, [localPayment.focussed, updatePaymentFetcher.data?.payment])

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
        value={localPayment.linkedAccount.id}
        form='amount-form'
      />
      <input
        type='hidden'
        name='sendCurrency'
        value={payment.senderAmount?.asset || ''}
        form='amount-form'
      />
      <input
        type='hidden'
        name='receiveCurrency'
        value={payment.receiverAmount?.asset || ''}
        form='amount-form'
      />
      <PaymentDetailsCard />
      <Card>
        <PaySelect
          id='send'
          label='Amount to send'
          name='send'
          form='amount-form'
          onFocus={() =>
            dispatchPayment({ type: 'focussed', focussed: 'send' })
          }
          value={localPayment.send}
          onChange={_onChangeSendAmount}
          linkedAccount={localPayment.linkedAccount}
          linkedAccountOptions={sendAccounts || []}
          onChangeLinkedAccount={_onChangeLinkedAccount}
          selectButton={
            <SelectRouter to={route('/accounts')}>
              <span>Connect new account</span> <Icon>add</Icon>
            </SelectRouter>
          }
          placeholder='0.00'
          prefixIcon={
            <div
              className={`flag:${localPayment.linkedAccount.sendCurrencyCountryCode}`}
            />
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
        />
        <CardContent className='mt-2 flex flex-col gap-y-4'>
          <div className='flex flex-col gap-y-1'>
            <div className='flex w-full justify-between'>
              <span className='text-weak'>Fees</span>
              <span className='text-medium'>{payment.formattedFees}</span>
            </div>
            <span className='text-xs text-weak'>
              For a limited time, the Interledger Wallet will absorb all fees.
            </span>
          </div>
          <div className='flex w-full justify-between'>
            <span className='font-medium text-medium'>
              Total amount to debit
            </span>
            <span className='font-medium text-error'>
              {localPayment.totalSendAmount}
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
          placeholder='0.00'
          onFocus={() =>
            dispatchPayment({ type: 'focussed', focussed: 'receive' })
          }
          value={localPayment.receive}
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
