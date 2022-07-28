import type { ActionArgs, LoaderArgs } from '@remix-run/node'
import { redirect } from '@remix-run/node'
import { json } from '@remix-run/node'
import { Form, useActionData, useLoaderData } from '@remix-run/react'
import { route } from 'routes-gen'
import { Button, TextField } from '~/components'
import { getCurrentFlow, updateFlow } from '~/lib/flows.server'

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

export async function action({ request, params }: ActionArgs) {
  const form = await request.formData()
  const sendToAddress = form.get('to')

  const headers = await updateFlow(request, {
    to: sendToAddress
  })

  const flow = await getCurrentFlow(request, params)
  return redirect(
    route('/flows/:flowId/send/amount', {
      flowId: flow?.id as string
    }),
    { headers }
  )
}
