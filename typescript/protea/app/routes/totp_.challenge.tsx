import type { SelfServiceLoginFlow } from '~/lib/kratos.server'
import type {
  ActionFunctionArgs,
  LoaderFunctionArgs,
  MetaFunction
} from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { useActionData, useLoaderData } from '@remix-run/react'
import { route } from 'routes-gen'
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
import { buildHeadersWithCookies, kratosPublic } from '~/lib/kratos-client.server'

export type TotpAction =
  | {
      errors: {
        totp_code?: string
      }
    }
  | undefined

export async function loader({ request }: LoaderFunctionArgs) {
  console.log("🪲 [totp/challenge] Loader headers", request.headers)
  const url = new URL(request.url)
  const flowId = url.searchParams.get('flow')
  const returnTo = safeReturnTo(url.searchParams.get('returnTo'))
  const cookie = String(request.headers.get('cookie'))
  console.log("🪲 [totp/challenge] Cookie", cookie)
  const refresh = url.searchParams.get('refresh')

  if (!flowId) {
    let initRes: any;
    try {
      initRes = await kratosPublic.createBrowserLoginFlow({
      aal: 'aal2',
      cookie
    })
    } catch(err) {
      throw new Error('Error initializing flow')
    }

    console.info("🪲 [totp/challenge] Init response")
    
    let flowFromRedirect: string | null | undefined

    if (initRes.status === 200 && initRes.data?.id) {
       console.log("🪲 [totp/challenge] Received 200 OK with flow data", initRes.data.id)
       flowFromRedirect = initRes.data.id
    } else if (initRes.status === 303 || initRes.status === 302) {
      const location = initRes.headers.get('location')
      console.log("🪲 [totp/challenge] Location", location)
      if (!location) {
        throw new Error('Expected redirect with flow ID, but got none.')
      }
      flowFromRedirect = new URL(location).searchParams.get('flow')
    } else {
      console.error("🪲 [totp/challenge] Unexpected response from Kratos", initRes)
      throw new Error('Expected redirect or 200 OK with flow data from Kratos')
    }

    console.log("🪲 [totp/challenge] Flow from redirect", flowFromRedirect)
    if (!flowFromRedirect) {
      throw new Error('Redirect did not include flow parameter')
    }

    const searchParams = new URLSearchParams()
    searchParams.set('returnTo', returnTo)
    searchParams.set('flow', flowFromRedirect)
    const headers = buildHeadersWithCookies(initRes)

    console.log("🪲 [totp/challenge] Redirecting to", `${route('/totp/challenge')}?${searchParams.toString()}`)
    return redirect(`${route('/totp/challenge')}?${searchParams.toString()}`, { headers })
  }

  console.log("🪲 [totp/challenge] Fetching flow", flowId)
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
    console.error("🪲 [totp/challenge] Expected redirect response from Kratos", kratosFlow)
    return redirect('/error')
  }

  const flow: SelfServiceLoginFlow = await kratosFlow.json()
  return json({ flowId, csrfToken: getCsrfTokenFromFlow(flow) })
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
      <OutlineButtonRouter to={route('/logout')} className='mt-4'>
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
    return json({
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
