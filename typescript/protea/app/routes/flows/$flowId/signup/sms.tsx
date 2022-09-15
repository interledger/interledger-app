import type { ActionArgs, LoaderArgs } from '@remix-run/node'
import { redirect } from '@remix-run/node'
import { json } from '@remix-run/node'
import { Form, useActionData, useLoaderData } from '@remix-run/react'
import { route } from 'routes-gen'
import { Button, Icon, TextField } from '~/components'
import { getCurrentFlow, updateFlow } from '~/lib/flows.server'
import type { GrpcError } from '~/lib/proto.server'
import { httpMapping } from '~/lib/proto.server'
import { grpcClient, StatusError, isGrpcError } from '~/lib/proto.server'

export async function loader({ request, params }: LoaderArgs) {
  const flow = await getCurrentFlow(request, params)

  return json({
    flow
  })
}

export default function Page() {
  const actionData = useActionData<typeof action>()
  const { flow } = useLoaderData<typeof loader>()

  return (
    <>
      <div className='col-span-full flex flex-col space-y-2 pt-4 pb-8 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <span className='font-display text-2xl font-medium'>
          SMS code verification
        </span>
        <span>Enter the verification code in the SMS sent to you.</span>
      </div>
      <Form
        id='signup-code-details'
        action={`/flows/${flow.id}/signup/sms`}
        method='post'
        className='hidden'
      />
      <div className='col-span-full mb-4 flex items-center justify-between rounded-xl bg-container p-3 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <div className='flex items-center space-x-3 text-medium'>
          <Icon>phone</Icon>
          <span>{flow.data.phone}</span>
        </div>
      </div>

      <TextField
        id='code'
        form='signup-code-details'
        label='SMS code'
        name='code'
        defaultValue={flow?.data.code}
        type='tel'
        className='col-span-full flex flex-col sm:col-span-6 sm:col-start-2 lg:col-start-4'
        aria-invalid={Boolean(actionData?.errors?.code) || undefined}
        aria-describedby={actionData?.errors?.code ? 'code-error' : undefined}
        required
        errorMessage={actionData?.errors?.code}
      />

      <div className='col-span-full flex justify-between pt-4 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <Button
          outline
          className='font-display text-sm font-medium text-primary disabled:text-medium'
          name='resend'
          value={flow.data.phone}
          form='signup-code-details'
        >
          {/* TODO: Update on feedback from justin */}
          'Resend code'
        </Button>
        <Button form='signup-code-details' type='submit'>
          Continue
        </Button>
      </div>
    </>
  )
}

// The field names given by the backend for field violations
type fieldErrorsMap = 'Code'

function mapper(field: fieldErrorsMap): 'code' | null {
  switch (field) {
    case 'Code':
      return 'code'
    default:
      return null
  }
}

export async function action({ request, params }: ActionArgs) {
  const flow = await getCurrentFlow(request, params)
  const onboardingId = flow?.data.id
  const phone = flow?.data.phone

  const form = await request.formData()
  const code = form.get('code') as string
  const resend = form.get('resend') as string

  const fieldErrors = {
    code: ''
  }

  if (resend) {
    await grpcClient
      .sendPhoneVerification({
        to: resend,
        onboardingId
      })
      .then((v) => v)
      .catch(StatusError)

    return json({
      errors: {
        ...fieldErrors
      }
    })
  }

  let response = await grpcClient
    .checkPhoneVerificationCode({
      to: phone,
      code,
      onboardingId
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

  const headers = await updateFlow(request, {
    code
  })
  return redirect(
    route('/flows/:flowId/signup/password', {
      flowId: flow?.id as string
    }),
    { headers }
  )
}
