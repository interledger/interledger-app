import type { ActionArgs, LoaderArgs, MetaFunction } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import type { ShouldRevalidateFunction } from '@remix-run/react'
import { useActionData, useLoaderData, useSubmit } from '@remix-run/react'
import { useEffect, useRef, useState } from 'react'
import { route } from 'routes-gen'
import type { ApplicationProps } from '~/components'
import {
  Button,
  ButtonRouter,
  Card,
  CardContent,
  Layouts,
  LoadingShapes
} from '~/components'
import { Code } from '~/generated/protobuf-ts/google/rpc/code'
import { getSessionWithCSRFToken, validateCSRFToken } from '~/lib/csrf.server'
import { getClientIP } from '~/lib/ip.server'
import { getUserSession } from '~/lib/kratos.server'
import {
  StatusError,
  grpcClient,
  httpMapping,
  isGrpcError,
  openPaymentsClient
} from '~/lib/proto.server'
import { flashSnackbar } from '~/lib/snackbar.server'
import type { ScriptElt } from '~/lib/useScript'
import { useScript } from '~/lib/useScript'
import { commitSession, getSession } from '~/session.server'

// The loader generates a new 3ds session. This must only be called on initial page load
// and not after submitting actions.
export const shouldRevalidate: ShouldRevalidateFunction = ({
  defaultShouldRevalidate,
  formAction,
  formMethod
}) => {
  if (formAction === route('/pay/3ds') && formMethod === 'POST') {
    return false
  }

  return defaultShouldRevalidate
}

export async function loader({ request }: LoaderArgs) {
  await getUserSession(request)
  const url = new URL(request.url)
  const session = await getSessionWithCSRFToken(request)

  const quoteId = url.searchParams.get('quoteId')

  if (!quoteId) throw json({}, httpMapping(Code.INVALID_ARGUMENT))

  let threeDSInit = await grpcClient
    .init3DS(
      {
        quoteID: quoteId
      },
      {
        meta: {
          cookies: String(request.headers.get('cookie'))
        }
      }
    )
    .then((v) => v)
    .catch(StatusError)
  if (isGrpcError(threeDSInit)) throw json({}, httpMapping(threeDSInit.code))

  return json(
    {
      quoteId,
      initJWT: threeDSInit.response.jwt,
      threeDsId: threeDSInit.response.id,
      songbirdURL: threeDSInit.response.songbirdURL,
      fynbosEnv: process.env.FYNBOS_ENV,
      csrfToken: session.get('csrf-token') as string
    },
    { headers: { 'Set-Cookie': await commitSession(session) } }
  )
}

export const handle: ApplicationProps = {
  layout: Layouts.Focus,
  scaffold: {
    header: {
      title: '3DS Verification'
    }
  }
}

export const meta: MetaFunction = () => {
  return {
    title: '3DS Verification'
  }
}

function cleanupSongbirdScript(script: ScriptElt) {
  if (script) {
    script.remove()
  }

  if (typeof window !== 'undefined' && (window as any).Cardinal) {
    ; (window as any).Cardinal.off('payments.setupComplete')
      ; (window as any).Cardinal.off('payments.validated')
    delete (window as any).Cardinal
  }
}

export default function Page() {
  const { quoteId, initJWT, threeDsId, songbirdURL, fynbosEnv, csrfToken } =
    useLoaderData<typeof loader>()
  const actionData = useActionData<typeof action>()
  const submit = useSubmit()
  const state = useScript(songbirdURL, cleanupSongbirdScript)
  let cardinalRef = useRef<any>(null)
  const [showingIssuerChallenge, setShowingIssuerChallenge] =
    useState<boolean>(false)
  const [threeDSError, setThreeDSError] = useState<boolean>(false)

  useEffect(() => {
    if (
      typeof window !== 'undefined' &&
      state === 'ready' &&
      cardinalRef.current === null
    ) {
      cardinalRef.current = (window as any).Cardinal
      cardinalRef.current.configure({
        logging: {
          level: fynbosEnv === 'prod' ? 'off' : 'on'
        }
      })
      cardinalRef.current.off('payments.setupComplete')
      cardinalRef.current.on('payments.setupComplete', (data: any) => {
        let formData = new FormData()
        formData.append('name', 'lookup')
        formData.append('threeDsId', threeDsId)
        formData.append(
          'colorDepth',
          window.screen && String(window.screen.colorDepth)
        )
        formData.append('screenHeight', String(window.innerHeight))
        formData.append('screenWidth', String(window.innerWidth))
        formData.append(
          'timezone',
          String(new Date(Date.now()).getTimezoneOffset())
        )
        formData.append('language', navigator.language)
        formData.append('userAgent', navigator.userAgent)

        formData.append('quoteId', quoteId)
        formData.append('csrfToken', csrfToken)

        submit(formData, {
          action: route('/pay/3ds'),
          method: 'POST',
          replace: true
        })
      })

      // List of action codes https://cardinaldocs.atlassian.net/wiki/spaces/CC/pages/557065/Songbird.js#Songbird.js-payments.validated
      cardinalRef.current.off('payments.validated')
      cardinalRef.current.on(
        'payments.validated',
        (data: { ActionCode: any }, jwt: string) => {
          switch (data.ActionCode) {
            case 'SUCCESS':
            case 'NOACTION':
              let formData = new FormData()
              formData.append('name', 'authenticate')
              formData.append('threeDsId', threeDsId)
              formData.append('jwt', jwt)

              formData.append('quoteId', quoteId)
              formData.append('csrfToken', csrfToken)

              submit(formData, {
                action: route('/pay/3ds'),
                method: 'POST',
                replace: true
              })
              break

            default:
              setThreeDSError(true)
          }
        }
      )

      cardinalRef.current.setup('init', {
        jwt: initJWT
      })
    }
  }, [initJWT, state, threeDsId, submit, fynbosEnv, quoteId, csrfToken])

  const showIssuerChallenge = () => {
    setShowingIssuerChallenge(true)
    if (typeof window !== 'undefined' && cardinalRef.current !== null) {
      cardinalRef.current.continue(
        'cca',
        {
          AcsUrl: actionData?.challengeURL,
          Payload: actionData?.payload
        },
        {
          OrderDetails: {
            TransactionId: actionData?.processorTransactionID
          }
        }
      )
    }
  }

  return (
    <>
      <LoadingShapes />
      {actionData?.challengeURL && !threeDSError && (
        <>
          <Card>
            <CardContent>
              Your card issuer has requested an extra security check.
            </CardContent>
          </Card>

          <Button
            disabled={showingIssuerChallenge}
            onClick={() => {
              showIssuerChallenge()
            }}
          >
            Continue
          </Button>
        </>
      )}
      {threeDSError && (
        <>
          {/* TODO Put this in an error boundary rather */}
          <Card>
            <CardContent>
              There has been an error processing your transaction, please try
              again.
            </CardContent>
          </Card>

          <ButtonRouter to={route('/pay')}>Retry</ButtonRouter>
        </>
      )}
    </>
  )
}

export async function action({ request }: ActionArgs) {
  const form = await request.formData()
  const formName = form.get('name')
  const threeDSID = form.get('threeDsId') as string
  const quoteId = form.get('quoteId') as string
  const csrfToken = form.get('csrfToken') as string
  const err = await validateCSRFToken(request, csrfToken).catch(
    (err: Error) => err
  )
  if (err) {
    throw json(
      {
        action: {
          route: route('/pay/3ds'),
          text: 'Try again'
        }
      },
      { status: 422, statusText: 'Invalid CSRF token.' }
    )
  }

  if (formName === 'lookup') {
    let lookup3DS = await grpcClient
      .lookup3DS(
        {
          threeDSID,
          colorDepth: String(form.get('colorDepth')) || '',
          header: String(request.headers.get('Accept')),
          ipAddress: getClientIP(request),
          javaEnabled: false,
          javascriptEnabled: true,
          language: String(form.get('language')),
          screenHeight: String(form.get('screenHeight')),
          screenWidth: String(form.get('screenWidth')),
          timezone: String(form.get('timezone')),
          userAgent: String(form.get('userAgent'))
        },
        {
          meta: {
            cookies: String(request.headers.get('cookie'))
          }
        }
      )
      .then((v) => v)
      .catch((e) => {
        return StatusError(e)
      })
    if (isGrpcError(lookup3DS)) {
      throw json({}, httpMapping(lookup3DS.code))
    }

    if (lookup3DS.response.challengeURL !== '') {
      return json(
        {
          challengeURL: lookup3DS.response.challengeURL,
          payload: lookup3DS.response.payload,
          processorTransactionID: lookup3DS.response.processorTransactionID
        },
        200
      )
    }
  }

  if (formName === 'authenticate') {
    let auth3DS = await grpcClient
      .authenticate3DS(
        {
          threeDSID,
          jwt: form.get('jwt') as string
        },
        {
          meta: {
            cookies: String(request.headers.get('cookie'))
          }
        }
      )
      .then((v) => v)
      .catch((e) => {
        return StatusError(e)
      })
    if (isGrpcError(auth3DS)) {
      throw json({}, httpMapping(auth3DS.code))
    }
  }

  const clientIpAddress = getClientIP(request)
  let payment = await openPaymentsClient
    .createOutgoingPayment(
      {
        threeDSID,
        quoteID: quoteId,
        ipAddress: clientIpAddress,
        // deprecated but still required by client..
        idempotencyKey: '',
        identity: '',
        identityType: '',
        description: '',
        externalRef: ''
      },
      {
        meta: {
          cookies: String(request.headers.get('cookie')) || ''
        }
      }
    )
    .then((v) => v)
    .catch(StatusError)
  if (isGrpcError(payment)) throw json({}, httpMapping(payment.code))

  return redirect(route('/'), {
    headers: {
      'Set-Cookie': await flashSnackbar(request, {
        message: 'Payment created successfully.',
        icon: 'close'
      })
    }
  })
}
