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
          Verify your phone number
        </span>
        <span>
          We require a verified phone number so that we can contact you if we
          need to.
        </span>
      </div>
      <Form
        id='signup-phone-details'
        action={`/flows/${flow.id}/signup/phone`}
        method='post'
        className='hidden'
      />

      <TextField
        id='phone'
        form='signup-phone-details'
        label='Phone number'
        name='phone'
        defaultValue={actionData?.fields?.phone || flow?.data.phone}
        type='tel'
        className='col-span-full flex flex-col sm:col-span-6 sm:col-start-2 lg:col-start-4'
        aria-invalid={Boolean(actionData?.fieldErrors?.phone) || undefined}
        aria-describedby={
          actionData?.fieldErrors?.phone ? 'phone-error' : undefined
        }
        required
        errorMessage={actionData?.fieldErrors?.phone}
      />

      <div className='col-span-full flex justify-end pt-4 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <Button form='signup-phone-details' type='submit'>
          Send SMS
        </Button>
      </div>
    </>
  )
}

type ActionData = {
  formError?: string
  fieldErrors?: {
    phone: string | undefined
  }
  fields?: {
    phone: string
  }
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
        phone: undefined
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
  // TODO Possbily try format phone to e.164 here and return if failed. libphonenumber-js
  const form = await request.formData()
  const phone = form.get('phone') as string

  const flow = await getCurrentFlow(request, params)
  const onboardingId = flow?.data.id

  let call = await grpcClient
    .sendPhoneVerification({
      to: phone,
      onboardingId
    })
    .then((v) => v)
    .catch(StatusError)

  const actionData = parseError(call, {
    phone
  })

  if (actionData != null) return actionData
  await stepFlow(request, {
    phone
  })
}
