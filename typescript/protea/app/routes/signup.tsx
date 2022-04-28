import type { ActionFunction, LoaderFunction } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { Form, useActionData, useLoaderData } from '@remix-run/react'
import type {
  GetCountriesQuery,
  GetCountriesQueryVariables
} from '~/generated/types'
import { GetCountriesDocument } from '~/generated/types'
import { Autocomplete, Button, Logo, Router, TextField } from '~/components'
import React, { useEffect, useState } from 'react'
import { route } from 'routes-gen'
import {
  KRATOS_URL,
  getCsrfTokenFromFlow,
  handleFlowError,
  requireNoUserSession
} from '~/lib/kratos.server'
import { apolloClient } from '~/lib/apollo.server'

type Country = {
  id: string
  name: string
}

type ActionData = {
  formError?: string
  fieldErrors?: {
    country?: string
    email?: string
    password?: string
  }
  fields?: {
    country: string
    email: string
    password: string
    csrf_token: string
  }
}

const badRequest = (data: ActionData) => json(data, { status: 400 })

export const action: ActionFunction = async ({ request }) => {
  const url = new URL(request.url)
  const flowId = url.searchParams.get('flow')
  const form = await request.formData()
  const csrfToken = form.get('csrf_token')
  const country = form.get('country')
  const email = form.get('email')
  const password = form.get('password')

  if (
    typeof csrfToken !== 'string' ||
    typeof country !== 'string' ||
    typeof email !== 'string' ||
    typeof password !== 'string'
  ) {
    return badRequest({
      formError: `Form not submitted correctly.`
    })
  }

  const fields = { csrf_token: csrfToken, email, password, country }
  const res = await fetch(
    `${KRATOS_URL}/self-service/registration?flow=${flowId}`,
    {
      method: 'POST',
      body: JSON.stringify({
        method: 'password',
        traits: {
          email: email
        },
        password: password,
        csrf_token: csrfToken
      }),
      headers: {
        'Content-type': 'application/json',
        cookie: String(request.headers.get('cookie'))
      }
    }
  )

  const data = await res.json()

  if (res.status >= 400) {
    let fieldErrors: ActionData['fieldErrors'] = {}
    for (let node of data.ui.nodes) {
      if (node.messages.length > 0) {
        Object.assign(fieldErrors, {
          [node.attributes.name]: node.messages[0].text
        })
      }
    }
    // TODO: add formError here and catch on frontend
    return badRequest({ fieldErrors: fieldErrors, fields })
  }
  return redirect(route('/verify'), {
    headers: res.headers
  })
}

export const loader: LoaderFunction = async ({ request }) => {
  await requireNoUserSession(request)
  const cookie = String(request.headers.get('cookie'))

  const url = new URL(request.url)
  const flowId = url.searchParams.get('flow')

  const countries = await apolloClient
    .query<GetCountriesQuery, GetCountriesQueryVariables>({
      query: GetCountriesDocument,
      context: {
        headers: request.headers
      }
    })
    .then((val) => val.data.countries)

  let flow
  if (flowId) {
    // If ?flow=.. was in the URL, we fetch it
    const flowRes = await fetch(
      `${KRATOS_URL}/self-service/registration/flows?id=${flowId}`,
      {
        headers: {
          cookie: cookie,
          Accept: 'application/json'
        }
      }
    )
    flow = await flowRes.json()
    if (flowRes.status >= 400) handleFlowError(flow, 'signup')
  } else {
    // Otherwise we initialize it
    const flowRes = await fetch(
      `${KRATOS_URL}/self-service/registration/browser?${url.searchParams}`,
      { headers: { Accept: 'application/json' } }
    )
    flow = await flowRes.json()
    if (flowRes.status >= 400) handleFlowError(flow, 'signup')
    return redirect(`/signup?flow=${flow.id}`, {
      headers: flowRes.headers
    })
  }
  return json({ flow, countries, csrfToken: getCsrfTokenFromFlow(flow) })
}

export default function Page() {
  const actionData = useActionData<ActionData>()
  const { flow, countries, csrfToken } = useLoaderData()

  const [country, setCountry] = useState<Country>()
  const [query, setQuery] = useState<string>('')
  const [filteredCountries, setFilteredCountries] = useState(countries)

  useEffect(() => {
    if (query === '') setFilteredCountries(countries)
    else {
      setFilteredCountries(
        countries.filter((country: Country) => {
          return (
            country.name
              .toLowerCase()
              .replace(/\s+/g, '')
              .includes(query.toLowerCase().replace(/\s+/g, '')) ||
            country.id
              .toLowerCase()
              .replace(/\s+/g, '')
              .includes(query.toLowerCase().replace(/\s+/g, ''))
          )
        })
      )
    }
  }, [query, countries])

  return (
    <main className='mx-auto grid min-h-screen w-full grid-cols-4 content-start gap-4 gap-y-2 overflow-y-auto p-4 sm:max-w-lg sm:grid-cols-8 sm:px-0 lg:max-w-3xl lg:grid-cols-12 lg:content-center xl:max-w-4xl'>
      <div className='col-span-full sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <Router to={route('/')}>
          <Logo className='h-8' />
        </Router>
      </div>
      <div className='col-span-full pt-4 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <h1 className='font-display text-4xl font-medium leading-normal text-strong'>
          Create a new account
        </h1>
      </div>
      <div className='col-span-full pb-8 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <p className='text-medium'>
          Or{' '}
          <Router to={route('/login')}>
            <span className='text-primary'>
              sign in to an existing account.
            </span>
          </Router>
        </p>
      </div>
      {/* Form */}
      <Form
        action={`/signup?flow=${flow.id}`}
        method='post'
        className='col-span-full flex flex-col items-end space-y-2 sm:col-span-6 sm:col-start-2 lg:col-start-4'
      >
        <Autocomplete
          id='country'
          value={country}
          onChange={setCountry}
          onQuery={setQuery}
          options={filteredCountries}
          label='Country'
          aria-invalid={Boolean(actionData?.fieldErrors?.country) || undefined}
          aria-describedby={
            actionData?.fieldErrors?.country ? 'country-error' : undefined
          }
          errorMessage={actionData?.fieldErrors?.country}
        />
        <input value={String(country?.id)} name='country' type='hidden' />

        <TextField
          id='email'
          label='Email'
          enterKeyHint='next'
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

        {/* <div className='flex w-full items-center space-x-3 rounded-xl bg-container p-4'>
          <div className='text-medium'>
            <Icon>tips_and_updates</Icon>
          </div>
          <span className='text-small font-normal text-medium'>
            We currently aren't released in your region. Feel free to join our
            waiting list.
          </span>
        </div> */}

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
          errorMessage={actionData?.fieldErrors?.password}
        />

        <input defaultValue={csrfToken} name='csrf_token' type='hidden' />
        <div className='pt-4'>
          <Button type='submit'>Create account</Button>
        </div>
      </Form>
    </main>
  )
}
