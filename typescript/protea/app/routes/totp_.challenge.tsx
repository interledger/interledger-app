import type { SelfServiceLoginFlow } from '@ory/kratos-client'
import type { ActionFunctionArgs, LoaderFunctionArgs, MetaFunction } from 'react-router';
import { data, redirect } from 'react-router';
import { useActionData, useLoaderData } from 'react-router';
import { href } from 'react-router'
import type { ApplicationProps } from '~/components'
import { Layouts, OutlineButtonRouter } from '~/components'
import { TotpChallenge } from '~/components/TotpChallenge'
import { validateCSRFToken } from '~/lib/csrf.server'
import {
  KRATOS_URL,
  getCsrfTokenFromFlow,
  isSessionAlreadyExitsMessage
} from '~/lib/kratos.server'
import { mergeMeta } from '~/lib/meta'
import { safeReturnTo } from '~/lib/url.server'
export type TotpAction =
  | {
      errors: {
        totp_code?: string
      }
    }
  | undefined

export async function loader({ request }: LoaderFunctionArgs) {
  const url = new URL(request.url)
  const flowId = url.searchParams.get('flow')
  const returnTo = safeReturnTo(url.searchParams.get('returnTo'))
  const cookie = String(request.headers.get('cookie'))
  const refresh = url.searchParams.get('refresh')

  if (!flowId) {
    const initRes = await fetch(
      `${KRATOS_URL}/self-service/login/browser?aal=aal2${
        refresh ? '&refresh=true' : ''
      }`,
      {
        headers: {
          cookie
        },
        redirect: 'manual'
      }
    )

    if (initRes.status !== 303 && initRes.status !== 302) {
      throw new Error('Expected redirect response from Kratos')
    }

    const location = initRes.headers.get('location')
    if (!location) {
      throw new Error('Expected redirect with flow ID, but got none.')
    }

    const flowFromRedirect = new URL(location).searchParams.get('flow')
    if (!flowFromRedirect) {
      throw new Error('Redirect did not include flow parameter')
    }

    const searchParams = new URLSearchParams()
    searchParams.set('returnTo', returnTo)
    searchParams.set('flow', flowFromRedirect)

    return redirect(`${href('/totp/challenge')}?${searchParams.toString()}`)
  }
  const kratosFlow = await fetch(
    `${KRATOS_URL}/self-service/login/flows?id=${flowId}`,
    {
      headers: {
        cookie: request.headers.get('cookie') ?? '',
        Accept: 'application/json'
      }
    }
  )
  if (!kratosFlow.ok) {
    return redirect('/error')
  }

  const flow: SelfServiceLoginFlow = await kratosFlow.json()
  return data({ flowId, csrfToken: getCsrfTokenFromFlow(flow) })
}

export const handle: ApplicationProps = {
  layout: Layouts.Focus,
  scaffold: {
    header: {
      title: 'Enter Authenticator Code'
    }
  }
}

export const meta: MetaFunction = mergeMeta(() => [
  {
    title: 'Enter TOTP Code'
  }
])

export default function Page() {
  const { flowId, csrfToken } = useLoaderData<typeof loader>()
  const actionData: TotpAction = useActionData<typeof action>()

  return (
    <>
      <TotpChallenge
        flowId={flowId}
        csrfToken={csrfToken}
        actionData={actionData}
      />
      <OutlineButtonRouter to={href('/logout')} className='mt-4'>
        Log out
      </OutlineButtonRouter>
    </>
  )
}

export async function action({ request }: ActionFunctionArgs) {
  const form = await request.formData()
  const flow = form.get('flow')
  const totp_code = form.get('totp_code')
  const csrf_token = form.get('csrf_token')

  await validateCSRFToken(request, form)

  const res = await fetch(`${KRATOS_URL}/self-service/login?flow=${flow}`, {
    method: 'POST',
    headers: {
      Accept: 'application/json',
      'Content-type': 'application/json',
      cookie: String(request.headers.get('cookie'))
    },
    body: JSON.stringify({
      method: 'totp',
      totp_code,
      csrf_token
    })
  })

  const returnTo = new URL(request.url).searchParams.get('returnTo') || '/'
  const response = redirect(returnTo ?? '/')

  if (res.status === 400) {
    const data = await res.json()
    // if the form is submitted twice the user already has a valid session
    const message: string =
      data.ui?.messages?.find((m: any) => m.type === 'error')?.text ?? ''
    if (isSessionAlreadyExitsMessage(message)) {
      return response
    }
    return data({
      errors: {
        totp_code: message || 'Invalid code'
      }
    })
  }

  if (!res.ok) {
    throw new Response('Unexpected error', { status: res.status })
  }

  const setCookie = res.headers.get('set-cookie')
  if (setCookie) {
    response.headers.set('cookie', setCookie)
  }

  return response
}
