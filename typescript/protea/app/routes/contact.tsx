import type { ActionArgs } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { Form, useActionData } from '@remix-run/react'
import { Button, TextArea, TextField } from '~/components'
import type { GrpcError } from '~/lib/proto.server'
import {
  grpcClient,
  httpMapping,
  isGrpcError,
  StatusError
} from '~/lib/proto.server'

export default function Page() {
  const actionData = useActionData<typeof action>()
  return (
    <div className='flex w-full flex-col rounded-2xl bg-page p-4 pb-8'>
      <span className='font-display text-2xl font-medium'>Contact us</span>
      <span className='mt-6 text-medium'>
        Get in touch and let us know how we can help.
      </span>
      <Form
        id='contact-form'
        action='/contact'
        method='post'
        className='hidden'
      />
      <TextField
        id='firstName'
        form='contact-form'
        label='First name'
        name='firstName'
        type='text'
        className='mt-6'
        aria-invalid={Boolean(actionData?.errors.firstName) || undefined}
        aria-describedby={
          actionData?.errors.firstName ? 'firstName-error' : undefined
        }
        errorMessage={actionData?.errors.firstName}
      />

      <TextField
        id='lastName'
        form='contact-form'
        label='Last name'
        name='lastName'
        type='text'
        className='mt-1'
        aria-invalid={Boolean(actionData?.errors.lastName) || undefined}
        aria-describedby={
          actionData?.errors.lastName ? 'lastName-error' : undefined
        }
        errorMessage={actionData?.errors.lastName}
      />

      <TextField
        id='email'
        form='contact-form'
        label='Email address'
        name='email'
        type='text'
        className='mt-1'
        aria-invalid={Boolean(actionData?.errors.email) || undefined}
        aria-describedby={actionData?.errors.email ? 'email-error' : undefined}
        required
        errorMessage={actionData?.errors.email}
      />

      <TextArea
        id='description'
        form='contact-form'
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
      <Button className='mt-12' form='contact-form' type='submit'>
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
    firstName: '',
    lastName: '',
    email: '',
    description: ''
  }

  let response = await grpcClient
    .createSupportTicket({
      description: description,
      firstName: firstName,
      lastName: lastName,
      email: email
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
