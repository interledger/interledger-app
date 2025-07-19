import type { UiNode } from '@ory/kratos-client'
import type {
  ActionFunctionArgs,
  LoaderFunctionArgs,
  MetaFunction,
  TypedResponse
} from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { Form, useLoaderData } from '@remix-run/react'
import { route } from 'routes-gen'
import type { ApplicationProps } from '~/components'
import {
  Button,
  Card,
  CardContent,
  GridColumn,
  Layouts,
  TextField
} from '~/components'
import { KRATOS_URL, getCsrfTokenFromFlow } from '~/lib/kratos.server'
import { mergeMeta } from '~/lib/meta'

type TotpForm = {
  flowId?: string
  qrNode?: string
  secretKey?: string
  totpUnlink?: boolean
  csrfToken?: string
}

export async function loader({
  request
}: LoaderFunctionArgs): Promise<TypedResponse<TotpForm>> {
  const cookie = String(request.headers.get('cookie') ?? '')
  try {
    const response = await fetch(
      `${KRATOS_URL}/self-service/settings/browser`,
      {
        headers: {
          Accept: 'application/json',
          cookie
        }
      }
    )

    if (response.status === 403) {
      return redirect(
        route('/totp/challenge') +
          '?redirectTo=/totp/two-factor-authentication',
        {
          headers: response.headers
        }
      )
    }

    if (!response.ok) throw new Error('Failed to initiate Kratos settings flow')
    const flow = await response.json()

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

    return json({ ...totpSchema, csrfToken: getCsrfTokenFromFlow(flow) })
  } catch (error) {
    console.error('Error loading settings flow:', error)
    return json({ csrfToken: undefined })
  }
}

export const handle: ApplicationProps = {
  layout: Layouts.Focus,
  scaffold: {
    header: {
      back: route('/'),
      title: 'Two-factor authentication'
    }
  }
}

export const meta: MetaFunction = mergeMeta(() => [
  { title: 'Two-factor authentication' }
])

export default function Page() {
  const { flowId, qrNode, secretKey, totpUnlink, csrfToken } =
    useLoaderData<typeof loader>()

  if (!flowId && !totpUnlink) return <p>Failed to load flow data.</p>

  return (
    <>
      <Form
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
              />
            )}

            <input type='hidden' name='flow' value={flowId} />
            <input type='hidden' name='csrf_token' value={csrfToken} />
          </Card>

          <Button form='2fa-form' type='submit'>
            {totpUnlink ? 'Unlink TOTP' : 'Verify TOTP'}
          </Button>
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

export async function action({ request }: ActionFunctionArgs) {
  const totpUnlink =
    new URL(request.url).searchParams.get('totpUnlink') === 'true'
  const form = await request.formData()
  const flowId = form.get('flow')
  const csrfToken = form.get('csrf_token')
  const totpCode = form.get('totp_code')
  console.log(
    'TOTP Code:',
    totpCode,
    'Flow ID:',
    flowId,
    'CSRF Token:',
    csrfToken,
    'Unlink:',
    totpUnlink
  )
  try {
    const res = await fetch(
      `${KRATOS_URL}/self-service/settings?flow=${flowId}`,
      {
        method: 'POST',
        headers: {
          Accept: 'application/json',
          'Content-type': 'application/json',
          cookie: String(request.headers.get('cookie'))
        },

        body: JSON.stringify(
          !totpUnlink
            ? {
                method: 'totp',
                totp_code: totpCode,
                csrf_token: csrfToken
              }
            : {
                method: 'totp',
                totp_unlink: true,
                csrf_token: csrfToken
              }
        )
      }
    )

    if (res.status === 403) {
      const refresh = await fetch(
        `${KRATOS_URL}/self-service/login/browser?refresh=true&return_to=/'`,
        {
          headers: {
            Accept: 'application/json',
            cookie: String(request.headers.get('cookie'))
          }
        }
      )
      // const flow = await refresh.json()
      console.log('Refresh response:', refresh.status)
      return redirect(route('/totp/two-factor-authentication'), {
        headers: res.headers
        
      })
    }

    if (res.ok) {
      return redirect(route('/'))
    }

    const data = await res.json()
    return json({ flow: data }, { status: res.status })
  } catch (error) {
    throw new Error('Failed to set up TOTP authentication')
  }
}
