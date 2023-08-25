import type { ActionArgs, LoaderArgs, MetaFunction } from '@remix-run/node'
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
import {
  Button,
  ButtonRouter,
  Card,
  CardContent,
  Layouts,
  LoadingShapes
} from '~/components'
import { Code } from '~/generated/protobuf-ts/google/rpc/code'
import { jsonWithCSRF, validateCSRFToken } from '~/lib/csrf.server'
import { error } from '~/lib/error.server'
import { getClientIP } from '~/lib/ip.server'
import { getUserSession } from '~/lib/kratos.server'
import {
  StatusError,
  grpcClient,
  httpMapping,
  isGrpcError,
  openPaymentsClient
} from '~/lib/proto.server'
import { redirectWithSnackbar } from '~/lib/snackbar.server'
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

export async function loader(args: LoaderArgs) {
  await getUserSession(args.request)
  const url = new URL(args.request.url)
  const quoteId = url.searchParams.get('quoteId')

  if (quoteId) {
    return openPayments3DSLoader(args)
  }

  return paymentsEngine3DSLoader(args)  
}

async function openPayments3DSLoader({ request }: LoaderArgs) {
  const url = new URL(request.url)

  const quoteId = url.searchParams.get('quoteId')

  if (!quoteId) throw json({}, httpMapping(Code.INVALID_ARGUMENT))

  const isInit = url.searchParams.has('init')
  if (isInit) {
    let threeDSInit = await grpcClient
      .initQuote3DS(
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

    return jsonWithCSRF(request, {
      quoteId,
      initJWT: threeDSInit.response.jwt,
      threeDsId: threeDSInit.response.id,
      songbirdURL: threeDSInit.response.songbirdURL,
      fynbosEnv: process.env.FYNBOS_ENV,
      paymentId: ''
    })
  }

  return jsonWithCSRF(request, {
    quoteId,
    initJWT: '',
    threeDsId: '',
    songbirdURL: '',
    fynbosEnv: process.env.FYNBOS_ENV,
    paymentId: ''
  })
}

async function paymentsEngine3DSLoader({ request }: LoaderArgs) {
  const url = new URL(request.url)

  const paymentId = url.searchParams.get('paymentId')

  if (!paymentId) throw json({}, httpMapping(Code.INVALID_ARGUMENT))

  const isInit = url.searchParams.has('init')
  if (isInit) {
    let threeDSInit = await grpcClient
      .init3DS(
        {
          paymentID: paymentId,
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

    return jsonWithCSRF(request, {
      paymentId,
      quoteId: '',
      initJWT: threeDSInit.response.jwt,
      threeDsId: threeDSInit.response.id,
      songbirdURL: threeDSInit.response.songbirdURL,
      fynbosEnv: process.env.FYNBOS_ENV,
    })
  }

  return jsonWithCSRF(request, {
    paymentId,
    quoteId: '',
    initJWT: '',
    threeDsId: '',
    songbirdURL: '',
    fynbosEnv: process.env.FYNBOS_ENV,
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
  const [songbirdURL, setSongbirdURL] = useState<string>('')
  const [initJWT, setInitJWT] = useState<string>('')
  const [threeDsId, setThreeDsId] = useState<string>('')
  const [fynbosEnv, setFynbosEnv] = useState<string>('')
  const [quoteId, setQuoteId] = useState<string>('')
  const [paymentId, setPaymentId] = useState<string>('')
  const csrfTokenRef = useRef<string>('')
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
    if (loaderData.quoteId) {
      setQuoteId(loaderData.quoteId)
    }
    if (loaderData.paymentId) {
      setPaymentId(loaderData.paymentId)
    }
  }, [loaderData])

  return <>
    <LoadingShapes />
    {
      loaderData.fynbosEnv === 'local' ? <DevThreeDSPage {...loaderData} /> : <ThreeDSPage {...loaderData} />
    }
  </>
}

function DevThreeDSPage(args: threeDsProps) {  
  const [searchParams, setSearchParams] = useSearchParams()
  const submit = useSubmit()
  useEffect(() => {
    if (searchParams.has('init')) {
      setSearchParams((prev: URLSearchParams) => {
        prev.delete('init')
        return prev
      })
    }

    submit(
      {
        name: 'confirm',
        paymentId: args.paymentId,
        csrfToken: args.csrfToken
      },
      {
        action: `${route('/pay/3ds')}?paymentId=${args.paymentId}`,
        method: 'POST',
        replace: true
      }
    )
  }, [searchParams])

  return <>
  </>
}

type threeDsProps = {
  quoteId: string
  paymentId: string
  csrfToken: string
  initJWT: string
  songbirdURL: string
  threeDsId: string
  fynbosEnv?: string
}

function ThreeDSPage(args: threeDsProps) {
  const actionData = useActionData<typeof action>()
  const submit = useSubmit()
  let cardinalRef = useRef<any>(null)
  const [showingIssuerChallenge, setShowingIssuerChallenge] =
    useState<boolean>(false)
  const [searchParams, setSearchParams] = useSearchParams()
  const [threeDSError, setThreeDSError] = useState<boolean>(false)
  const state = useScript(args.songbirdURL, cleanupSongbirdScript)

  useEffect(() => {
    if (
      typeof window !== 'undefined' &&
      state === 'ready' &&
      cardinalRef.current === null &&
      searchParams.has('init')
    ) {
      let actionUrl = `${route('/pay/3ds')}?quoteId=${args.quoteId}`
      if (args.paymentId) {
        actionUrl = `${route('/pay/3ds')}?paymentId=${args.paymentId}`
      } 

      cardinalRef.current = (window as any).Cardinal
      cardinalRef.current.configure({
        logging: {
          level: args.fynbosEnv === 'prod' ? 'off' : 'on'
        }
      })
      cardinalRef.current.off('payments.setupComplete')
      cardinalRef.current.on('payments.setupComplete', (data: any) => {
        submit(
          {
            name: 'lookup',
            threeDsId: args.threeDsId,
            colorDepth: window.screen && String(window.screen.colorDepth),
            screenHeight: String(window.innerHeight),
            screenWidth: String(window.innerWidth),
            timezone: String(new Date(Date.now()).getTimezoneOffset()),
            language: navigator.language,
            userAgent: navigator.userAgent,
            quoteId: args.quoteId,
            paymentId: args.paymentId,
            csrfToken: args.csrfToken
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
                  threeDsId: args.threeDsId,
                  jwt,
                  quoteId: args.quoteId,
                  paymentId: args.paymentId,
                  csrfToken: args.csrfToken
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
          }
        }
      )

      cardinalRef.current.setup('init', {
        jwt: args.initJWT
      })
      // remove the init param so the loader doesn't initialise another 3DS session on subsequent calls.
      setSearchParams((prev: URLSearchParams) => {
        prev.delete('init')
        return prev
      })
    }
  }, [
    state,
    submit,
    args,
    searchParams,
    setSearchParams
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
      .catch(StatusError)
    if (isGrpcError(lookup3DS)) {
      return error(request, data, {
        message: 'There was an error processing your payment.',
        action: 'Contact support'
      })
    }

    if (lookup3DS.response.challengeURL !== '') {
      Object.assign(data, {
        challengeURL: lookup3DS.response.challengeURL,
        payload: lookup3DS.response.payload,
        processorTransactionID: lookup3DS.response.processorTransactionID
      })
      return json(data)
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
      .catch(StatusError)
    if (isGrpcError(auth3DS)) {
      return error(request, data, {
        message: 'There was an error processing your payment.',
        action: 'Contact support'
      })
    }
  }

  const quoteId = form.get('quoteId') as string
  if (quoteId) {
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
    if (isGrpcError(payment)) {
      return error(request, data, {
        message: 'There was an error processing your payment.',
        action: 'Contact support'
      })
    }
  } else {
    let payment = await grpcClient
      .confirmPayment(
        {
          id: form.get('paymentId') as string
        },
        {
          meta: {
            cookies: String(request.headers.get('cookie')) || ''
          }
        }
      )
      .then((v) => v)
      .catch(StatusError)
    if (isGrpcError(payment)) {
      return error(request, data, {
        message: 'There was an error processing your payment.',
        action: 'Contact support'
      })
    }
  }

  return redirectWithSnackbar(request, route('/'), {
    message: 'Payment created successfully.',
    icon: 'close'
  })
}
