import type { ActionArgs, LoaderArgs, MetaFunction } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
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
import { exitFlow, flowType, requireFlow } from '~/lib/flows.server'
import { getClientIP } from '~/lib/ip.server'
import { getUserSession } from '~/lib/kratos.server'
import {
  StatusError,
  grpcClient,
  httpMapping,
  isGrpcError,
  openPaymentsClient
} from '~/lib/proto.server'
import { useScript } from '~/lib/useScript'

export async function loader({ request }: LoaderArgs) {
  await getUserSession(request)
  const flow = await requireFlow(request, flowType.Pay)

  let threeDSInit = await grpcClient
    .init3DS(
      {
        idempotencyKey: flow.data.idempotencyKey,
        quoteID: flow.data.quoteID
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

  return json({
    flow,
    jwt: threeDSInit.response.jwt,
    threeDsId: threeDSInit.response.id,
    songbirdURL: threeDSInit.response.songbirdURL
  })
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

export default function Page() {
  const {
    flow,
    jwt: initJWT,
    threeDsId,
    songbirdURL
  } = useLoaderData<typeof loader>()
  const actionData = useActionData<typeof action>()
  const submit = useSubmit()
  const state = useScript(songbirdURL)
  let cardinalRef = useRef<any>(null)
  const [showingIssuerChallenge, setShowingIssuerChallenge] =
    useState<boolean>(false)
  const [threeDSError, setThreeDSError] = useState<boolean>(false)

  useEffect(() => {
    if (
      typeof window !== 'undefined' &&
      state == 'ready' &&
      cardinalRef.current === null
    ) {
      cardinalRef.current = (window as any).Cardinal
      cardinalRef.current.configure({
        logging: {
          level: 'on'
        }
      })
      cardinalRef.current.setup('init', {
        jwt: initJWT
      })
      cardinalRef.current.on('payments.setupComplete', (data: any) => {
        let formData = new FormData()
        formData.append('name', 'lookup')
        formData.append('threeDsId', threeDsId)
        formData.append('idempotencyKey', flow?.data?.idempotencyKey)
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

        submit(formData, {
          action: '/pay/3ds',
          method: 'post'
        })
      })

      // List of action codes https://cardinaldocs.atlassian.net/wiki/spaces/CC/pages/557065/Songbird.js#Songbird.js-payments.validated
      cardinalRef.current.on(
        'payments.validated',
        (data: { ActionCode: any }, jwt: string) => {
          switch (data.ActionCode) {
            case 'SUCCESS':
            case 'NOACTION':
              let formData = new FormData()
              formData.append('name', 'authenticate')
              formData.append('threeDsId', threeDsId)
              formData.append('idempotencyKey', flow?.data?.idempotencyKey)
              formData.append('jwt', jwt)

              submit(formData, {
                action: '/pay/3ds',
                method: 'post'
              })
              break

            default:
              setThreeDSError(true)
          }
        }
      )
    }
  }, [initJWT, state, threeDsId, flow?.data?.idempotencyKey, submit])

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
  const formName = await form.get('name')
  const idempotencyKey = await form.get('idempotencyKey')
  const threeDsId = await form.get('threeDsId')

  if (formName === 'lookup') {
    let lookup3DS = await grpcClient
      .lookup3DS(
        {
          idempotencyKey: String(idempotencyKey),
          threeDSID: String(threeDsId),
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
    const jwt = await form.get('jwt')
    let auth3DS = await grpcClient
      .authenticate3DS(
        {
          idempotencyKey: String(idempotencyKey),
          threeDSID: String(threeDsId),
          jwt: String(jwt)
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

  const flow = await requireFlow(request, flowType.Pay)
  const clientIpAddress = getClientIP(request)
  let payment = await openPaymentsClient
    .createOutgoingPayment(
      {
        idempotencyKey: flow.data.idempotencyKey || '',
        quoteID: flow.data.quoteID,
        description: flow.data.note,
        externalRef: '',
        ipAddress: clientIpAddress,
        threeDSID: String(threeDsId)
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

  await exitFlow(request, flowType.Pay)
  return redirect(route('/'))
}
