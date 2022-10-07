import { useEffect, useState } from 'react'
import type { ActionArgs, LoaderArgs } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import {
  Form,
  useActionData,
  useFetcher,
  useLoaderData
} from '@remix-run/react'
import type { PhoneAutocompleteOptions } from '~/components'
import {
  Button,
  Dialog,
  Icon,
  PhoneTextField,
  Router,
  Shape,
  TextButton,
  TextField
} from '~/components'
import {
  flowType,
  getCurrentFlow,
  requireFlow,
  updateFlow
} from '~/lib/flows.server'
import type { GrpcError } from '~/lib/proto.server'
import {
  grpcClient,
  httpMapping,
  isGrpcError,
  StatusError
} from '~/lib/proto.server'
import { route } from 'routes-gen'
import styles from '~/styles/flags.css'

export async function loader({ request, params }: LoaderArgs) {
  await requireFlow(request, flowType.Signup, params)
  const flow = await getCurrentFlow(request, flowType.Signup)
  let countries = await grpcClient
    .getCountries({})
    .then((v) => v)
    .catch(StatusError)
  if (isGrpcError(countries)) {
    throw json({}, httpMapping(countries.code))
  }

  return json({
    flow,
    hasVerified:
      typeof flow.data.phone !== 'undefined' &&
      typeof flow.data.otp !== 'undefined',
    countries: countries.response.countries
  })
}

export function links() {
  return [{ rel: 'stylesheet', href: styles }]
}

export default function Page() {
  /**
   * When a form needs to open a dialog to complete the process, the initial form can use a fetcher.Form - because the submission of this form doesn't cause navigation.
   */
  const otpFetcher = useFetcher()
  const actionData = useActionData<typeof action>()
  const { flow, hasVerified, countries } = useLoaderData<typeof loader>()
  const [showDialog, setShowDialog] = useState<boolean>(false)

  useEffect(() => {
    if (
      !showDialog &&
      otpFetcher.state == 'loading' &&
      otpFetcher?.data?.success
    ) {
      setShowDialog(true)
    }
  }, [otpFetcher?.data, otpFetcher.state, showDialog])

  return (
    <div className='flex w-full flex-col rounded-2xl bg-page p-4 pb-8'>
      <div className='flex flex-col space-y-6'>
        <div className='flex justify-between'>
          <span className='font-display text-2xl font-medium'>
            Phone number
          </span>
          <div className='hidden sm:flex'>
            <Shape
              width={'w-8'}
              radius={'rounded-bl-full'}
              color={'bg-slate-300'}
            />
            <Shape
              width={'w-8'}
              radius={'rounded-full'}
              color={'bg-lime-500'}
            />
          </div>
        </div>
        {hasVerified && <p>Your mobile phone number is verified.</p>}
        {!hasVerified && <p>We require you to verify a mobile phone number.</p>}
      </div>
      <otpFetcher.Form
        id='signup-phone-otp'
        action='/api/sendOtp'
        method='post'
        className='hidden'
      />

      {hasVerified && (
        <div className='mt-6 flex space-x-2 rounded-xl bg-container p-3 text-medium'>
          <Icon>call</Icon>
          <span>{flow?.data?.phone}</span>
        </div>
      )}

      {!hasVerified && (
        <PhoneTextField
          id='phone'
          form='signup-phone-otp'
          name='phone'
          defaultCountry={flow?.data.country}
          options={countries as PhoneAutocompleteOptions[]}
          label='Mobile number'
          className='mt-6'
          aria-invalid={Boolean(otpFetcher.data?.errors?.phone) || undefined}
          aria-describedby={
            otpFetcher.data?.errors?.phone ? 'phone-error' : undefined
          }
          errorMessage={otpFetcher.data?.errors?.phone}
        />
      )}

      <div className='mt-12'>
        {hasVerified && (
          <Router
            to={route('/signup/:flowId/password', { flowId: flow.id })}
            className='flex h-[50px] w-full items-center justify-center rounded-full bg-primary px-10'
          >
            <span className='font-display font-medium text-white'>
              Continue
            </span>
          </Router>
        )}
        {!hasVerified && (
          <Button form='signup-phone-otp' type='submit'>
            Continue
          </Button>
        )}
      </div>
      <Dialog open={showDialog} setOpen={setShowDialog}>
        <Form
          id='signup-phone-otp-validation'
          action={`/signup/${flow.id}/phone`}
          method='post'
          className='hidden'
        />
        <h1 className='font-display text-2xl'>Two-step verification</h1>
        <span className='text-medium'>
          Enter the six digit verification code sent to your mobile number.
        </span>
        <div className='flex space-x-2 rounded-xl bg-container p-3 text-medium'>
          <Icon>call</Icon>
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
          className='col-span-full flex flex-col pt-4 sm:col-span-6 sm:col-start-2'
          aria-invalid={Boolean(actionData?.errors.otp) || undefined}
          aria-describedby={actionData?.errors.otp ? 'email-error' : undefined}
          required
          errorMessage={actionData?.errors.otp}
        />

        <div className='flex w-full justify-end space-x-6'>
          <TextButton type='submit' form='signup-phone-otp'>
            Resend code
          </TextButton>
          <TextButton type='submit' form='signup-phone-otp-validation'>
            Verify
          </TextButton>
        </div>
      </Dialog>
    </div>
  )
}

// The field names given by the backend for field violations
type fieldErrorsMap = 'MobileNumber' | 'OTP'

function mapper(field: fieldErrorsMap): 'phone' | 'otp' | null {
  switch (field) {
    case 'MobileNumber':
      return 'phone'
    case 'OTP':
      return 'otp'
    default:
      return null
  }
}

export async function action({ request }: ActionArgs) {
  const form = await request.formData()
  const otp = form.get('otp') as string
  const phone = form.get('phone') as string

  const fieldErrors = {
    otp: '',
    phone: ''
  }

  const flow = await getCurrentFlow(request, flowType.Signup)

  let response = await grpcClient
    .setSignupMobileNumber({
      id: flow.id,
      mobile: phone,
      otp: otp
    })
    .then((v) => v)
    .catch(StatusError)

  if (isGrpcError(response)) {
    if (response.code == 3) {
      for (let violation of (response as GrpcError).details[0]
        .fieldViolations) {
        const field = mapper(violation.field as fieldErrorsMap)
        if (field != null) fieldErrors[field] = violation.description
      }
      return json({ errors: { ...fieldErrors } }, { status: 400 })
    } else throw json({}, httpMapping(response.code))
  }

  const headers = await updateFlow(request, flowType.Signup, {
    phone,
    otp
  })

  return redirect(
    route('/signup/:flowId/password', {
      flowId: flow?.id as string
    }),
    { headers }
  )
}
