import type {
  ActionFunctionArgs,
  LoaderFunctionArgs,
  MetaFunction
} from '@remix-run/node'
import { json } from '@remix-run/node'
import type { ShouldRevalidateFunction } from '@remix-run/react'
import {
  useActionData,
  useLoaderData,
  useSearchParams,
  useSubmit
} from '@remix-run/react'
import { useEffect, useRef, useState } from 'react'
import { route } from 'routes-gen'
import type { ApplicationProps } from '~/components'
import { Button, ButtonRouter, Card, CardContent, Layouts } from '~/components'
import { jsonWithCSRF, validateCSRFToken } from '~/lib/csrf.server'
import { isConnectError } from '~/lib/error.server'
import { grpc } from '~/lib/grpc.server'
import { getClientIP } from '~/lib/ip.server'
import { getUserSession } from '~/lib/kratos.server'
import { mergeMeta } from '~/lib/meta'
import { redirectWithSnackbar } from '~/lib/snackbar.server'
import { PaymentIdentityType } from '~/lib/types/payment'
import { useScaffoldStore } from '~/lib/useScaffoldStore'
import type { ScriptElt } from '~/lib/useScript'
import { useScript } from '~/lib/useScript'

// The loader generates a new 3ds session. This must only be called on initial page load
// and not after submitting actions.
export const shouldRevalidate: ShouldRevalidateFunction = ({
  defaultShouldRevalidate,
  nextUrl
}) => {
  // don't initialise a new 3DS session.
  if (nextUrl.searchParams.has('init')) {
    return false
  }

  return defaultShouldRevalidate
}

export async function loader({ request }: LoaderFunctionArgs) {
  await getUserSession(request)
  const url = new URL(request.url)
  const paymentId = url.searchParams.get('paymentId') as string

  const isInit = url.searchParams.has('init')
  if (isInit) {
    let response = await grpc.init3DS(request, {
      paymentID: paymentId
    })

    if (isConnectError(response)) throw response.errorResponse

    return jsonWithCSRF(request, {
      paymentId,
      quoteId: '',
      initJWT: response.jwt,
      threeDsId: response.id,
      songbirdURL: response.songbirdURL,
      fynbosEnv: process.env.FYNBOS_ENV
    })
  }

  return jsonWithCSRF(request, {
    paymentId,
    quoteId: '',
    initJWT: '',
    threeDsId: '',
    songbirdURL: '',
    fynbosEnv: process.env.FYNBOS_ENV
  })
}

export const handle: ApplicationProps = {
  layout: Layouts.Focus,
  scaffold: {
    header: {
      title: 'Verifying'
    }
  }
}

export const meta: MetaFunction = mergeMeta(() => [
  {
    title: 'Verifying'
  }
])

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
  const loaderData = useLoaderData<typeof loader>()

  const [setLoading] = useScaffoldStore((state) => [state.setLoading])

  useEffect(() => {
    // This ensures that loading is false when this route is unmounted.
    return () => {
      setLoading(false)
    }
  }, [setLoading])

  return loaderData.fynbosEnv === 'local' ? <DevThreeDSPage /> : <ThreeDSPage />
}

function DevThreeDSPage() {
  const loaderData = useLoaderData<typeof loader>()
  const [searchParams, setSearchParams] = useSearchParams()
  const submit = useSubmit()
  useEffect(() => {
    if (searchParams.has('init')) {
      setSearchParams(
        (prev: URLSearchParams) => {
          prev.delete('init')
          return prev
        },
        { replace: true }
      )
    }

    submit(
      {
        name: 'confirm',
        paymentId: loaderData.paymentId,
        csrfToken: loaderData.csrfToken
      },
      {
        action: `${route('/pay/3ds')}?paymentId=${loaderData.paymentId}`,
        method: 'POST',
        replace: true
      }
    )
  }, [
    searchParams,
    setSearchParams,
    loaderData.csrfToken,
    loaderData.paymentId,
    submit
  ])

  return <></>
}

function ThreeDSPage() {
  const loaderData = useLoaderData<typeof loader>()
  const actionData = useActionData<typeof action>()
  const submit = useSubmit()
  let cardinalRef = useRef<any>(null)

  const [setLoading] = useScaffoldStore((state) => [state.setLoading])

  const [showingIssuerChallenge, setShowingIssuerChallenge] =
    useState<boolean>(false)
  const [searchParams, setSearchParams] = useSearchParams()
  const [threeDSError, setThreeDSError] = useState<boolean>(false)
  const [songbirdURL, setSongbirdURL] = useState<string>('')
  const [initJWT, setInitJWT] = useState<string>('')
  const [threeDsId, setThreeDsId] = useState<string>('')
  const [fynbosEnv, setFynbosEnv] = useState<string>('')
  const [paymentId, setPaymentId] = useState<string>('')
  const csrfTokenRef = useRef<string>('')
  const state = useScript(songbirdURL, cleanupSongbirdScript)

  useEffect(() => {
    if (actionData?.challengeURL) setLoading(false)
  }, [actionData?.challengeURL, setLoading])

  useEffect(() => {
    if (loaderData.songbirdURL) {
      setSongbirdURL(loaderData.songbirdURL)
    }
    if (loaderData.initJWT) {
      setInitJWT(loaderData.initJWT)
    }
    if (loaderData.threeDsId) {
      setThreeDsId(loaderData.threeDsId)
    }
    if (loaderData.csrfToken) {
      csrfTokenRef.current = loaderData.csrfToken
    }
    if (loaderData.fynbosEnv) {
      setFynbosEnv(loaderData.fynbosEnv)
    }
    if (loaderData.paymentId) {
      setPaymentId(loaderData.paymentId)
    }
  }, [loaderData])

  useEffect(() => {
    if (
      typeof window !== 'undefined' &&
      state === 'ready' &&
      cardinalRef.current === null &&
      searchParams.has('init')
    ) {
      setLoading(true)
      let actionUrl = `${route('/pay/3ds')}?paymentId=${paymentId}`
      cardinalRef.current = (window as any).Cardinal
      cardinalRef.current.configure({
        logging: {
          level: fynbosEnv === 'prod' ? 'off' : 'on'
        }
      })
      cardinalRef.current.off('payments.setupComplete')
      cardinalRef.current.on('payments.setupComplete', (data: any) => {
        submit(
          {
            name: 'lookup',
            threeDsId,
            colorDepth: window.screen && String(window.screen.colorDepth),
            screenHeight: String(window.innerHeight),
            screenWidth: String(window.innerWidth),
            timezone: String(new Date(Date.now()).getTimezoneOffset()),
            language: navigator.language,
            userAgent: navigator.userAgent,
            paymentId: paymentId,
            csrfToken: csrfTokenRef.current
          },
          {
            action: actionUrl,
            method: 'POST',
            replace: true
          }
        )
      })

      // List of action codes https://cardinaldocs.atlassian.net/wiki/spaces/CC/pages/557065/Songbird.js#Songbird.js-payments.validated
      cardinalRef.current.off('payments.validated')
      cardinalRef.current.on(
        'payments.validated',
        (data: { ActionCode: any }, jwt: string) => {
          switch (data.ActionCode) {
            case 'SUCCESS':
            case 'NOACTION':
              submit(
                {
                  name: 'authenticate',
                  threeDsId,
                  jwt,
                  paymentId,
                  csrfToken: csrfTokenRef.current
                },
                {
                  action: actionUrl,
                  method: 'POST',
                  replace: true
                }
              )
              break

            default:
              setThreeDSError(true)
              setLoading(false)
          }
        }
      )

      cardinalRef.current.setup('init', {
        jwt: initJWT
      })
      // remove the init param so the loader doesn't initialise another 3DS session on subsequent calls.
      setSearchParams(
        (prev: URLSearchParams) => {
          prev.delete('init')
          return prev
        },
        { replace: true }
      )
    }
  }, [
    initJWT,
    state,
    threeDsId,
    submit,
    fynbosEnv,
    searchParams,
    setSearchParams,
    paymentId,
    setLoading
  ])

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
      {!actionData?.challengeURL && !threeDSError && (
        <Card>
          <CardContent>Verifying payment, please wait.</CardContent>
        </Card>
      )}
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

export async function action({ request }: ActionFunctionArgs) {
  const form = await request.formData()
  const formName = form.get('name')
  const threeDSID = form.get('threeDsId') as string

  await validateCSRFToken(request, form)

  const fieldErrors = {
    form: ''
  }

  const data = {
    errors: { ...fieldErrors },
    challengeURL: '',
    payload: '',
    processorTransactionID: ''
  }

  if (formName === 'lookup') {
    const response = await grpc.lookup3DS(request, {
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
    })

    if (isConnectError(response)) {
      return response.error(
        data,
        {},
        {
          message: 'There was an error processing your payment.',
          action: 'Contact support'
        }
      )
    }

    if (response.challengeURL !== '') {
      Object.assign(data, {
        challengeURL: response.challengeURL,
        payload: response.payload,
        processorTransactionID: response.processorTransactionID
      })
      return json(data)
    }
  }

  if (formName === 'authenticate') {
    const response = await grpc.authenticate3DS(request, {
      threeDSID,
      jwt: form.get('jwt') as string
    })

    if (isConnectError(response)) {
      return response.error(
        data,
        {},
        {
          message: 'There was an error processing your payment.',
          action: 'Contact support'
        }
      )
    }
  }

  const response = await grpc.confirmPayment(request, {
    id: form.get('paymentId') as string
  })

  if (isConnectError(response)) {
    return response.error(
      data,
      {},
      {
        message: 'There was an error processing your payment.',
        action: 'Contact support'
      }
    )
  }

  if (response.receiverIdentityType == PaymentIdentityType.Unknown) {
    return redirectWithSnackbar(request, route('/payments/:paymentId/share', {
      paymentId: response.senderTransactionId
    }), {
      message: 'Payment created successfully.',
      icon: 'close'
    })
  }

  return redirectWithSnackbar(request, route('/'), {
    message: 'Payment created successfully.',
    icon: 'close'
  })
}
