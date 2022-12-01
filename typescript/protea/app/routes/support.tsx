import type { ActionArgs, LoaderArgs } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { Form, useActionData, useLoaderData } from '@remix-run/react'
import { Button, Layouts, TextArea } from '~/components'
import type { GrpcError } from '~/lib/proto.server'
import {
  grpcClient,
  httpMapping,
  isGrpcError,
  StatusError
} from '~/lib/proto.server'
import { requireUserSession } from '~/lib/kratos.server'

export async function loader({ request }: LoaderArgs) {
  const session = await requireUserSession(request)
  return json({ traits: session.identity.traits })
}

export const handle = {
  layout: Layouts.FocusLayout
}

export default function Page() {
  const { traits } = useLoaderData<typeof loader>()
  const actionData = useActionData<typeof action>()
  return (
    <div className='flex w-full flex-col rounded-2xl bg-page p-4 pb-8'>
      <span className='font-display text-2xl font-medium'>Contact us</span>
      <span className='mt-6 text-medium'>
        Get in touch and let us know how we can help.
      </span>
      <Form
        id='support-form'
        action='/contact'
        method='post'
        className='hidden'
      />
      <input
        defaultValue={traits.firstName}
        name='firstName'
        form='support-form'
        type='hidden'
      />
      <input
        defaultValue={traits.lastName}
        name='lastName'
        form='support-form'
        type='hidden'
      />
      <input
        defaultValue={traits.email}
        name='email'
        form='support-form'
        type='hidden'
      />

      <TextArea
        id='description'
        form='support-form'
        label='Details / comments'
        name='description'
        className='mt-1'
        aria-invalid={Boolean(actionData?.errors.description) || undefined}
        aria-describedby={
          actionData?.errors.description ? 'description-error' : undefined
        }
        required
        errorMessage={actionData?.errors.description}
      />
      <Button className='mt-12' form='support-form' type='submit'>
        Submit
      </Button>
    </div>
  )
}

// The field names given by the backend for field violations
type fieldErrorsMap = 'FirstName' | 'Email' | 'LastName' | 'Description'

function mapper(
  field: fieldErrorsMap
): 'firstName' | 'email' | 'lastName' | 'description' | null {
  switch (field) {
    case 'Email':
      return 'email'
    case 'FirstName':
      return 'firstName'
    case 'LastName':
      return 'lastName'
    case 'Description':
      return 'description'
    default:
      return null
  }
}

export async function action({ request }: ActionArgs) {
  const form = await request.formData()
  const firstName = form.get('firstName') as string
  const lastName = form.get('lastName') as string
  const email = form.get('email') as string
  const description = form.get('description') as string

  const fieldErrors = {
    form: '',
    firstName: '',
    lastName: '',
    email: '',
    description: ''
  }

  let response = await grpcClient
    .createSupportTicket({
      description,
      firstName,
      lastName,
      email
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

  return redirect('/contact/success')
}
