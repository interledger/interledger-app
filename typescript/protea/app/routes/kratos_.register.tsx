import type {
  ActionFunctionArgs,
  LoaderFunctionArgs,
  MetaFunction
} from '@remix-run/node'
import { json } from '@remix-run/node'
import { Form, useActionData, useLoaderData } from '@remix-run/react'
import type { ApplicationProps } from '~/components'
import {
  Button,
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  Layouts,
  TextField
} from '~/components'
import { mergeMeta } from '~/lib/meta'
import {
  kratosPublic
} from '~/lib/kratos/kratos-client.server'
import {
  withCookie,
  getCookie,
  buildHeadersWithCookies
} from '~/lib/kratos/cookie.util'
import { getCsrfTokenFromFlow } from '~/lib/kratos/flow.util'

export const handle: ApplicationProps = {
  layout: Layouts.Focus,
  scaffold: {
    header: {}
  }
}

export const meta: MetaFunction = mergeMeta(() => [
  {
    title: 'Kratos - Registration Flow Test'
  }
])

export async function loader({ request }: LoaderFunctionArgs) {
  const url = new URL(request.url)
  const cookie = getCookie(request)

  try {
    // Initialize registration flow using SDK
    const response = await kratosPublic.createBrowserRegistrationFlow(
      {},
      withCookie(cookie)
    )
    const flow = response.data
    const headers = buildHeadersWithCookies(response)

    return json(
      {
        flowId: flow.id,
        csrfToken: getCsrfTokenFromFlow(flow),
        flowData: JSON.stringify(flow, null, 2),
        success: true,
        error: null
      },
      { headers }
    )
  } catch (err: any) {
    return json({
      flowId: null,
      csrfToken: '',
      flowData: null,
      success: false,
      error: err.message || 'Failed to initialize registration flow'
    })
  }
}

export default function KratosRegisterTest() {
  const loaderData = useLoaderData<typeof loader>()
  const actionData = useActionData<typeof action>()

  return (
    <>
      <Card>
        <CardHeader>
          <CardTitle>Registration Flow Test - SDK</CardTitle>
        </CardHeader>
        <CardContent>
          <p className='mb-4 text-sm text-medium'>
            Test the Kratos registration flow using the official SDK.
          </p>

          {loaderData.success ? (
            <Form method='post' className='space-y-4'>
              <input
                type='hidden'
                name='csrf_token'
                value={loaderData.csrfToken}
              />
              <input type='hidden' name='flow_id' value={loaderData.flowId!} />

              <TextField
                id='email'
                name='email'
                label='Email'
                type='email'
                required
                errorMessage={actionData?.errors?.email}
              />

              <TextField
                id='firstName'
                name='firstName'
                label='First Name'
                type='text'
                required
                errorMessage={actionData?.errors?.firstName}
              />

              <TextField
                id='lastName'
                name='lastName'
                label='Last Name'
                type='text'
                required
                errorMessage={actionData?.errors?.lastName}
              />

              <TextField
                id='phone'
                name='phone'
                label='Phone'
                type='tel'
                errorMessage={actionData?.errors?.phone}
              />

              <TextField
                id='countryCode'
                name='countryCode'
                label='Country Code'
                type='text'
                placeholder='US'
                errorMessage={actionData?.errors?.countryCode}
              />

              <TextField
                id='password'
                name='password'
                label='Password'
                type='password'
                required
                errorMessage={actionData?.errors?.password}
              />

              <Button type='submit'>Submit Registration</Button>
            </Form>
          ) : (
            <div className='rounded-lg bg-red-50 p-4 text-red-700'>
              Error: {loaderData.error}
            </div>
          )}
        </CardContent>
      </Card>

      {loaderData.flowData && (
        <Card className='mt-6'>
          <CardHeader>
            <CardTitle>Flow Data (Loader)</CardTitle>
          </CardHeader>
          <CardContent>
            <pre className='overflow-auto rounded bg-mercury p-4 text-xs'>
              {loaderData.flowData}
            </pre>
          </CardContent>
        </Card>
      )}

      {actionData && (
        <Card className='mt-6'>
          <CardHeader>
            <CardTitle>
              Response Data (Action) -{' '}
              {actionData.success ? 'Success' : 'Error'}
            </CardTitle>
          </CardHeader>
          <CardContent>
            <pre className='overflow-auto rounded bg-mercury p-4 text-xs'>
              {JSON.stringify(actionData, null, 2)}
            </pre>
          </CardContent>
        </Card>
      )}
    </>
  )
}

export async function action({ request }: ActionFunctionArgs) {
  const formData = await request.formData()
  const email = formData.get('email')
  const firstName = formData.get('firstName')
  const lastName = formData.get('lastName')
  const phone = formData.get('phone')
  const countryCode = formData.get('countryCode')
  const password = formData.get('password')
  const csrfToken = formData.get('csrf_token')
  const flowId = formData.get('flow_id')
  const cookie = getCookie(request)

  const fieldErrors = {
    form: '',
    email: '',
    firstName: '',
    lastName: '',
    phone: '',
    countryCode: '',
    password: ''
  }

  if (!email || !firstName || !lastName || !password || !csrfToken || !flowId) {
    return json({
      success: false,
      errors: {
        ...fieldErrors,
        form: 'Missing required fields'
      }
    })
  }

  const clean = (val: any) => String(val).replace(/[\n\r\t]/g, '').trim()

  try {
    // Submit registration using SDK
    const response = await kratosPublic.updateRegistrationFlow(
      {
        flow: clean(flowId),
        updateRegistrationFlowBody: {
          method: 'password',
          traits: {
            email: clean(email),
            firstName: clean(firstName),
            lastName: clean(lastName),
            phone: phone ? clean(phone) : undefined,
            countryCode: countryCode ? clean(countryCode) : undefined
          },
          password: String(password),
          csrf_token: clean(csrfToken)
        }
      },
      withCookie(cookie)
    )

    const headers = buildHeadersWithCookies(response)

    return json(
      {
        success: true,
        data: response.data,
        status: response.status,
        errors: null
      },
      { headers }
    )
  } catch (err: any) {
    console.error('Registration error:', err)

    // Handle Kratos validation errors
    if (err.response?.data?.ui) {
      const kratosData = err.response.data

      // Extract field errors from UI nodes
      for (const node of kratosData.ui.nodes || []) {
        if (node.messages && node.messages.length > 0) {
          const fieldName = node.attributes?.name
          if (fieldName) {
            // Handle nested trait fields (e.g., traits.email)
            const simplifiedName = fieldName.replace('traits.', '')
            if (simplifiedName in fieldErrors) {
              fieldErrors[simplifiedName as keyof typeof fieldErrors] = node.messages[0].text
            }
          }
        }
      }

      // Extract form-level errors
      if (kratosData.ui.messages && kratosData.ui.messages.length > 0) {
        fieldErrors.form = kratosData.ui.messages[0].text
      }
    }

    return json({
      success: false,
      errors: fieldErrors,
      errorDetails: err.response?.data || err.message
    })
  }
}
