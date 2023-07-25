import type { ActionArgs, LoaderArgs, MetaFunction } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import {
  Form,
  useActionData,
  useFetcher,
  useLoaderData
} from '@remix-run/react'
import { useEffect, useState } from 'react'
import { route } from 'routes-gen'
import type { ApplicationProps, PhoneAutocompleteOptions } from '~/components'
import {
  Button,
  ButtonRouter,
  Card,
  CardContent,
  CardHeader,
  Dialog,
  Icon,
  Layouts,
  PhoneTextField,
  TextButton,
  TextField
} from '~/components'
import { Label } from '~/components/Label'
import { Code } from '~/generated/protobuf-ts/google/rpc/code'
import { getSessionWithCSRFToken, validateCSRFToken } from '~/lib/csrf.server'
import { flowType, requireFlow, updateFlow } from '~/lib/flows.server'
import { requireNoUserSession } from '~/lib/kratos.server'
import type { GrpcError } from '~/lib/proto.server'
import {
  StatusError,
  grpcClient,
  httpMapping,
  isGrpcError
} from '~/lib/proto.server'
import { canSignup } from '~/lib/signupCheck.server'
import { commitSession } from '~/session.server'
import styles from '~/styles/flags.css'

export async function loader({ request }: LoaderArgs) {
  const session = await getSessionWithCSRFToken(request)
  await requireNoUserSession(request)
  await canSignup(request)
  const flow = await requireFlow(request, flowType.Signup)
  let countries = await grpcClient
    .getCountries({})
    .then((v) => v)
    .catch(StatusError)
  if (isGrpcError(countries)) {
    throw json({}, httpMapping(countries.code))
  }

  return json(
    {
      flow,
      hasVerified:
        typeof flow.data.phone !== 'undefined' &&
        typeof flow.data.otp !== 'undefined',
      countries: countries.response.countries,
      csrfToken: session.get('csrf-token')
    },
    { headers: { 'Set-Cookie': await commitSession(session) } }
  )
}

export function links() {
  return [{ rel: 'stylesheet', href: styles }]
}

export const handle: ApplicationProps = {
  layout: Layouts.Focus,
  scaffold: {
    header: {
      back: route('/signup/about'),
      title: 'Mobile phone number'
    }
  }
}

export const meta: MetaFunction = () => {
  return {
    title: 'Sign up | Mobile phone number'
  }
}

export default function Page() {
  /**
   * When a form needs to open a dialog to complete the process, the initial form can use a fetcher.Form - because the submission of this form doesn't cause navigation.
   */
  const otpFetcher = useFetcher()
  const actionData = useActionData<typeof action>()
  const { flow, hasVerified, countries, csrfToken } =
    useLoaderData<typeof loader>()
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

  useEffect(() => {
    if (actionData?.errors?.phone) {
      setShowDialog(false)
    }
  }, [actionData])

  return (
    <>
      <otpFetcher.Form
        id='signup-phone-otp'
        action='/api/sendOtp'
        method='post'
        className='hidden'
      />
      <input
        id='csrfToken'
        form='signup-phone-otp'
        value={csrfToken}
        name='csrfToken'
        type='hidden'
      />
      <Card>
        <CardContent>
          {hasVerified && <p>Your mobile phone number is verified.</p>}
          {!hasVerified && (
            <p>We require you to verify a mobile phone number.</p>
          )}
        </CardContent>
        {hasVerified && (
          <div className='mt-2 flex space-x-2 rounded-xl bg-nav p-3 text-medium'>
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
            className='mt-2'
            aria-invalid={
              Boolean(
                otpFetcher.data?.errors?.phone || actionData?.errors?.phone
              ) || undefined
            }
            aria-describedby={
              otpFetcher.data?.errors?.phone || actionData?.errors?.phone
                ? 'phone-error'
                : undefined
            }
            errorMessage={
              otpFetcher.data?.errors?.phone || actionData?.errors?.phone
            }
          />
        )}
      </Card>
      {hasVerified && (
        <ButtonRouter to={route('/signup/password')}>
          <span className='font-medium text-white'>Continue</span>
        </ButtonRouter>
      )}
      {!hasVerified && (
        <Button form='signup-phone-otp' type='submit'>
          Continue
        </Button>
      )}

      <Form
        id='signup-phone-otp-validation'
        action='/signup/phone'
        method='post'
        className='hidden'
      />
      <input
        id='csrfToken'
        form='signup-phone-otp-validation'
        value={csrfToken}
        name='csrfToken'
        type='hidden'
      />
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
          aria-invalid={Boolean(actionData?.errors.otp) || undefined}
          aria-describedby={actionData?.errors.otp ? 'email-error' : undefined}
          required
          errorMessage={actionData?.errors.otp}
        />
        <CardContent className='mt-2 flex w-full justify-end space-x-6'>
          <TextButton type='submit' form='signup-phone-otp'>
            Resend code
          </TextButton>
          <TextButton type='submit' form='signup-phone-otp-validation'>
            Verify
          </TextButton>
        </CardContent>
      </Dialog>
    </>
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
  const csrfToken = form.get('csrfToken') as string
  const err = await validateCSRFToken(request, csrfToken).catch(
    (err: Error) => err
  )
  if (err) {
    return json(
      {
        errors: {
          form: '',
          otp: 'Please try again.',
          phone: 'Please try again.'
        }
      },
      { status: 422, statusText: 'Invalid CSRF token.' }
    )
  }

  const fieldErrors = {
    form: '',
    otp: '',
    phone: ''
  }

  const flow = await requireFlow(request, flowType.Signup)

  let response = await grpcClient
    .setSignupMobileNumber({
      id: flow.data.id,
      mobile: phone,
      otp: otp
    })
    .then((v) => v)
    .catch(StatusError)

  if (isGrpcError(response)) {
    if (response.code == Code.INVALID_ARGUMENT) {
      for (let violation of (response as GrpcError).details[0]
        .fieldViolations) {
        const field = mapper(violation.field as fieldErrorsMap)
        if (field != null) fieldErrors[field] = violation.description
      }
      return json({ errors: { ...fieldErrors } }, { status: 400 })
    } else if (response.code == Code.ALREADY_EXISTS) {
      fieldErrors['phone'] = 'Mobile phone number is already registered.'
      return json({ errors: { ...fieldErrors } }, { status: 409 })
    } else throw json({}, httpMapping(response.code))
  }

  await updateFlow(request, flowType.Signup, {
    phone,
    otp
  })

  return redirect(route('/signup/password'))
}
