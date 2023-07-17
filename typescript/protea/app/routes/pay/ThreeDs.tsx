import type { ActionArgs } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import type { ShouldRevalidateFunction } from '@remix-run/react'
import { useLoaderData, useSubmit } from '@remix-run/react'
import { useRef, useState } from 'react'
import { route } from 'routes-gen'
import { ButtonRouter, Card, CardContent, LoadingShapes } from '~/components'
import { exitFlow, flowType, requireFlow } from '~/lib/flows.server'
import { getClientIP } from '~/lib/ip.server'
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
import type { loader } from '~/routes/pay/route'

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

export function ThreeDs() {
  const {
    // initJWT, threeDsId, songbirdURL,
    init3DS,
    fynbosEnv
  } = useLoaderData<typeof loader>()
  // const actionData = useActionData<typeof action>()
  const submit = useSubmit()
  const state = useScript(init3DS?.songbirdURL, (script: ScriptElt) => {
    if (script) {
      script.remove()
    }

    if (typeof window !== 'undefined' && (window as any).Cardinal) {
      ;(window as any).Cardinal.off('payments.setupComplete')
      ;(window as any).Cardinal.off('payments.validated')
      delete (window as any).Cardinal
    }
  })
  let cardinalRef = useRef<any>(null)
  const [showingIssuerChallenge, setShowingIssuerChallenge] =
    useState<boolean>(false)
  const [threeDSError, setThreeDSError] = useState<boolean>(false)

  // useEffect(() => {
  //   if (
  //     typeof window !== 'undefined' &&
  //     state === 'ready' &&
  //     cardinalRef.current === null
  //   ) {
  //     cardinalRef.current = (window as any).Cardinal
  //     cardinalRef.current.configure({
  //       logging: {
  //         level: fynbosEnv === 'prod' ? 'off' : 'on'
  //       }
  //     })
  //     cardinalRef.current.off('payments.setupComplete')
  //     cardinalRef.current.on('payments.setupComplete', (data: any) => {
  //       let formData = new FormData()
  //       formData.append('name', 'lookup')
  //       formData.append('threeDsId', threeDsId)
  //       formData.append(
  //         'colorDepth',
  //         window.screen && String(window.screen.colorDepth)
  //       )
  //       formData.append('screenHeight', String(window.innerHeight))
  //       formData.append('screenWidth', String(window.innerWidth))
  //       formData.append(
  //         'timezone',
  //         String(new Date(Date.now()).getTimezoneOffset())
  //       )
  //       formData.append('language', navigator.language)
  //       formData.append('userAgent', navigator.userAgent)
  //
  //       submit(formData, {
  //         action: route('/pay/3ds'),
  //         method: 'POST'
  //       })
  //     })
  //
  //     // List of action codes https://cardinaldocs.atlassian.net/wiki/spaces/CC/pages/557065/Songbird.js#Songbird.js-payments.validated
  //     cardinalRef.current.off('payments.validated')
  //     cardinalRef.current.on(
  //       'payments.validated',
  //       (data: { ActionCode: any }, jwt: string) => {
  //         switch (data.ActionCode) {
  //           case 'SUCCESS':
  //           case 'NOACTION':
  //             let formData = new FormData()
  //             formData.append('name', 'authenticate')
  //             formData.append('threeDsId', threeDsId)
  //             formData.append('jwt', jwt)
  //
  //             submit(formData, {
  //               action: route('/pay/3ds'),
  //               method: 'POST'
  //             })
  //             break
  //
  //           default:
  //             setThreeDSError(true)
  //         }
  //       }
  //     )
  //
  //     cardinalRef.current.setup('init', {
  //       jwt: initJWT
  //     })
  //   }
  // }, [initJWT, state, threeDsId, submit, fynbosEnv])

  // const showIssuerChallenge = () => {
  //   setShowingIssuerChallenge(true)
  //   if (typeof window !== 'undefined' && cardinalRef.current !== null) {
  //     cardinalRef.current.continue(
  //       'cca',
  //       {
  //         AcsUrl: actionData?.challengeURL,
  //         Payload: actionData?.payload
  //       },
  //       {
  //         OrderDetails: {
  //           TransactionId: actionData?.processorTransactionID
  //         }
  //       }
  //     )
  //   }
  // }

  return (
    <>
      <LoadingShapes />
      {/*{actionData?.challengeURL && !threeDSError && (*/}
      {/*  <>*/}
      {/*    <Card>*/}
      {/*      <CardContent>*/}
      {/*        Your card issuer has requested an extra security check.*/}
      {/*      </CardContent>*/}
      {/*    </Card>*/}

      {/*    <Button*/}
      {/*      disabled={showingIssuerChallenge}*/}
      {/*      onClick={() => {*/}
      {/*        showIssuerChallenge()*/}
      {/*      }}*/}
      {/*    >*/}
      {/*      Continue*/}
      {/*    </Button>*/}
      {/*  </>*/}
      {/*)}*/}
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
  const flow = await requireFlow(request, flowType.Pay)
  const idempotencyKey = flow.data.idempotencyKey as string

  if (formName === 'lookup') {
    let lookup3DS = await grpcClient
      .lookup3DS(
        {
          idempotencyKey,
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
          idempotencyKey,
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
        idempotencyKey,
        threeDSID,
        quoteID: flow.data.quoteID,
        description: flow.data.note,
        externalRef: '',
        ipAddress: clientIpAddress,
        identityType: flow.data.address.type,
        identity: flow.data.address.handle
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
  return redirect(route('/'), {
    headers: {
      'Set-Cookie': await flashSnackbar(request, {
        message: 'Payment created successfully.',
        icon: 'close'
      })
    }
  })
}
