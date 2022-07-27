import type { ActionFunction, LoaderFunction } from '@remix-run/node'
import { json } from '@remix-run/node'
import { Form, useActionData, useLoaderData } from '@remix-run/react'
import { Button, TextField } from '~/components'
import { getCurrentFlow, stepFlow } from '~/lib/flows.server'

type ActionData = {
  formError?: string
  fieldErrors?: {
    to: string | undefined
  }
  fields?: {
    to: string
  }
}

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
      <Form
        id='send-to'
        action={`/flows/${flow.id}/send/to`}
        method='post'
        className='hidden'
      />

      <TextField
        id='to'
        form='send-to'
        label='Send to'
        name='to'
        type='text'
        className='col-span-full flex flex-col sm:col-span-6 sm:col-start-2 lg:col-start-4'
        aria-invalid={Boolean(actionData?.fieldErrors?.to) || undefined}
        aria-describedby={actionData?.fieldErrors?.to ? 'to-error' : undefined}
        required
        errorMessage={actionData?.fieldErrors?.to}
      />

      <div className='col-span-full flex justify-end pt-4 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <Button form='send-to' type='submit'>
          Continue
        </Button>
      </div>
    </>
  )
}

export const action: ActionFunction = async ({ request }) => {
  const form = await request.formData()
  const sendToAddress = form.get('to')
  await stepFlow(request, {
    to: sendToAddress
  })
}
