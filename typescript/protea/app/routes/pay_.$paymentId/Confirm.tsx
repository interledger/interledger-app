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
import { usePayStore } from '~/lib/usePayStore'
import type { action as otpAction } from '~/routes/api_.sendOtp'

import { PaymentDetailsCard } from './PaymentDetailsCard'
import type { confirmPaymentAction, loader } from './route'

export function Confirm() {
  const { payment, account, phoneMask, requiresOTP, csrfToken } =
    useLoaderData<typeof loader>()
  const actionData = useActionData<typeof confirmPaymentAction>()
  const otpFetcher = useFetcher<typeof otpAction>()

  const [showOTPDialog, setShowOTPDialog] = useState<boolean>(false)

  const [displayAmount] = usePayStore((state) => [state.displayAmount])

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
            <span className='text-weak'>Total fees</span>
            <span className='text-medium'>
              Free <sup>*</sup>
            </span>
          </div>
          <div className='mt-2 flex w-full justify-between'>
            <span className='text-weak'>They receive</span>
            <span className='text-medium'>{displayAmount}</span>
          </div>
          <div className='mt-4 flex w-full space-x-2'>
            <span className='text-xs text-medium'>*</span>
            <span className='text-xs text-medium'>
              For a limited time, Fynbos will absorb the fees associated with
              making a payment.
            </span>
          </div>
        </CardContent>
      </Card>
      <Card>
        <CardContent>
          <div className='flex w-full flex-col justify-between space-y-1'>
            <span className='text-weak'>Source</span>
            <span className='text-medium'>{account?.name}</span>
          </div>
          {payment.note && (
            <div className='mt-4 flex w-full flex-col space-y-1'>
              <span className='text-weak'>Note</span>
              <span className='text-medium'>{payment.note}</span>
            </div>
          )}
        </CardContent>
      </Card>
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
            I authorize Fynbos to debit
            {account?.type == 'card' ? ' the card indicated ' : ' my account '}
            for the amount noted on today’s date. I will not dispute Fynbos
            debiting my account, so long as the transaction corresponds to the
            terms in this online form and my agreement with Fynbos.
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
