import type { ActionArgs } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { Form, useActionData } from '@remix-run/react'
import { Button, TextArea, TextField } from '~/components'
import {
  grpcClient,
  GrpcError,
  isGrpcError,
  StatusError
} from '../lib/proto.server'

export default function Page() {
  const actionData = useActionData<typeof action>()
  return (
    <main className='w-full overflow-hidden'>
      <section className='relative mx-auto grid w-full grid-cols-4 content-start gap-4 gap-y-2 overflow-x-visible px-8 sm:max-w-lg sm:grid-cols-8 sm:px-0 lg:max-w-3xl lg:grid-cols-12 xl:max-w-4xl'>
        <div className='col-span-full flex justify-center pt-20'>
          <span className='font-display text-2xl font-medium lg:text-4xl'>
            Contact us
          </span>
        </div>
        <div className='col-span-full flex justify-center pt-1'>
          <span className='font-sans text-base font-normal'>
            Get in touch and let us know how we can help.
          </span>
        </div>
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
          className='col-span-full mt-10 flex flex-col sm:col-span-6 sm:col-start-2 lg:col-start-4'
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
          className='col-span-full flex flex-col sm:col-span-6 sm:col-start-2 lg:col-start-4'
          aria-invalid={Boolean(actionData?.errors.lastName) || undefined}
          aria-describedby={
            actionData?.errors.lastName ? 'lastName-error' : undefined
          }
          errorMessage={actionData?.errors.lastName}
        />

        <TextField
          id='email'
          form='contact-form'
          label='Email address*'
          name='email'
          type='text'
          className='col-span-full flex flex-col sm:col-span-6 sm:col-start-2 lg:col-start-4'
          aria-invalid={Boolean(actionData?.errors.email) || undefined}
          aria-describedby={
            actionData?.errors.email ? 'email-error' : undefined
          }
          required
          errorMessage={actionData?.errors.email}
        />

        <TextArea
          id='description'
          form='contact-form'
          label='Details / comments*'
          name='description'
          className='col-span-full flex flex-col sm:col-span-6 sm:col-start-2 lg:col-start-4'
          aria-invalid={Boolean(actionData?.errors.description) || undefined}
          aria-describedby={
            actionData?.errors.description ? 'description-error' : undefined
          }
          required
          errorMessage={actionData?.errors.description}
        />
        <div className='col-span-full mb-48 flex justify-end pt-4 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
          <Button form='contact-form' type='submit'>
            Submit
          </Button>
        </div>
        {/* Both */}
        <div className='absolute right-4 top-0 h-20 w-20 rounded-full bg-lime-300' />
        <div className='absolute right-4 bottom-20 h-20 w-20 rounded-full bg-rose-300 lg:right-16 lg:bottom-0' />
        <div className='absolute left-4 bottom-0 h-20 w-20 rounded-tl-full bg-orange-200 lg:-left-20 lg:bottom-40' />

        <div className='absolute -right-40 top-40 hidden h-20 w-20 rounded-br-full bg-slate-100 lg:block' />
        <div className='absolute -left-20 top-60 hidden h-20 w-20 rounded-br-full bg-rose-500 lg:block' />
        <div className='absolute -left-[calc(100vw)] bottom-20 hidden h-20 w-screen bg-orange-400 lg:block' />
        <div className='absolute -right-[calc(100vw-4rem)] bottom-20 hidden h-20 w-screen rounded-tl-full bg-slate-700 lg:block' />
        <div className='absolute left-0 bottom-0 hidden h-20 w-20 rounded-br-full bg-slate-100 lg:block' />
        <div className='absolute -right-4 bottom-40 hidden h-20 w-20 rounded-bl-full bg-slate-100 lg:block' />
        <div className='absolute -right-24 bottom-60 hidden h-20 w-20 rounded-br-full bg-slate-200 lg:block' />
      </section>
    </main>
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
    } else throw response
  }

  return redirect('/contact/success')
}
