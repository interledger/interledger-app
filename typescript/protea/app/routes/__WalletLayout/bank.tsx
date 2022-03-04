import {
  LinkUsdBankAccountMutationVariables,
  LinkUsdBankAccountMutation,
  LinkUsdBankAccountDocument
} from '~/generated/types'
import React, { useState } from 'react'
import {
  ActionFunction,
  Form,
  json,
  Link,
  LoaderFunction,
  redirect,
  useActionData
} from 'remix'
import { route } from 'routes-gen'
import { BackIcon, Button, Select, TextField } from '~/components'
import { apolloClient } from '~/lib/apollo.server'
import { requireUserSession } from '~/lib/kratos'

type ActionData = {
  formError?: string
  fieldErrors?: {
    accountNumber: string | undefined
    institution: string | undefined
    name: string | undefined
    routingNumber: string | undefined
    type: string | undefined
  }
  fields?: {
    accountNumber: string
    institution: string
    name: string
    routingNumber: string
    type: string
  }
}

const badRequest = (data: ActionData) => json(data, { status: 400 })

export const action: ActionFunction = async ({ request }) => {
  const cookie = request.headers.get('cookie')
  const form = await request.formData()
  const accountNumber = form.get('accountNumber')
  const institution = form.get('institution')
  const name = form.get('name')
  const routingNumber = form.get('routingNumber')
  const type = form.get('type')

  console.table({
    accountNumber: accountNumber,
    institution: institution,
    name: name,
    routingNumber: routingNumber,
    type: type
  })

  // TODO: proper validation of input based on providers requirements
  if (
    typeof accountNumber !== 'string' ||
    typeof institution !== 'string' ||
    typeof name !== 'string' ||
    typeof routingNumber !== 'string' ||
    typeof type !== 'string'
  ) {
    return badRequest({
      formError: `Form not submitted correctly.`
    })
  }

  const linkUsdBankAccountMutationVariables = {
    input: {
      accountNumber: accountNumber,
      institution: institution,
      name: name,
      routingNumber: routingNumber,
      type: type
    }
  }

  const res = await apolloClient.mutate<
    LinkUsdBankAccountMutation,
    LinkUsdBankAccountMutationVariables
  >({
    mutation: LinkUsdBankAccountDocument,
    variables: linkUsdBankAccountMutationVariables,
    context: {
      headers: {
        cookie: cookie
      }
    }
  })
  if (res.data?.linkUsdBankAccount.success) return redirect(route('/deposit'))
  return null
}

export const loader: LoaderFunction = async ({ request }) => {
  return requireUserSession(request)
}

const people = [
  { id: '1', name: 'Checking' },
  { id: '2', name: 'Savings' }
]

export default function DepositPage() {
  const actionData = useActionData<ActionData>()
  // TODO: use loaderData to get the user's fundingSources
  // const loaderData = useLoaderData()
  const [selected, setSelected] = useState(people[0])
  return (
    <div className='w-full'>
      {/* Header */}
      <header className='sticky top-0 mx-auto flex h-16 w-full select-none items-center justify-start bg-white p-4 text-medium sm:max-w-lg lg:max-w-3xl xl:max-w-4xl'>
        <Link to={route('/home')}>
          <div className='-ml-3 p-3 text-medium'>
            <BackIcon />
          </div>
        </Link>
        <div className='flex items-center justify-start font-display text-2xl font-medium'>
          Bank details
        </div>
      </header>
      {/* Body */}
      <Form
        action='/bank'
        method='post'
        className='mx-auto grid min-h-[calc(100vh-9rem)] w-full grid-cols-4 content-start gap-4 gap-y-2 overflow-y-auto p-4 pb-24 sm:max-w-lg sm:grid-cols-8 sm:px-0 lg:max-w-3xl lg:grid-cols-12 xl:max-w-4xl'
      >
        <div className='col-span-full flex min-w-full flex-col sm:col-span-6 sm:col-start-2 lg:col-start-4'></div>
        <Select
          id='type'
          value={selected}
          onChange={setSelected}
          options={people}
          label='Account type'
          // defaultValue={actionData?.fields?.type}
          className='col-span-full flex flex-col sm:col-span-6 sm:col-start-2 lg:col-start-4'
          aria-invalid={Boolean(actionData?.fieldErrors?.type) || undefined}
          aria-describedby={
            actionData?.fieldErrors?.type ? 'type-error' : undefined
          }
          errorMessage={actionData?.fieldErrors?.type}
        />
        <input value={String(selected.name)} name='type' type='hidden' />

        <TextField
          id='institution'
          label='Institution'
          name='institution'
          defaultValue={actionData?.fields?.institution}
          type='text'
          className='col-span-full flex flex-col sm:col-span-6 sm:col-start-2 lg:col-start-4'
          aria-invalid={
            Boolean(actionData?.fieldErrors?.institution) || undefined
          }
          aria-describedby={
            actionData?.fieldErrors?.institution
              ? 'institution-error'
              : undefined
          }
          required
          errorMessage={actionData?.fieldErrors?.institution}
        />

        <TextField
          id='accountNumber'
          label='Account Number'
          name='accountNumber'
          defaultValue={actionData?.fields?.accountNumber}
          type='text'
          className='col-span-full flex flex-col sm:col-span-6 sm:col-start-2 lg:col-start-4'
          aria-invalid={
            Boolean(actionData?.fieldErrors?.accountNumber) || undefined
          }
          aria-describedby={
            actionData?.fieldErrors?.accountNumber
              ? 'accountNumber-error'
              : undefined
          }
          required
          errorMessage={actionData?.fieldErrors?.accountNumber}
        />

        <TextField
          id='routingNumber'
          label='Routing Number'
          name='routingNumber'
          defaultValue={actionData?.fields?.routingNumber}
          type='text'
          className='col-span-full flex flex-col sm:col-span-6 sm:col-start-2 lg:col-start-4'
          aria-invalid={
            Boolean(actionData?.fieldErrors?.routingNumber) || undefined
          }
          aria-describedby={
            actionData?.fieldErrors?.routingNumber
              ? 'routingNumber-error'
              : undefined
          }
          required
          errorMessage={actionData?.fieldErrors?.routingNumber}
        />

        <TextField
          id='name'
          label='Name'
          name='name'
          defaultValue={actionData?.fields?.name}
          type='text'
          className='col-span-full flex flex-col sm:col-span-6 sm:col-start-2 lg:col-start-4'
          aria-invalid={Boolean(actionData?.fieldErrors?.name) || undefined}
          aria-describedby={
            actionData?.fieldErrors?.name ? 'name-error' : undefined
          }
          required
          errorMessage={actionData?.fieldErrors?.name}
        />

        <div className='col-span-full flex min-w-full flex-col items-end sm:col-span-6 sm:col-start-2 lg:col-start-4'>
          <Button type='submit'>Save account</Button>
        </div>
      </Form>
    </div>
  )
}
