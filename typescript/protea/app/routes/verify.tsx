import { json, Form, redirect, useLoaderData } from 'remix'
import type { ActionFunction, LoaderFunction } from 'remix'
import { Button, Logo, Router } from '~/components'
import axios, { AxiosError } from 'axios'
import React from 'react'
import { route } from 'routes-gen'
import { getCsrfTokenFromFlow, kratos, handleFlowError } from '~/lib/kratos'
import { Session } from '@ory/kratos-client'

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
      `http://kratos-public/self-service/verification?flow=${flowId}`,
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
    .then((res) =>
      redirect(route('/verify'), {
        headers: res.headers
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
  // return kratos
  //   .submitSelfServiceVerificationFlow(String(flowId), undefined, {
  //     method: 'link',
  //     email: email,
  //     csrf_token: csrfToken
  //   })
  //   .then(() => redirect(route('/verify')))
  //   .catch(handleFlowError('verify'))
  //   .catch((err: AxiosError) => {
  //     // If the previous handler did not catch the error it's most likely a form validation error
  //     if (err.response?.status === 400) {
  //       // Yup, it is!
  //       return badRequest({ fieldErrors: err.response?.data, fields })
  //     }
  //     return err
  //   })
}

export const loader: LoaderFunction = async ({ request }) => {
  const url = new URL(request.url)
  const flowId = url.searchParams.get('flow')
  const returnTo = url.searchParams.get('return_to')
  const cookie = request.headers.get('cookie') || undefined

  const session = await kratos
    .toSession(undefined, cookie)
    .then((res) => {
      const session = res.data as Session

      // Check the user has at least one verifiable address.
      if (!session.identity.verifiable_addresses)
        return redirect(route('/signup'))
      // We currently only allow one email per user.
      if (session.identity.verifiable_addresses[0].verified)
        return redirect(route('/home'))

      return session.identity.verifiable_addresses[0].value
    })
    .catch((err) => {
      switch ((err as AxiosError)?.response?.status) {
        case 403:
        case 422: // Need to complete 2FA.
          return redirect(route('/login') + '?aal=aal2')
      }
      return redirect(route('/login'))
    })

  // Ensure any redirects are thrown
  if (session instanceof Response) return session

  // TODO: get flow from kratos and handle flow errors appropriately.
  // If ?flow=.. was in the URL, we fetch it
  if (flowId) {
    return kratos
      .getSelfServiceVerificationFlow(String(flowId), cookie)
      .then((res) => {
        return json({ flow: res.data, email: String(session) })
      })
      .catch(handleFlowError('verify'))
  }

  // Otherwise we initialize it
  return kratos
    .initializeSelfServiceVerificationFlowForBrowsers(
      returnTo ? String(returnTo) : undefined
    )
    .then(async (res) =>
      redirect(`/verify?flow=${res.data.id}`, {
        headers: res.headers
      })
    )
    .catch(handleFlowError('verify'))
}

export default function VerifyPage() {
  const loaderData = useLoaderData()

  return (
    <main className='mx-auto grid min-h-screen w-full grid-cols-4 content-start gap-4 gap-y-2 overflow-y-auto p-4 sm:max-w-lg sm:grid-cols-8 sm:px-0 lg:max-w-3xl lg:grid-cols-12 lg:content-center xl:max-w-4xl'>
      <div className='col-span-full sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <Router to={route('/')}>
          <Logo className='h-8' />
        </Router>
      </div>
      <div className='col-span-full pt-4 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <h1 className='font-display text-4xl font-medium leading-normal text-strong'>
          Verify your email
        </h1>
      </div>
      <div className='col-span-full pb-8 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <p className='text-medium'>
          We’ve sent a verification link to your email:
          <br /> {loaderData.email}
        </p>
      </div>
      {/* Form */}
      <Form
        action={`/verify?flow=${loaderData.flow.id}`}
        method='post'
        className='col-span-full flex flex-col items-end space-y-2 sm:col-span-6 sm:col-start-2 lg:col-start-4'
      >
        <input
          defaultValue={getCsrfTokenFromFlow(loaderData.flow)}
          name='csrf_token'
          type='hidden'
        />
        <input defaultValue={loaderData.email} name='email' type='hidden' />
        <div className='pt-4'>
          <Button type='submit'>Resend verification</Button>
        </div>
      </Form>
    </main>
  )
}
