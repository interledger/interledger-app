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
  Layouts
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

export const handle: ApplicationProps = {
  layout: Layouts.Focus,
  scaffold: {
    header: {}
  }
}

export const meta: MetaFunction = mergeMeta(() => [
  {
    title: 'Kratos - Logout Flow Test'
  }
])

export async function loader({ request }: LoaderFunctionArgs) {
  const cookie = getCookie(request)

  try {
    // Create logout flow using SDK
    const response = await kratosPublic.createBrowserLogoutFlow(
      {},
      withCookie(cookie)
    )
    const flow = response.data
    const headers = buildHeadersWithCookies(response)

    return json(
      {
        logoutToken: flow.logout_token,
        logoutUrl: flow.logout_url,
        flowData: JSON.stringify(flow, null, 2),
        success: true,
        error: null
      },
      { headers }
    )
  } catch (err: any) {
    return json({
      logoutToken: null,
      logoutUrl: null,
      flowData: null,
      success: false,
      error: err.message || 'Failed to create logout flow'
    })
  }
}

export default function KratosLogoutTest() {
  const loaderData = useLoaderData<typeof loader>()
  const actionData = useActionData<typeof action>()

  return (
    <>
      <Card>
        <CardHeader>
          <CardTitle>Logout Flow Test - SDK</CardTitle>
        </CardHeader>
        <CardContent>
          <p className='mb-4 text-sm text-medium'>
            Test the Kratos logout flow using the official SDK.
          </p>

          {loaderData.success ? (
            <div className='space-y-4'>
              <div className='rounded-lg bg-mercury p-4'>
                <p className='text-sm'>
                  <strong>Logout Token:</strong>{' '}
                  <code className='text-xs'>{loaderData.logoutToken}</code>
                </p>
                <p className='mt-2 text-sm'>
                  <strong>Logout URL:</strong>{' '}
                  <code className='text-xs'>{loaderData.logoutUrl}</code>
                </p>
              </div>

              <Form method='post'>
                <input
                  type='hidden'
                  name='logout_token'
                  value={loaderData.logoutToken!}
                />
                <Button type='submit'>Execute Logout</Button>
              </Form>
            </div>
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
  const logoutToken = formData.get('logout_token')
  const cookie = getCookie(request)

  if (!logoutToken) {
    return json({
      success: false,
      error: 'Missing logout token'
    })
  }

  try {
    // Execute logout using SDK
    const response = await kratosPublic.updateLogoutFlow(
      {
        token: String(logoutToken)
      },
      withCookie(cookie)
    )

    const headers = buildHeadersWithCookies(response)

    return json(
      {
        success: true,
        message: 'Logout successful',
        status: response.status
      },
      { headers }
    )
  } catch (err: any) {
    console.error('Logout error:', err)

    return json({
      success: false,
      error: err.response?.data || err.message
    })
  }
}
