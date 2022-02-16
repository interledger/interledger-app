import { ActionFunction, LoaderFunction, redirect, useLoaderData } from 'remix'
import { useActionData, json, Form } from 'remix'
import { Button, Logo, Router, TextField } from '~/components'
import axios, { AxiosError, AxiosResponse } from 'axios'
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
    password: string | undefined
  }
  fields?: {
    email: string
    password: string
    csrf_token: string
  }
}

const setAllCookiesHeaders = (response: AxiosResponse): Headers => {
  const headers = new Headers()
  const setCookieHeaders = response.headers['set-cookie']
  if (typeof setCookieHeaders === 'undefined') return headers
  setCookieHeaders.forEach((val) => {
    headers.append('set-cookie', val)
  })
  return headers
}

const badRequest = (data: ActionData) => json(data, { status: 400 })

export const action: ActionFunction = async ({ request }) => {
  const url = new URL(request.url)
  const flowId = url.searchParams.get('flow')
  const form = await request.formData()
  const csrfToken = form.get('csrf_token')
  const email = form.get('email')
  const password = form.get('password')

  if (
    typeof csrfToken !== 'string' ||
    typeof email !== 'string' ||
    typeof password !== 'string'
  ) {
    return badRequest({
      formError: `Form not submitted correctly.`
    })
  }

  const fields = { csrf_token: csrfToken, email, password }

  return await axios
    .post(
      `http://kratos-public/self-service/login?flow=${flowId}`,
      {
        method: 'password',
        password_identifier: email,
        password: password,
        csrf_token: csrfToken
      },
      {
        headers: {
          'Content-type': 'application/json',
          cookie: String(request.headers.get('cookie'))
        }
      }
    )
    .then((res) =>
      redirect(route('/home'), {
        headers: setAllCookiesHeaders(res)
      })
    )
    .catch(handleFlowError('login'))
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
  // Refresh means we want to refresh the session. This is needed, for example, when we want to update the password
  // of a user.
  const refresh = url.searchParams.get('refresh')
  // AAL = Authorization Assurance Level. This implies that we want to upgrade the AAL, meaning that we want
  // to perform two-factor authentication/verification.
  const aal = url.searchParams.get('aal')
  const returnTo = url.searchParams.get('return_to')

  // Check if user has session already
  const session = await checkUserSession(request)
  if (session != null) return session

  // TODO: get flow from kratos and handle flow errors appropriately.
  // If ?flow=.. was in the URL, we fetch it
  if (flowId) {
    return kratos
      .getSelfServiceLoginFlow(
        String(flowId),
        request.headers.get('cookie') ?? undefined
      )
      .then((res) => json(res.data))
      .catch(handleFlowError('login'))
  }

  // Otherwise we initialize it
  return kratos
    .initializeSelfServiceLoginFlowForBrowsers(
      Boolean(refresh),
      aal ? String(aal) : undefined,
      returnTo ? String(returnTo) : undefined
    )
    .then(async (res) =>
      redirect(`/login?flow=${res.data.id}`, {
        headers: res.headers
      })
    )
    .catch(handleFlowError('login'))
}

export default function LoginPage() {
  const actionData = useActionData<ActionData>()
  const loaderData = useLoaderData()

  return (
    <main className='mx-auto flex h-screen max-w-sm flex-col items-start justify-center px-4'>
      <Router to={route('/')} aria-label='Fynbos logo'>
        <Logo className='h-8' />
      </Router>
      <h1 className='mt-6 mb-1 font-display text-4xl font-medium leading-normal text-strong'>
        Sign in to your account
      </h1>
      <p className='mb-10 text-medium'>
        Or{' '}
        <Router to={route('/signup')}>
          <span className='text-primary'>create a new account.</span>
        </Router>
      </p>
      <Form
        action={`/login?flow=${loaderData.id}`}
        method='post'
        className='flex min-w-full flex-col space-y-4'
      >
        <TextField
          id='email'
          label='Email'
          name='email'
          defaultValue={actionData?.fields?.email}
          type='email'
          aria-invalid={Boolean(actionData?.fieldErrors?.email) || undefined}
          aria-describedby={
            actionData?.fieldErrors?.email ? 'email-error' : undefined
          }
          required
          errorMessage={actionData?.fieldErrors?.email}
        />
        <TextField
          id='password'
          label='Password'
          name='password'
          defaultValue={actionData?.fields?.password}
          type='password'
          aria-invalid={Boolean(actionData?.fieldErrors?.password) || undefined}
          aria-describedby={
            actionData?.fieldErrors?.password ? 'password-error' : undefined
          }
          required
          minLength={8}
          errorMessage={actionData?.fieldErrors?.password}
        />

        <input
          defaultValue={getCsrfTokenFromFlow(loaderData)}
          name='csrf_token'
          type='hidden'
        />

        <div className='flex items-center justify-between'>
          <Router to={route('/recovery')} aria-label='Forgot password?'>
            <span className='text-primary'>Forgot password?</span>
          </Router>
          <Button type='submit'>Login</Button>
        </div>
      </Form>
    </main>
  )
}
