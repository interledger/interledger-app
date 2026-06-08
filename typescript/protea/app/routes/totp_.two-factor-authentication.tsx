import type { UiNode } from '@ory/client'
import { useRef } from 'react'
import {
  Form,
  data,
  href,
  redirect,
  redirectDocument,
  useActionData,
  useLoaderData,
  useSubmit
} from 'react-router'
import type { ApplicationProps } from '~/components'
import {
  Button,
  Card,
  CardContent,
  GridColumn,
  Layouts,
  OutlineButtonRouter,
  TextField
} from '~/components'
import {
  buildHeadersWithCookies,
  getCookie,
  withCookie
} from '~/lib/kratos/cookie.server'
import {
  handleFlowError,
  mapFlowToFieldErrors
} from '~/lib/kratos/error.server'
import { getCsrfTokenFromFlow } from '~/lib/kratos/flow.server'
import { kratosPublic } from '~/lib/kratos/kratos-client.server'
import type { KratosError } from '~/lib/kratos/types.server'
import logger, { addRequestId } from '~/lib/logger.server'
import { mergeMeta } from '~/lib/meta'
import { extractOrGenerateRequestId } from '~/lib/requestContext.server'
import { useTotpChallenge } from '~/lib/useTotpChallenge'
import type { Route } from './+types/totp_.two-factor-authentication'

type TotpForm = {
  flowId?: string
  qrNode?: string
  secretKey?: string
  totpUnlink?: boolean
  csrfToken?: string
}

export async function loader({ request }: Route.LoaderArgs) {
  const url = new URL(request.url)
  const flowId = url.searchParams.get('flow')
  const cookie = getCookie(request)

  try {
    let flow

    if (flowId) {
      const { data: flowData } = await kratosPublic.getSettingsFlow(
        { id: flowId },
        withCookie(cookie)
      )
      flow = flowData
    } else {
      const response = await kratosPublic.createBrowserSettingsFlow(
        { returnTo: url.searchParams.get('returnTo') ?? undefined },
        withCookie(cookie)
      )
      const headers = buildHeadersWithCookies(response)
      return redirect(
        `/totp/two-factor-authentication?flow=${response.data.id}`,
        { headers }
      )
    }

    const nodes = flow?.ui?.nodes ?? []
    const totpSchema: TotpForm = nodes.reduce(
      (acc: TotpForm, node: UiNode) => {
        if (node.group !== 'totp') return acc
        if ('src' in node.attributes) acc.qrNode ??= node.attributes.src
        if (
          'text' in node.attributes &&
          node.attributes?.id === 'totp_secret_key'
        )
          acc.secretKey ??= node.attributes?.text?.text
        if (
          'name' in node.attributes &&
          node.attributes?.name === 'totp_unlink'
        )
          acc.totpUnlink ??= true

        return acc
      },
      {
        flowId: flow.id,
        qrNode: undefined,
        secretKey: undefined,
        totpUnlink: undefined
      }
    )

    return data({
      ...totpSchema,
      csrfToken: getCsrfTokenFromFlow(flow)
    } as TotpForm)
  } catch (error) {
    handleFlowError(error, 'totp')
    const requestId = extractOrGenerateRequestId(request)
    logger.error(
      { ...addRequestId(requestId), error },
      'Error loading TOTP settings flow'
    )
    return data({ csrfToken: undefined } as TotpForm)
  }
}

export const handle: ApplicationProps = {
  layout: Layouts.Focus,
  scaffold: {
    header: {
      title: 'Two-factor authentication'
    }
  }
}

export const meta = mergeMeta(() => [{ title: 'Two-factor authentication' }])

export default function Page() {
  const { flowId, qrNode, secretKey, totpUnlink, csrfToken } = useLoaderData()
  const actionData = useActionData()
  const formRef = useRef<HTMLFormElement>(null)
  const submit = useSubmit()
  const { withTotpChallenge } = useTotpChallenge()

  if (!flowId && !totpUnlink)
    return (
      <>
        <p>Failed to load flow data.</p>
        <OutlineButtonRouter to={href('/logout')} className='mt-4'>
          Log out
        </OutlineButtonRouter>
      </>
    )

  const handleUnlinkClick = (e: React.MouseEvent<HTMLButtonElement>) => {
    e.preventDefault()
    withTotpChallenge(() => {
      if (formRef.current) {
        submit(formRef.current)
      }
    })
  }

  return (
    <>
      <Form
        ref={formRef}
        id='2fa-form'
        action={`/totp/two-factor-authentication?flow=${flowId}${
          totpUnlink ? '&totpUnlink=true' : ''
        }`}
        method='post'
      >
        <GridColumn className='col-span-full lg:col-span-6'>
          {qrNode && secretKey && (
            <TOTPSetupForm qrNode={qrNode} secretKey={secretKey} />
          )}
          <Card>
            {totpUnlink ? (
              <CardContent>
                <p>Remove TOTP at your own risk</p>
              </CardContent>
            ) : (
              <TextField
                id='totp_code'
                label='Enter verification code from your authenticator app'
                name='totp_code'
                type='number'
                form='2fa-form'
                className='mt-6'
                placeholder='Enter code'
                errorMessage={actionData?.errors.totpCode}
              />
            )}

            <input type='hidden' name='flow' value={flowId} />
            <input type='hidden' name='csrf_token' value={csrfToken} />
          </Card>

          {totpUnlink ? (
            <>
              <Button onClick={handleUnlinkClick}>Unlink TOTP</Button>
              <OutlineButtonRouter to={href('/settings')} className='mt-4'>
                Back
              </OutlineButtonRouter>
            </>
          ) : (
            <Button form='2fa-form' type='submit'>
              Verify TOTP
            </Button>
          )}
          <OutlineButtonRouter to={href('/logout')} className='mt-4'>
            Log out
          </OutlineButtonRouter>
        </GridColumn>
      </Form>
    </>
  )
}

function TOTPSetupForm(totp: { qrNode: string; secretKey: string }) {
  return (
    <Card>
      <CardContent>
        <p>Scan this QR and submit the code</p>
      </CardContent>
      <img
        className='rounded-[1.25rem] p-8'
        alt='Authenticator QR Code'
        src={totp.qrNode}
      />
      <CardContent>
        <p>Or copy code:</p>
        <p>{totp.secretKey}</p>
      </CardContent>
    </Card>
  )
}

export async function action({ request }: Route.ActionArgs) {
  const url = new URL(request.url)
  const totpUnlink = url.searchParams.get('totpUnlink') === 'true'
  const form = await request.formData()
  const flowId = form.get('flow') as string
  const csrfToken = form.get('csrf_token') as string
  const totpCode = form.get('totp_code') as string
  const cookie = getCookie(request)

  if (!flowId) {
    return redirect('/totp/two-factor-authentication')
  }

  const updateBody = totpUnlink
    ? { method: 'totp' as const, totp_unlink: true, csrf_token: csrfToken }
    : { method: 'totp' as const, totp_code: totpCode, csrf_token: csrfToken }

  try {
    const response = await kratosPublic.updateSettingsFlow(
      { flow: flowId, updateSettingsFlowBody: updateBody },
      withCookie(cookie)
    )

    const returnTo = url.searchParams.get('returnTo') || '/'
    const headers = buildHeadersWithCookies(response)
    // Hard reload so the root loader is also run
    return redirectDocument(returnTo, { headers })
  } catch (err) {
    const errResponse = (err as KratosError).response
    const status = errResponse?.status
    const flowData = errResponse?.data

    if (status === 400) {
      const fieldErrors = { form: '', totp_code: '' }
      mapFlowToFieldErrors(flowData, fieldErrors)
      return data({
        errors: {
          totpCode:
            fieldErrors.totp_code ||
            fieldErrors.form ||
            'Invalid code. Please scan the QR code again or add the new code to your authenticator application.'
        }
      })
    }

    handleFlowError(err, 'totp')
    logger.error(
      { error: err, route: 'totp.two-factor-authentication' },
      'Failed to set up TOTP authentication'
    )
    throw new Error('Failed to set up TOTP authentication')
  }
}
