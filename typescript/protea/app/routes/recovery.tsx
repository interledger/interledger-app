import { useActionData, json, Form, redirect, useLoaderData } from 'remix'
import type { ActionFunction, LoaderFunction } from 'remix'
import { Button, Logo, Router, TextField } from '~/components'
import axios, { AxiosError } from 'axios'
import React from 'react'
import { route } from 'routes-gen'
import {
  getCsrfTokenFromFlow,
  kratos,
  handleFlowError,
  requireNoUserSession
} from '~/lib/kratos'

type ActionData = {
  formError?: string
  fieldErrors?: {
    email: string | undefined
  }
  fields?: {
    email: string
    csrf_token: string
  }
}

const badRequest = (data: ActionData) => json(data, { status: 400 })

export const action: ActionFunction = async ({ request }) => {
  const url = new URL(request.url)
  const flowId = url.searchParams.get('flow')
  const form = await request.formData()
  const csrfToken = form.get('csrf_token')
  const email = form.get('email')

  if (typeof csrfToken !== 'string' || typeof email !== 'string') {
    return badRequest({
      formError: `Form not submitted correctly.`
    })
  }

  const fields = { csrf_token: csrfToken, email }

  return axios
    .post(
      `http://kratos-public/self-service/recovery?flow=${flowId}`,
      {
        method: 'link',
        email: email,
        csrf_token: csrfToken
      },
      {
        headers: {
          'Content-type': 'application/json',
          cookie: String(request.headers.get('cookie'))
        }
      }
    )
    .then((res) => json(res.data))
    .catch((err: AxiosError) => {
      // If the previous handler did not catch the error it's most likely a form validation error
      if (err.response?.status === 400) {
        // Yup, it is!
        return badRequest({ fieldErrors: err.response?.data, fields })
      }
      return err
    })
}

export const loader: LoaderFunction = async ({ request }) => {
  await requireNoUserSession(request)
  const url = new URL(request.url)
  const flowId = url.searchParams.get('flow')
  const returnTo = url.searchParams.get('return_to')

  // TODO: get flow from kratos and handle flow errors appropriately.
  // If ?flow=.. was in the URL, we fetch it
  if (flowId) {
    return kratos
      .getSelfServiceRecoveryFlow(
        String(flowId),
        request.headers.get('cookie') ?? undefined
      )
      .then((res) => json(res.data))
      .catch(handleFlowError('recovery'))
  }

  // Otherwise we initialize it
  return kratos
    .initializeSelfServiceRecoveryFlowForBrowsers(
      returnTo ? String(returnTo) : undefined
    )
    .then(async (res) =>
      redirect(`/recovery?flow=${res.data.id}`, {
        headers: res.headers
      })
    )
    .catch(handleFlowError('recovery'))
}

export default function RecoveryPage() {
  const actionData = useActionData<ActionData>()
  const loaderData = useLoaderData()

  return (
    <main className='mx-auto grid min-h-screen w-full grid-cols-4 content-start gap-4 gap-y-2 overflow-y-auto p-4 sm:max-w-lg sm:grid-cols-8 sm:px-0 lg:max-w-3xl lg:grid-cols-12 lg:content-center xl:max-w-4xl'>
      <div className='col-span-full sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <Router to={route('/')}>
          <Logo className='h-8' />
        </Router>
      </div>
      {loaderData.state === 'sent_email' && (
        <>
          <div className='col-span-full pt-4 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
            <h1 className='font-display text-4xl font-medium leading-normal text-strong'>
              Email sent!
            </h1>
          </div>
          <div className='col-span-full pb-8 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
            <p className='text-medium'>
              We’ve sent you an email to change your password. Please click on
              the link in the email to continue.
            </p>
          </div>
        </>
      )}
      {loaderData.state === 'choose_method' && (
        <>
          <div className='col-span-full pt-4 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
            <h1 className='font-display text-4xl font-medium leading-normal text-strong'>
              Recover your account
            </h1>
          </div>
          <div className='col-span-full pb-8 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
            <p className='text-medium'>
              We’ll send you an email to change your password.
            </p>
          </div>
        </>
      )}
      <Form
        action={`/recovery?flow=${loaderData.id}`}
        method='post'
        className='col-span-full flex flex-col items-end space-y-2 sm:col-span-6 sm:col-start-2 lg:col-start-4'
      >
        <TextField
          id='email'
          label='Email'
          name='email'
          defaultValue={actionData?.fields?.email}
          type='email'
          disabled={loaderData.state === 'sent_email'}
          aria-invalid={Boolean(actionData?.fieldErrors?.email) || undefined}
          aria-describedby={
            actionData?.fieldErrors?.email ? 'email-error' : undefined
          }
          required
          errorMessage={actionData?.fieldErrors?.email}
        />

        <input
          defaultValue={getCsrfTokenFromFlow(loaderData)}
          name='csrf_token'
          type='hidden'
        />
        <div className='pt-4'>
          <Button disabled={loaderData.state === 'sent_email'} type='submit'>
            Recover account
          </Button>
        </div>
      </Form>
    </main>
  )
}
