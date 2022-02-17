import { ActionFunction, LoaderFunction, redirect, useLoaderData } from 'remix'
import { useActionData, json, Form } from 'remix'
import { Button, Logo, Router, TextField } from '~/components'
import axios, { AxiosError } from 'axios'
import React from 'react'
import { route } from 'routes-gen'
import {
  getCsrfTokenFromFlow,
  kratos,
  handleFlowError,
  checkUserSession
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
  const url = new URL(request.url)
  const flowId = url.searchParams.get('flow')
  const returnTo = url.searchParams.get('return_to')

  // Check if user has session already
  const session = await checkUserSession(request)
  if (session != null) return session

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
    <main className='mx-auto flex h-screen max-w-sm flex-col items-start justify-center px-4'>
      <Router to={route('/')} aria-label='Fynbos logo'>
        <Logo className='h-8' />
      </Router>
      {loaderData.state === 'sent_email' && (
        <>
          <h1 className='mt-6 mb-1 font-display text-4xl font-medium leading-normal text-strong'>
            Email sent!
          </h1>
          <p className='mb-10 text-medium'>
            We’ve sent you an email to change your password. Please click on the
            link in the email to continue.
          </p>
        </>
      )}
      {loaderData.state === 'choose_method' && (
        <>
          <h1 className='mt-6 mb-1 font-display text-4xl font-medium leading-normal text-strong'>
            Recover your account
          </h1>
          <p className='mb-10 text-medium'>
            We’ll send you an email to change your password.
          </p>
        </>
      )}
      <Form
        action={`/recovery?flow=${loaderData.id}`}
        method='post'
        className='flex min-w-full flex-col items-end space-y-4'
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

        <Button disabled={loaderData.state === 'sent_email'} type='submit'>
          Recover account
        </Button>
      </Form>
    </main>
  )
}
