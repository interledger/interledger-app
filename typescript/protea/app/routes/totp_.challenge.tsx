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
import { getCsrfTokenFromFlow } from '~/lib/kratos.server'
import { mergeMeta } from '~/lib/meta'
import { safeReturnTo } from '~/lib/url.server'
import { kratosPublic } from '~/lib/kratos/kratos-client.server'
import { buildHeadersWithCookies, getCookie, withCookie } from '~/lib/kratos/cookie.util'
import { mapFlowToFieldErrors } from '~/lib/kratos/error'
import { CreateBrowserLoginFlowResponse, KratosError } from '~/lib/kratos/types'

export type TotpAction =
  | {
    errors: {
      totp_code?: string
    }
  }
  | undefined

const getFlowFromRedirect = (flowResponse: CreateBrowserLoginFlowResponse) => {
  if (flowResponse.status === 200 && flowResponse.data?.id) {
    return flowResponse.data.id
  }

  if (flowResponse.status === 303 || flowResponse.status === 302) {
    const location = flowResponse.headers["location"]
    if (!location) {
      throw new Error('Expected redirect with flow ID, but got none.')
    }
    const flowFromRedirect = new URL(location).searchParams.get('flow')
    if (!flowFromRedirect) {
      throw new Error('No redirect from flow')
    }
    return flowFromRedirect
  }

  throw new Error('Expected redirect or 200 OK with flow data from Kratos')
}

export async function loader({ request }: LoaderFunctionArgs) {
  const url = new URL(request.url)
  const flowId = url.searchParams.get('flow')
  const returnTo = safeReturnTo(url.searchParams.get('returnTo'))
  const cookie = request.headers.get('cookie')

  if (!cookie) {
    // todo: create a must have cookies guard for most paths
    throw redirect("/")
  }

  try {
    const session = await kratosPublic.toSession({ cookie })
    if (session.data.authenticator_assurance_level === 'aal2') {
      return redirect(returnTo)
    }
  } catch (_) { }

  if (!flowId) {
    let aal2Flow: CreateBrowserLoginFlowResponse
    try {
      aal2Flow = await kratosPublic.createBrowserLoginFlow({
        aal: 'aal2',
      }, withCookie(cookie))
    } catch (err) {
      const kratosError = err as KratosError
      const errorIdData = kratosError.response?.data?.error?.id
      const errorId = kratosError.id
      const status = kratosError.response?.status
      throw new Error(`Error initializing flow, errorIdData ${errorIdData} and status ${status} and errorId ${errorId}`)
    }

    const flowFromRedirect = getFlowFromRedirect(aal2Flow)

    const searchParams = new URLSearchParams()
    searchParams.set('returnTo', returnTo)
    searchParams.set('flow', flowFromRedirect)
    const headers = buildHeadersWithCookies(aal2Flow)

    return redirect(`${route('/totp/challenge')}?${searchParams.toString()}`, { headers })
  }

  let loginFlow;
  try {
    loginFlow = await kratosPublic.getLoginFlow({ id: flowId, cookie })
  } catch (err) {
    const kratosError = err as KratosError
    const status = kratosError.response?.status

    switch (status) {
      case 404:
        console.log("Invalid FLOW ID, redirect to blank challenge")
        throw redirect("/totp/challenge")
      default:
        console.error("Unknown error getting flow on getLoginFlow, redirect to blank challenge")
        throw redirect("/totp/challenge")
    }
  }

  const flow: SelfServiceLoginFlow = loginFlow.data
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
  const cookie = getCookie(request)
  const returnTo = new URL(request.url).searchParams.get('returnTo') || '/'

  await validateCSRFToken(request, form)

  try {
    const submitTotpResponse = await kratosPublic.updateLoginFlow({
      flow: flow as string,
      updateLoginFlowBody: {
        method: 'totp',
        totp_code: totp_code as string,
        csrf_token: csrf_token as string
      }
    }, withCookie(cookie))

    if (submitTotpResponse.data.session) {
      const headers = buildHeadersWithCookies(submitTotpResponse)
      return redirect(returnTo ?? '/', { headers })
    }

    console.error("No session after updateLoginFlow, status", submitTotpResponse.status)
    return redirect("/totp/challenge")
  } catch (err) {
    const kratosError = err as KratosError
    const flowData = kratosError.response.data
    const flowStatus = kratosError.response.status

    switch (flowStatus) {
      case 400:
        const errorMapping = { form: '' }
        mapFlowToFieldErrors(flowData, errorMapping)
        return json({ errors: { totp_code: errorMapping.form } })

      case 410:
        // Flow expired
        throw redirect("/totp/challenge")

      default:
        console.error("Unknwon case when updateLoginFlow")
        throw redirect("/totp/challenge")
    }

  }



}
