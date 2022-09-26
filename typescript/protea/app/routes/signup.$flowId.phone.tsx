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
import { Button, Icon, PhoneAutoComplete, TextField } from '~/components'
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
import { Dialog } from '~/components/Dialog'

export async function loader({ request, params }: LoaderArgs) {
  await requireFlow(request, flowType.Signup, params)
  const flow = await getCurrentFlow(request, flowType.Signup)
  let response = await grpcClient
    .getCountries({})
    .then((v) => v)
    .catch(StatusError)
  if (isGrpcError(response)) {
    throw json({}, httpMapping(response.code))
  }

  return json({
    flow,
    countries: response.response.countries
  })
}

export default function Page() {
  /**
   * When a form needs to open a dialog to complete the process, the initial form can use a fetcher.Form - because the submission of this form doesn't cause navigation.
   */
  const otpFetcher = useFetcher()
  const actionData = useActionData<typeof action>()
  const { flow, countries } = useLoaderData<typeof loader>()

  const [showDialog, setShowDialog] = useState<boolean>(false)

  useEffect(() => {
    if (
      !showDialog &&
      otpFetcher.state == 'loading' &&
      otpFetcher?.data?.success
    ) {
      setShowDialog(true)
    }
  }, [otpFetcher?.data, showDialog])

  return (
    <div className='mx-auto grid w-full grid-cols-4 content-start gap-4 gap-y-2 overflow-y-auto rounded-2xl bg-page px-4 pb-16 pt-6 sm:max-w-lg sm:grid-cols-8 sm:px-0 lg:max-w-3xl lg:pt-12 xl:max-w-4xl'>
      <div className='col-span-full mb-4 flex flex-col space-y-2 sm:col-span-6 sm:col-start-2'>
        <span className='font-display text-2xl font-medium'>Phone number</span>
        <span>We require you to verify a mobile phone number.</span>
      </div>
      <otpFetcher.Form
        id='signup-phone-otp'
        action='/api/sendOtp'
        method='post'
        className='hidden'
      />

      <PhoneAutoComplete
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

      <div className='col-span-full flex justify-end pt-4 sm:col-span-6 sm:col-start-2'>
        <Button form='signup-phone-otp' type='submit'>
          Continue
        </Button>
      </div>
      <Dialog open={showDialog} setOpen={setShowDialog}>
        <Form
          id='signup-phone-otp-validation'
          action='/api/sendOtp'
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
      </Dialog>
    </div>
  )
}

// The field names given by the backend for field violations
type fieldErrorsMap = 'To'

function mapper(field: fieldErrorsMap): 'phone' | null {
  switch (field) {
    case 'To':
      return 'phone'
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
