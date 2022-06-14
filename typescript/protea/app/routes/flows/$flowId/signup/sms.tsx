import React from 'react'
import type { ActionFunction, LoaderFunction } from '@remix-run/node'
import { json } from '@remix-run/node'
import { Form, useActionData, useLoaderData } from '@remix-run/react'
import { Button, TextField } from '~/components'
import { getCurrentFlow, stepFlow } from '~/lib/flows.server'
import type { GrpcError } from '~/lib/proto.server'
import { grpcClient, StatusError, isGrpcError } from '~/lib/proto.server'

export const loader: LoaderFunction = async ({ request, params }) => {
  const flow = await getCurrentFlow(request, params)

  return json({
    flow
  })
}

export default function Page() {
  const actionData = useActionData<ActionData>()
  const { flow } = useLoaderData()

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

      <TextField
        id='code'
        form='signup-code-details'
        label='SMS code'
        name='code'
        defaultValue={actionData?.fields?.code || flow?.data.code}
        type='tel'
        className='col-span-full flex flex-col sm:col-span-6 sm:col-start-2 lg:col-start-4'
        aria-invalid={Boolean(actionData?.fieldErrors?.code) || undefined}
        aria-describedby={
          actionData?.fieldErrors?.code ? 'code-error' : undefined
        }
        required
        errorMessage={actionData?.fieldErrors?.code}
      />

      <div className='col-span-full flex justify-end pt-4 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <Button form='signup-code-details' type='submit'>
          Send SMS
        </Button>
      </div>
    </>
  )
}

type ActionData = {
  formError?: string
  fieldErrors?: {
    code: string | undefined
  }
  fields?: {
    code: string
  }
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

/**
 * parseError handles potention errors from grpc client calls.
 * @param response The response from a grpc call
 * @param fields Any data passed to the grpc call
 * @returns ActionData response for field validation errors, throws other errors or null if not an error
 */
function parseError(response: any, fields: any): Response | null {
  if (isGrpcError(response)) {
    if (response.code == 3) {
      let fieldErrors: ActionData['fieldErrors'] = {
        code: undefined
      }
      for (let violation of (response as GrpcError).details[0]
        .fieldViolations) {
        const field = mapper(violation.field as fieldErrorsMap)
        if (field != null) fieldErrors[field] = violation.description
      }
      return json({
        fields,
        fieldErrors
      })
    } else throw response
  }
  return null
}

export const action: ActionFunction = async ({ request, params }) => {
  const form = await request.formData()
  const code = form.get('code') as string

  const flow = await getCurrentFlow(request, params)
  const onboardingId = flow?.data.id
  const phone = flow?.data.phone

  let call = await grpcClient
    .checkPhoneVerificationCode({
      to: phone,
      code,
      onboardingId
    })
    .then((v) => v)
    .catch(StatusError)

  const actionData = parseError(call, {
    code
  })

  if (actionData != null) return actionData
  await stepFlow(request, {
    code
  })
}
