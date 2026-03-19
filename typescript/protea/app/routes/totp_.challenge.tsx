import type { LoginFlow } from '@ory/client'
import type { Route } from './+types/totp_.challenge'
import { data, redirect, href } from 'react-router'
import { useActionData, useLoaderData } from 'react-router'
import type { ApplicationProps } from '~/components'
import { Layouts, OutlineButtonRouter } from '~/components'
import { TotpChallenge } from '~/components/TotpChallenge'
import { validateCSRFToken } from '~/lib/csrf.server'
import { getCsrfTokenFromFlow } from '~/lib/kratos/flow.server'
import { mergeMeta } from '~/lib/meta'
import { safeReturnTo } from '~/lib/url.server'
import { kratosPublic } from '~/lib/kratos/kratos-client.server'
import { buildHeadersWithCookies, getCookie, withCookie } from '~/lib/kratos/cookie.server'
import { mapFlowToFieldErrors, printKratosError } from '~/lib/kratos/error.server'
import type { CreateBrowserLoginFlowResponse, KratosError } from '~/lib/kratos/types.server'
import logger from '~/lib/logger.server'

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

export async function loader({ request }: Route.LoaderArgs) {
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
      throw new Error(printKratosError(err))
    }

    const flowFromRedirect = getFlowFromRedirect(aal2Flow)

    const searchParams = new URLSearchParams()
    searchParams.set('returnTo', returnTo)
    searchParams.set('flow', flowFromRedirect)
    const headers = buildHeadersWithCookies(aal2Flow)

    return redirect(`${href('/totp/challenge')}?${searchParams.toString()}`, { headers })
  }

  let loginFlow;
  try {
    loginFlow = await kratosPublic.getLoginFlow({ id: flowId, cookie })
  } catch (err) {
    printKratosError(err)
    throw redirect("/totp/challenge")
  }

  const flow: LoginFlow = loginFlow.data
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

export const meta = mergeMeta(() => [
  {
    title: 'Enter TOTP Code'
  }
])

export default function Page() {
  const { flowId, csrfToken } = useLoaderData()
  const actionData: TotpAction = useActionData()

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

export async function action({ request }: Route.ActionArgs) {
  const form = await request.formData()
  const flow = form.get('flow')
  const totp_code = form.get('totp_code')
  const csrf_token = form.get('csrf_token')

  if (!flow || typeof flow !== 'string') {
    throw redirect('/totp/challenge')
  }
  if (!totp_code || typeof totp_code !== 'string') {
    logger.error({ route: "totp.challange" }, "TOTP code is required")
    return data({ errors: { totp_code: "TOTP code is required" } })
  }
  if (!csrf_token || typeof csrf_token !== 'string') {
    logger.error({ route: "totp.challange" }, "CSRF token is required")
    return data({ errors: { totp_code: "Unknown error, please retry." } })
  }

  const cookie = getCookie(request)
  const url = new URL(request.url)
  const returnTo = safeReturnTo(url.searchParams.get('returnTo'))

  await validateCSRFToken(request, form)

  try {
    const submitTotpResponse = await kratosPublic.updateLoginFlow({
      flow: flow as string,
      updateLoginFlowBody: {
        method: 'totp',
        totp_code: totp_code,
        csrf_token: csrf_token
      }
    }, withCookie(cookie))

    if (submitTotpResponse.data.session) {
      const headers = buildHeadersWithCookies(submitTotpResponse)
      return redirect(returnTo, { headers })
    }

    logger.error({ status: submitTotpResponse.status, route: "totp.challange" }, "No session after updateLoginFlow", submitTotpResponse.status)
    return redirect("/totp/challenge")
  } catch (err) {
    const kratosError = err as KratosError
    const flowData = kratosError.response.data
    const flowStatus = kratosError.response.status

    switch (flowStatus) {
      case 400:
        const errorMapping = { form: '' }
        mapFlowToFieldErrors(flowData, errorMapping)
        return data({ errors: { totp_code: errorMapping.form } })

      case 410:
        // Flow expired
        throw redirect("/totp/challenge")

      default:
        logger.error({ status: flowStatus, flowData, route: "totp.challange" }, "Unknown case when updateLoginFlow")
        throw redirect("/totp/challenge")
    }
  }
}
