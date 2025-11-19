import type { UiNode } from '@ory/kratos-client'
import type {
  ActionFunctionArgs,
  LoaderFunctionArgs,
  MetaFunction,
  TypedResponse
} from '@remix-run/node'
import { json, redirectDocument } from '@remix-run/node'
import { Form, useActionData, useLoaderData, useSubmit } from '@remix-run/react'
import { useRef } from 'react'
import { route } from 'routes-gen'
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
import { KRATOS_URL, getCsrfTokenFromFlow } from '~/lib/kratos.server'
import { mergeMeta } from '~/lib/meta'
import { useTotpChallenge } from '~/lib/useTotpChallenge'

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
      back: route('/'), // THIS IS AN ISSUE, SHOULD BE A REDIRECT
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
  const actionData = useActionData<typeof action>()
  const formRef = useRef<HTMLFormElement>(null)
  const submit = useSubmit()
  const { withTotpChallenge } = useTotpChallenge()

  if (!flowId && !totpUnlink)
    return (
      <>
        <p>Failed to load flow data.</p>
        <OutlineButtonRouter to={route('/logout')} className='mt-4'>
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
            <Button onClick={handleUnlinkClick}>Unlink TOTP</Button>
          ) : (
            <Button form='2fa-form' type='submit'>
              Verify TOTP
            </Button>
          )}
          <OutlineButtonRouter to={route('/logout')} className='mt-4'>
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

export async function action({ request }: ActionFunctionArgs) {
  const totpUnlink =
    new URL(request.url).searchParams.get('totpUnlink') === 'true'
  const form = await request.formData()
  const flowId = form.get('flow')
  const csrfToken = form.get('csrf_token')
  const totpCode = form.get('totp_code')
  const errors = {
    totpCode: ''
  }

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

    // if (res.ok && totpUnlink) {
    //   return redirect(route('/logout'))
    // }

    if (res.ok) {
      const returnTo = new URL(request.url).searchParams.get('returnTo') || '/'
      const response = redirectDocument(returnTo ?? '/')
      // Hard reload so the root loader is also run
      return response
    }

    if (res.status === 400) {
      errors.totpCode =
        'Invalid code. Please scan the QR code again or add the new code to your authenticator application.'
    }

    return json({ errors }, { status: res.status })
  } catch (error) {
    throw new Error('Failed to set up TOTP authentication')
  }
}
