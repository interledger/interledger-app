import { Form, useActionData, useLoaderData } from 'react-router';
import { href } from 'react-router'
import { Button, Card, CardContent, Checkbox } from '~/components'

import { DateTime } from 'luxon'
import { usePTISdk } from '~/lib/usePTISdk'
import { PaymentDetailsCard } from './PaymentDetailsCard'
import type { loader } from './route'
import { confirmPaymentAction } from './action.server';


export function Confirm() {
  const { payment, account, csrfToken, PTIClientId } =
    useLoaderData<typeof loader>()
  const actionData = useActionData<typeof confirmPaymentAction>()

  usePTISdk(payment.id, PTIClientId)

  return (
    <>
      <Form
        id='pay-confirm'
        action={href('/pay/:paymentId', { paymentId: payment.id })}
        method='post'
        className='hidden'
      />
      <input
        form='pay-confirm'
        value={csrfToken}
        name='csrfToken'
        type='hidden'
      />
      <PaymentDetailsCard />
      <Card>
        <CardContent>
          <div className='flex w-full justify-between'>
            <span className='text-weak'>Payment from</span>
            <span className='text-medium'>
              {account.name} {account.mask}
            </span>
          </div>
          <div className='mt-2 flex w-full justify-between'>
            <span className='text-weak'>Payment date</span>
            <span className='text-medium'>
              {DateTime.now().toFormat('dd MMMM yyyy')}
            </span>
          </div>
          {/* TODO Convert amount to currency*/}
          {/*<div className='mt-2 flex w-full justify-between'>*/}
          {/*  <span className='text-weak'>Amount to send</span>*/}
          {/*  <span className='text-medium'>{payment.senderAmount?.amount}</span>*/}
          {/*</div>*/}
          <div className='mt-2 flex w-full justify-between'>
            <span className='text-weak'>Fees</span>
            <span className='text-medium'>{payment.formattedFees}</span>
          </div>
          <div className='mt-4 flex w-full justify-between font-medium'>
            <span className='text-medium'>Total amount to debit</span>
            <span className='text-error'>{payment.totalSendAmount}</span>
          </div>
        </CardContent>
      </Card>
      {payment.note && (
        <Card>
          <CardContent>
            <div className='flex w-full flex-col space-y-1'>
              <span className='text-weak'>Payment note</span>
              <span className='text-medium'>{payment.note}</span>
            </div>
          </CardContent>
        </Card>
      )}
      <Card>
        <CardContent>
          <Checkbox
            id='service-agreement'
            name='serviceAgreement'
            form='pay-confirm'
            data-testid='pay-confirm-agreement'
            className='flex'
            aria-invalid={
              Boolean(actionData?.errors?.serviceAgreement) || undefined
            }
            aria-describedby={
              actionData?.errors?.serviceAgreement
                ? 'serviceAgreement-error'
                : undefined
            }
            errorMessage={actionData?.errors?.serviceAgreement}
          >
            I authorize Interledger Wallet to debit
            {account?.type == 'card' ? ' the card indicated ' : ' my account '}
            for the amount noted on today’s date. I will not dispute Interledger
            Wallet debiting my account, so long as the transaction corresponds
            to the terms in this online form and my agreement with Interledger
            Wallet.
          </Checkbox>
        </CardContent>
      </Card>

      <Button
        form='pay-confirm'
        name='formName'
        value='confirmPayment'
        type='submit'
        data-testid='pay-confirm-submit'
      >
        Confirm payment
      </Button>
    </>
  )
}
