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
    <div className='mx-auto grid w-full grid-cols-4 content-start gap-4 gap-y-2 overflow-y-auto rounded-2xl bg-page px-4 pb-16 pt-6 sm:max-w-lg sm:grid-cols-8 sm:px-0 lg:max-w-3xl lg:pt-12 xl:max-w-4xl'>
      <div className='col-span-full flex flex-col space-y-4 pb-8 sm:col-span-6 sm:col-start-2 sm:space-y-6'>
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
        {hasVerified && <span>Your mobile phone number is verified.</span>}
        {!hasVerified && (
          <span>We require you to verify a mobile phone number.</span>
        )}
      </div>
      <otpFetcher.Form
        id='signup-phone-otp'
        action='/api/sendOtp'
        method='post'
        className='hidden'
      />

      {hasVerified && (
        <div className='col-span-full flex space-x-2 rounded-xl bg-container p-3 text-medium  sm:col-span-6 sm:col-start-2'>
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
          className='col-span-full flex flex-col sm:col-span-6 sm:col-start-2'
          aria-invalid={Boolean(otpFetcher.data?.errors?.phone) || undefined}
          aria-describedby={
            otpFetcher.data?.errors?.phone ? 'phone-error' : undefined
          }
          errorMessage={otpFetcher.data?.errors?.phone}
        />
      )}

      <div className='col-span-full flex justify-end pt-4 sm:col-span-6 sm:col-start-2'>
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
        <button
          className='flex items-start p-4 text-sm font-medium text-primary focus-visible:outline-none'
          type='submit'
          form='signup-phone-otp'
        >
          Resend code
        </button>
        <div className='flex w-full justify-end space-x-6'>
          <button
            className='flex items-start text-sm font-medium text-medium focus-visible:outline-none'
            type='button'
            onClick={() => setShowDialog(false)}
            form='signup-phone-otp'
          >
            Cancel
          </button>
          <button
            className='flex items-start  text-sm font-medium text-primary focus-visible:outline-none'
            type='submit'
            form='signup-phone-otp-validation'
          >
            OK
          </button>
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
