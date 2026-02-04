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
  kratosPublic,
  getCsrfTokenFromFlow,
  withCookie,
  getCookie,
  buildHeadersWithCookies
} from '~/lib/kratos-client.server'

export const handle: ApplicationProps = {
  layout: Layouts.Focus,
  scaffold: {
    header: {}
  }
}

export const meta: MetaFunction = mergeMeta(() => [
  {
    title: 'Kratos - Login Flow Test'
  }
])

export async function loader({ request }: LoaderFunctionArgs) {
  const url = new URL(request.url)
  const flowId = url.searchParams.get('flow')
  const cookie = getCookie(request)

  let flow
  let headers: Headers | undefined = undefined

  try {
    if (flowId) {
      // Fetch existing flow using SDK
      const response = await kratosPublic.getLoginFlow(
        { id: flowId },
        withCookie(cookie)
      )
      flow = response.data
    } else {
      // Initialize new flow using SDK
      const response = await kratosPublic.createBrowserLoginFlow(
        {},
        withCookie(cookie)
      )
      flow = response.data
      headers = buildHeadersWithCookies(response)
    }

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
      error: err.message || 'Failed to initialize login flow'
    })
  }
}

export default function KratosLoginTest() {
  const loaderData = useLoaderData<typeof loader>()
  const actionData = useActionData<typeof action>()

  return (
    <>
      <Card>
        <CardHeader>
          <CardTitle>Login Flow Test - SDK</CardTitle>
        </CardHeader>
        <CardContent>
          <p className='mb-4 text-sm text-medium'>
            Test the Kratos login flow using the official SDK.
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
                id='password'
                name='password'
                label='Password'
                type='password'
                required
                errorMessage={actionData?.errors?.password}
              />

              <Button type='submit'>Submit Login</Button>
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
  const password = formData.get('password')
  const csrfToken = formData.get('csrf_token')
  const flowId = formData.get('flow_id')
  const cookie = getCookie(request)

  const fieldErrors = {
    form: '',
    email: '',
    password: ''
  }

  if (!email || !password || !csrfToken || !flowId) {
    return json({
      success: false,
      errors: {
        ...fieldErrors,
        form: 'Missing required fields'
      }
    })
  }

  try {
    // Submit login using SDK
    const response = await kratosPublic.updateLoginFlow(
      {
        flow: String(flowId),
        updateLoginFlowBody: {
          method: 'password',
          identifier: String(email),
          password: String(password),
          csrf_token: String(csrfToken)
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
    console.error('Login error:', err)

    // Handle Kratos validation errors
    if (err.response?.data?.ui) {
      const kratosData = err.response.data
      
      // Extract field errors from UI nodes
      for (const node of kratosData.ui.nodes || []) {
        if (node.messages && node.messages.length > 0) {
          const fieldName = node.attributes?.name
          if (fieldName && fieldName in fieldErrors) {
            fieldErrors[fieldName as keyof typeof fieldErrors] = node.messages[0].text
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
