import { useState } from 'react'
import type { PhoneAutocompleteOptions } from '~/components'
import { Button, Card, CardContent, Icon, PhoneTextField } from '~/components'
import { SignupStep, useSignupStore } from '~/lib/useSignupStore'

export function Phone() {
  // const otpFetcher = useFetcher<typeof sendOtpAction>()
  const [phoneNumber, setPhoneNumber] = useState<string>('')
  const [country, countries, phone, setPhone, setStep] = useSignupStore(
    (state) => [
      state.country,
      state.countries,
      state.phone,
      state.setPhone,
      state.setStep
    ]
  )

  return (
    <>
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
            onInput={(event) => {
              setPhoneNumber(event.currentTarget.value)
            }}
            required
          />
        )}
      </Card>
      {phone && (
        <Button type='button' onClick={() => setStep(SignupStep.PASSWORD)}>
          <span className='font-medium text-white'>Continue</span>
        </Button>
      )}
      {!phone && (
        <Button
          onClick={() => {
            setPhone(phoneNumber)
            setStep(SignupStep.PASSWORD)
          }}
        >
          Continue
        </Button>
      )}

      {/* <Dialog open={showDialog} setOpen={setShowDialog}>
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
      </Dialog> */}
    </>
  )
}
