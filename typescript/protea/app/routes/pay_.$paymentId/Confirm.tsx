import {
  Form,
  useActionData,
  useFetcher,
  useLoaderData
} from '@remix-run/react'
import { useEffect, useState } from 'react'
import { route } from 'routes-gen'
import {
  Button,
  Card,
  CardContent,
  CardHeader,
  Checkbox,
  Dialog,
  Icon,
  TextButton,
  TextField
} from '~/components'
import { Label } from '~/components/Label'
import type { action as otpAction } from '~/routes/api_.sendOtp'

import { DateTime } from 'luxon'
import { usePTISdk } from '~/lib/usePTISdk'
import { PaymentDetailsCard } from './PaymentDetailsCard'
import type { confirmPaymentAction, loader } from './route'

export function Confirm() {
  const { payment, account, phoneMask, requiresOTP, csrfToken, PTIClientId } =
    useLoaderData<typeof loader>()
  const actionData = useActionData<typeof confirmPaymentAction>()
  const otpFetcher = useFetcher<typeof otpAction>()

  const [showOTPDialog, setShowOTPDialog] = useState<boolean>(false)

  usePTISdk(payment.id, PTIClientId)

  useEffect(() => {
    if (
      otpFetcher.state == 'submitting' &&
      otpFetcher.formMethod == 'POST' &&
      otpFetcher.formAction == route('/api/sendOtp')
    ) {
      setShowOTPDialog(true)
    }
  }, [otpFetcher.formAction, otpFetcher.formMethod, otpFetcher.state])

  return (
    <>
      <otpFetcher.Form
        id='pay-phone-otp'
        action={route('/api/sendOtp')}
        method='post'
        className='hidden'
      />
      <input
        form='pay-phone-otp'
        value={csrfToken}
        name='csrfToken'
        type='hidden'
      />
      <Form
        id='pay-confirm'
        action={route('/pay/:paymentId', { paymentId: payment.id })}
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
      {requiresOTP && (
        <Button form='pay-phone-otp' type='submit'>
          Confirm payment
        </Button>
      )}
      {!requiresOTP && (
        <Button
          form='pay-confirm'
          name='formName'
          value='confirmPayment'
          type='submit'
        >
          Confirm payment
        </Button>
      )}
      <Dialog open={showOTPDialog} setOpen={setShowOTPDialog}>
        <CardHeader>
          <h1 className='text-xl font-medium'>Two-step verification</h1>
        </CardHeader>
        <CardContent>
          <span className='text-medium'>
            Enter the six digit code sent to your mobile number.
          </span>
        </CardContent>
        <Label className='mt-2'>Your mobile phone number</Label>
        <div className='mt-1 flex space-x-2 rounded-xl bg-nav p-3 text-medium'>
          <Icon>phone_android</Icon>
          <span>{phoneMask}</span>
        </div>

        <TextField
          id='otp'
          form='pay-confirm'
          label='Verification code'
          name='otp'
          type='number'
          className='mt-4'
          aria-invalid={Boolean(actionData?.errors.otp) || undefined}
          aria-describedby={actionData?.errors.otp ? 'email-error' : undefined}
          required
          errorMessage={actionData?.errors.otp}
        />
        <CardContent className='mt-2 flex w-full justify-end space-x-6'>
          <TextButton type='submit' form='pay-phone-otp'>
            Resend code
          </TextButton>
          <TextButton
            form='pay-confirm'
            name='formName'
            value='confirmPayment'
            type='submit'
          >
            Verify
          </TextButton>
        </CardContent>
      </Dialog>
    </>
  )
}
