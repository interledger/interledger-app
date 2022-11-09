import type { LoaderArgs, ActionArgs } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { Form, useLoaderData } from '@remix-run/react'
import { Button, Layouts } from '~/components'
import { KRATOS_URL, handleFlowError } from '~/lib/kratos.server'
import { trimHeaders } from '~/lib/headers.server'
import { route } from 'routes-gen'
import { destroySession, getSession } from '~/session.server'

export async function loader({ request }: LoaderArgs) {
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
  return json({ logoutToken: flow.logout_token })
}

export const handle = {
  layout: Layouts.FocusLayout
}

export default function Page() {
  const { logoutToken } = useLoaderData<typeof loader>()

  return (
    <div className='flex w-full flex-col rounded-2xl bg-page p-4 pb-8'>
      <h1 className='mb-6 font-display text-2xl font-medium'>Log out</h1>
      <span>Are you sure you want to log out?</span>

      <Form id='logout' action='/logout' method='post' className='hidden' />
      <div className='mt-6'>
        <Button
          type='submit'
          form='logout'
          name='logoutToken'
          value={logoutToken}
        >
          Log out
        </Button>
      </div>
    </div>
  )
}

export async function action({ request }: ActionArgs) {
  const cookie = request.headers.get('cookie') as string
  const form = await request.formData()
  const token = form.get('logoutToken')

  const res = await fetch(`${KRATOS_URL}/self-service/logout?token=${token}`, {
    method: 'GET',
    headers: {
      Accept: 'application/json',
      cookie
    }
  })

  if (res.status >= 400) {
    return json(
      { errors: { form: 'Something went wrong trying to logout.' } },
      { status: 400 }
    )
  }

  const session = await getSession(cookie)
  const sessionHeaders = await destroySession(session)

  // Remove all headers besides set-cookie
  const headers = trimHeaders(res.headers, ['set-cookie'])
  headers.append('Set-Cookie', sessionHeaders)

  return redirect(route('/'), {
    headers: headers
  })
}
