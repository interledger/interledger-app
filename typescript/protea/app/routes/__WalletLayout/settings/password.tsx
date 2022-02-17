import { ActionFunction, LoaderFunction, redirect, useLoaderData } from 'remix'
import { useActionData, json, Form } from 'remix'
import { Button, Logo, Router, TextField } from '~/components'
import axios, { AxiosError, AxiosResponse } from 'axios'
import React from 'react'
import { route } from 'routes-gen'
import { getCsrfTokenFromFlow, kratos, handleFlowError } from '~/lib/kratos'

type ActionData = {
  formError?: string
  fieldErrors?: {
    password: string | undefined
  }
  fields?: {
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
  const password = form.get('password')

  if (typeof csrfToken !== 'string' || typeof password !== 'string') {
    return badRequest({
      formError: `Form not submitted correctly.`
    })
  }

  const fields = { csrf_token: csrfToken, password }

  return axios
    .post(
      `http://kratos-public/self-service/settings?flow=${flowId}`,
      {
        method: 'password',
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
      redirect(route('/settings'), {
        headers: setAllCookiesHeaders(res)
      })
    )
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

  // TODO: Require user has session already

  // TODO: get flow from kratos and handle flow errors appropriately.
  // If ?flow=.. was in the URL, we fetch it
  if (flowId) {
    return kratos
      .getSelfServiceSettingsFlow(
        String(flowId),
        undefined,
        request.headers.get('cookie') ?? undefined
      )
      .then((res) => {
        return json(res.data)
      })
      .catch(handleFlowError('settings'))
  }

  // Otherwise we initialize it
  return kratos
    .initializeSelfServiceSettingsFlowForBrowsers(
      returnTo ? String(returnTo) : undefined
    )
    .then(async (res) =>
      redirect(`/settings/password?flow=${res.data.id}`, {
        headers: res.headers
      })
    )
    .catch(handleFlowError('settings'))
}

export default function SettingsPasswordPage() {
  const actionData = useActionData<ActionData>()
  const loaderData = useLoaderData()

  return (
    <main className='mx-auto flex h-screen max-w-sm flex-col items-start justify-center px-4'>
      <Router to={route('/')} aria-label='Fynbos logo'>
        <Logo className='h-8' />
      </Router>
      <h1 className='mt-6 mb-1 font-display text-4xl font-medium leading-normal text-strong'>
        Set a new password
      </h1>
      <p className='mb-10 text-medium'>
        You’ve successfully recovered your account.
        <br />
        Set a new password to continue.
      </p>
      <Form
        action={`/settings/password?flow=${loaderData.id}`}
        method='post'
        className='flex min-w-full flex-col items-end space-y-4'
      >
        <TextField
          id='password'
          label='New password'
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

        <Button type='submit'>Save password</Button>
      </Form>
    </main>
  )
}
