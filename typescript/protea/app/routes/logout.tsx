import { Form, data, href, redirect, useLoaderData } from 'react-router'
import type { ApplicationProps } from '~/components'
import { Button, Card, CardContent, Layouts } from '~/components'
import { jsonWithCSRF, validateCSRFToken } from '~/lib/csrf.server'
import { buildHeadersWithCookies, getCookie } from '~/lib/kratos/cookie.server'
import { handleFlowError } from '~/lib/kratos/error.server'
import { kratosPublic } from '~/lib/kratos/kratos-client.server'
import { mergeMeta } from '~/lib/meta'
import type { Route } from './+types/logout'

export async function loader({ request }: Route.LoaderArgs) {
  const cookie = getCookie(request)

  try {
    const logoutFlow = await kratosPublic.createBrowserLogoutFlow({ cookie })
    return jsonWithCSRF(request, { logoutToken: logoutFlow.data.logout_token })
  } catch (err) {
    handleFlowError(err, 'logout')
    throw redirect('/')
  }
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

export const meta = mergeMeta(() => [
  {
    title: 'Log out'
  }
])

export default function Page() {
  const { logoutToken, csrfToken } = useLoaderData()

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

export async function action({ request }: Route.ActionArgs) {
  const cookie = getCookie(request)
  const form = await request.formData()
  const token = form.get('logoutToken')

  if (typeof token !== 'string') {
    return data(
      { errors: { form: 'Something went wrong trying to logout.' } },
      { status: 400 }
    )
  }

  await validateCSRFToken(request, form)

  try {
    const logoutResponse = await kratosPublic.updateLogoutFlow({
      token,
      cookie
    })
    const headers = buildHeadersWithCookies(logoutResponse)

    return redirect(href('/'), { headers })
  } catch (error) {
    return data(
      { errors: { form: 'Something went wrong trying to logout.' } },
      { status: 400 }
    )
  }
}
