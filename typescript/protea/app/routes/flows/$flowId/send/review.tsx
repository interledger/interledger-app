import type { ActionArgs, LoaderArgs } from '@remix-run/node'
import { redirect } from '@remix-run/node'
import { json } from '@remix-run/node'
import { Form, useActionData, useLoaderData } from '@remix-run/react'
import { Button, TextField } from '~/components'
import { updateFlow, getCurrentFlow } from '~/lib/flows.server'
import { route } from 'routes-gen'
import { requireUserSession } from '~/lib/kratos.server'
import {
  grpcClient,
  GrpcError,
  isGrpcError,
  StatusError
} from '~/lib/proto.server'

export async function loader({ request, params }: LoaderArgs) {
  const flow = await getCurrentFlow(request, params)

  await requireUserSession(request)

  const resp = await grpcClient
    .getIdentity(
      {},
      {
        meta: {
          cookies: request.headers.get('cookie') || ''
        }
      }
    )
    .then((v) => v)
    .catch(StatusError)
  if (isGrpcError(resp)) {
    throw resp
  }

  const identity = resp.response

  const otpResp = await grpcClient
    .sendOTP(
      {},
      {
        meta: {
          cookies: request.headers.get('cookie') || ''
        }
      }
    )
    .then((v) => v)
    .catch(StatusError)
  if (isGrpcError(otpResp)) {
    throw otpResp
  }

  return json({
    flow,
    identity
  })
}

export default function Page() {
  const actionData = useActionData<typeof action>()
  const { flow, identity } = useLoaderData<typeof loader>()
  const { linkedAccountMask, displayAmount, displayFee, displayTotal } =
    flow?.data
  return (
    <>
      <Form
        id='send-review'
        action={`/flows/${flow.id}/send/review`}
        method='post'
        className='hidden'
      />

      <div className='col-span-full flex justify-between pb-4 text-medium sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <span className='text-sm font-medium'>Payment method</span>
        <span>{linkedAccountMask}</span>
      </div>
      <div className='text medium col-span-full flex justify-between sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <span className='font-display text-sm font-medium'>Amount</span>
        <span className='text-sm'>{displayAmount || '$ 0.00'}</span>
      </div>
      <div className='text medium col-span-full flex justify-between sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <span className='font-display text-sm font-medium'>Fees</span>
        <span className='text-sm'>{displayFee || '$ 0.00'}</span>
      </div>
      <div className='col-span-full flex items-end justify-between py-3 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <span className='font-display text-2xl font-medium'>Total</span>
        <span className='text-4xl font-medium'>{displayTotal || '$ 0.00'}</span>
      </div>
      <div className='mx-auto grid min-h-[calc(100vh-9rem)] w-full grid-cols-4 content-start gap-4 gap-y-2 overflow-y-auto p-4 pb-24 sm:max-w-lg sm:grid-cols-8 sm:px-0 lg:max-w-3xl lg:grid-cols-12 xl:max-w-4xl'>
        <div className='col-span-full flex flex-col space-y-2 pt-4 pb-8 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
          <span className='font-display text-4xl font-medium'>
            Enter the six digit verification code sent to your mobile device. (
            {identity.mobileNumber})
          </span>
        </div>
        <TextField
          id='otp'
          form='send-review'
          label='Verification Code'
          name='otp'
          type='text'
          className='col-span-full flex flex-col sm:col-span-6 sm:col-start-2 lg:col-start-4'
          aria-invalid={Boolean(actionData?.errors.otp) || undefined}
          aria-describedby={actionData?.errors.otp ? 'email-error' : undefined}
          required
          errorMessage={actionData?.errors.otp}
        />
      </div>

      <div className='col-span-full flex justify-end pt-4 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <Button form='send-review' type='submit'>
          Confirm
        </Button>
      </div>
    </>
  )
}

// The field names given by the backend for field violations
type fieldErrorsMap = 'OTP'

function mapper(field: fieldErrorsMap): 'otp' | null {
  switch (field) {
    case 'OTP':
      return 'otp'
    default:
      return null
  }
}

export async function action({ request, params }: ActionArgs) {
  const flow = await getCurrentFlow(request, params)
  const form = await request.formData()
  const { to, total } = flow?.data
  const cookie = request.headers.get('cookie')

  const fieldErrors = {
    otp: ''
  }

  let response = await grpcClient
    .initiateOutgoingPayment(
      {
        amount: total.toFixed(2).replace('.', ''),
        otp: form.get('otp') as string,
        to: to
      },
      {
        meta: {
          cookies: cookie || ''
        }
      }
    )
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
    } else throw response
  }

  const headers = await updateFlow(request, null, true)

  return redirect(
    route('/confirmation/:flowId/send', {
      flowId: flow?.id as string
    }),
    { headers }
  )
}
