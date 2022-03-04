import {
  GetFundingSourcesQuery,
  GetFundingSourcesQueryVariables,
  GetFundingSourcesDocument
} from '~/generated/types'
import React, { useState } from 'react'
import {
  ActionFunction,
  Form,
  json,
  Link,
  LoaderFunction,
  redirect,
  useActionData,
  useLoaderData
} from 'remix'
import { route } from 'routes-gen'
import { BackIcon, Button, Select, TextField } from '~/components'
import { apolloClient } from '~/lib/apollo.server'
import { requireUserSession } from '~/lib/kratos'

type ActionData = {
  formError?: string
  fieldErrors?: {
    fundingsource: string | undefined
    amount: string | undefined
  }
  fields?: {
    fundingsource: string
    amount: string
  }
}

const badRequest = (data: ActionData) => json(data, { status: 400 })

export const action: ActionFunction = async ({ request }) => {
  const form = await request.formData()
  const fundingsource = form.get('fundingsource')
  const amount = form.get('amount')

  /**
   * TODO:
   * - Form validation
   * - Submission flow
   * - Route to /deposit/preview on success
   */

  if (typeof fundingsource !== 'string' || typeof amount !== 'string') {
    return badRequest({
      formError: `Form not submitted correctly.`
    })
  }
  return null
}

export const loader: LoaderFunction = async ({ request }) => {
  requireUserSession(request)

  const cookie = request.headers.get('cookie')
  const fundingSources = await apolloClient
    .query<GetFundingSourcesQuery, GetFundingSourcesQueryVariables>({
      query: GetFundingSourcesDocument,
      context: {
        headers: {
          cookie: cookie
        }
      }
    })
    .then((val) => val.data.fundingSources)

  if (fundingSources.length === 0) return redirect(route('/bank'))
  return fundingSources
}

export default function DepositPage() {
  const actionData = useActionData<ActionData>()
  // TODO: use loaderData to get the user's fundingSources
  const loaderData = useLoaderData()
  const [selected, setSelected] = useState(loaderData[0])
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
          Deposit
        </div>
      </header>
      {/* Body */}
      <Form
        action='/deposit'
        method='post'
        className='mx-auto grid min-h-[calc(100vh-9rem)] w-full grid-cols-4 content-start gap-4 gap-y-2 overflow-y-auto p-4 pb-24 sm:max-w-lg sm:grid-cols-8 sm:px-0 lg:max-w-3xl lg:grid-cols-12 xl:max-w-4xl'
      >
        {/* <div className='col-span-full flex min-w-full flex-col sm:col-span-6 sm:col-start-2 lg:col-start-4'></div> */}
        <Select
          id='fundingsource'
          value={selected}
          onChange={setSelected}
          options={loaderData}
          label='Deposit source'
          // defaultValue={actionData?.fields?.fundingsource}
          className='col-span-full flex flex-col sm:col-span-6 sm:col-start-2 lg:col-start-4'
          aria-invalid={
            Boolean(actionData?.fieldErrors?.fundingsource) || undefined
          }
          aria-describedby={
            actionData?.fieldErrors?.fundingsource
              ? 'fundingsource-error'
              : undefined
          }
          errorMessage={actionData?.fieldErrors?.fundingsource}
        />
        <input value={String(selected.id)} name='fundingsource' type='hidden' />

        <TextField
          id='amount'
          label='Amount'
          name='amount'
          defaultValue={actionData?.fields?.amount}
          type='number'
          className='col-span-full flex flex-col sm:col-span-6 sm:col-start-2 lg:col-start-4'
          aria-invalid={Boolean(actionData?.fieldErrors?.amount) || undefined}
          aria-describedby={
            actionData?.fieldErrors?.amount ? 'amount-error' : undefined
          }
          required
          errorMessage={actionData?.fieldErrors?.amount}
        />

        <div className='col-span-full flex min-w-full flex-col items-end sm:col-span-6 sm:col-start-2 lg:col-start-4'>
          <Button type='submit'>Preview deposit</Button>
        </div>
      </Form>
    </div>
  )
}
