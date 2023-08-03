import { useFetcher, useLoaderData } from '@remix-run/react'
import { useEffect, useState } from 'react'
import { route } from 'routes-gen'
import type { PhoneAutocompleteOptions } from '~/components'
import {
  Button,
  Card,
  CardContent,
  CardHeader,
  Dialog,
  Icon,
  PhoneTextField,
  TextButton,
  TextField
} from '~/components'
import { Label } from '~/components/Label'
import { SignupStep, useSignupStore } from '~/lib/useSignupStore'
import type { loader } from './route'

export function Phone() {
  const otpFetcher = useFetcher()
  const validateFetcher = useFetcher()
  const { csrfToken } = useLoaderData<typeof loader>()
  const [showDialog, setShowDialog] = useState<boolean>(false)

  const [id, country, countries, phone, setPhone, setStep] = useSignupStore(
    (state) => [
      state.id,
      state.country,
      state.countries,
      state.phone,
      state.setPhone,
      state.setStep
    ]
  )

  useEffect(() => {
    if (
      !showDialog &&
      otpFetcher.state == 'loading' &&
      otpFetcher?.data?.success
    ) {
      setShowDialog(true)
    }
  }, [otpFetcher?.data, otpFetcher.state, showDialog])

  useEffect(() => {
    if (validateFetcher.data?.errors?.phone) {
      setShowDialog(false)
    }
  }, [validateFetcher.data])

  useEffect(() => {
    if (validateFetcher.data?.id == id && validateFetcher.data?.phone) {
      setPhone(validateFetcher.data?.phone)
      setStep(SignupStep.PASSWORD)
    }
  }, [
    id,
    setPhone,
    setStep,
    validateFetcher.data?.id,
    validateFetcher.data?.phone
  ])

  return (
    <>
      <otpFetcher.Form
        id='signup-phone-otp'
        action={route('/api/sendOtp')}
        method='post'
        className='hidden'
      />
      <input
        form='signup-phone-otp'
        value={csrfToken}
        name='csrfToken'
        type='hidden'
      />
      <validateFetcher.Form
        id='signup-phone-otp-validation'
        action={route('/signup')}
        method='post'
        className='hidden'
      />
      <input
        form='signup-phone-otp-validation'
        value={csrfToken}
        name='csrfToken'
        type='hidden'
      />

      <input
        form='signup-phone-otp-validation'
        value={id}
        name='id'
        type='hidden'
      />
      <Card>
        <CardContent>
          {phone && <p>Your mobile phone number is verified.</p>}
          {!phone && <p>We require you to verify a mobile phone number.</p>}
        </CardContent>
        {phone && (
          <div className='mt-2 flex space-x-2 rounded-xl bg-nav p-3 text-medium'>
            <Icon>call</Icon>
            <span>{phone}</span>
          </div>
        )}
        {!phone && (
          <PhoneTextField
            id='phone'
            form='signup-phone-otp'
            name='phone'
            defaultCountry={country?.id as string}
            options={countries as PhoneAutocompleteOptions[]}
            label='Mobile number'
            className='mt-2'
            aria-invalid={
              Boolean(
                otpFetcher.data?.errors?.phone ||
                  validateFetcher.data?.errors?.phone
              ) || undefined
            }
            aria-describedby={
              otpFetcher.data?.errors?.phone ||
              validateFetcher.data?.errors?.phone
                ? 'phone-error'
                : undefined
            }
            errorMessage={
              otpFetcher.data?.errors?.phone ||
              validateFetcher.data?.errors?.phone
            }
          />
        )}
      </Card>
      {phone && (
        <Button type='button' onClick={() => setStep(SignupStep.PASSWORD)}>
          <span className='font-medium text-white'>Continue</span>
        </Button>
      )}
      {!phone && (
        <Button form='signup-phone-otp' type='submit'>
          Continue
        </Button>
      )}

      <Dialog open={showDialog} setOpen={setShowDialog}>
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
          <span>{otpFetcher?.data?.phone}</span>
        </div>

        <input
          form='signup-phone-otp-validation'
          value={otpFetcher?.data?.phone}
          name='phone'
          type='hidden'
        />
        <TextField
          id='otp'
          form='signup-phone-otp-validation'
          label='Verification code'
          name='otp'
          type='number'
          className='mt-4'
          aria-invalid={Boolean(validateFetcher.data?.errors?.otp) || undefined}
          aria-describedby={
            validateFetcher.data?.errors?.otp ? 'email-error' : undefined
          }
          required
          errorMessage={validateFetcher.data?.errors?.otp}
        />
        <CardContent className='mt-2 flex w-full justify-end space-x-6'>
          <TextButton type='submit' form='signup-phone-otp'>
            Resend code
          </TextButton>
          <TextButton
            type='submit'
            name='formName'
            value='otp'
            form='signup-phone-otp-validation'
          >
            Verify
          </TextButton>
        </CardContent>
      </Dialog>
    </>
  )
}
