import type { ActionFunctionArgs, LoaderFunctionArgs, MetaFunction } from 'react-router';
import { data, redirect } from 'react-router';
import { Form, useLoaderData } from 'react-router';
import { href } from 'react-router'
import type { ApplicationProps } from '~/components'
import { Button, Card, CardContent, Layouts } from '~/components'
import { jsonWithCSRF, validateCSRFToken } from '~/lib/csrf.server'
import { trimHeaders } from '~/lib/headers.server'
import { KRATOS_URL, handleFlowError } from '~/lib/kratos.server'
import { mergeMeta } from '~/lib/meta'
import { destroySession, getSession } from '~/session.server'

export async function loader({ request }: LoaderFunctionArgs) {
  const cookie = String(request.headers.get('cookie'))
  let flow
  const flowRes = await fetch(`${KRATOS_URL}/self-service/logout/browser`, {
    headers: {
      cookie: cookie,
      Accept: 'application/json'
    }
  })
  flow = await flowRes.json()
  if (flowRes.status >= 400) handleFlowError(flow, 'logout')
  return jsonWithCSRF(request, { logoutToken: flow.logout_token })
}

export const handle: ApplicationProps = {
  layout: Layouts.Focus,
  scaffold: {
    header: {
      back: href('/'),
      title: 'Log out'
    }
  }
}

export const meta: MetaFunction = mergeMeta(() => [
  {
    title: 'Log out'
  }
])

export default function Page() {
  const { logoutToken, csrfToken } = useLoaderData<typeof loader>()

  return (
    <>
      <Form id='logout' action='/logout' method='post' className='hidden' />
      <input form='logout' value={csrfToken} name='csrfToken' type='hidden' />
      <Card>
        <CardContent>
          <p>Are you sure you want to log out?</p>
        </CardContent>
      </Card>
      <Button
        type='submit'
        form='logout'
        name='logoutToken'
        value={logoutToken}
      >
        Log out
      </Button>
    </>
  )
}

export async function action({ request }: ActionFunctionArgs) {
  const cookie = request.headers.get('cookie') as string
  const form = await request.formData()
  const token = form.get('logoutToken')

  await validateCSRFToken(request, form)

  const res = await fetch(`${KRATOS_URL}/self-service/logout?token=${token}`, {
    method: 'GET',
    headers: {
      Accept: 'application/json',
      cookie
    }
  })

  if (res.status >= 400) {
    return data(
      { errors: { form: 'Something went wrong trying to logout.' } },
      { status: 400 }
    )
  }

  const session = await getSession(cookie)
  const sessionHeaders = await destroySession(session)

  // Remove all headers besides set-cookie
  const headers = trimHeaders(res.headers, ['set-cookie'])
  headers.append('Set-Cookie', sessionHeaders)

  return redirect(href('/'), {
    headers: headers
  })
}
