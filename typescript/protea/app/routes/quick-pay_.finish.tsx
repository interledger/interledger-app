import type {
  ActionFunctionArgs,
  LoaderFunctionArgs,
  MetaFunction
} from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { Form, useActionData, useFetcher, useLoaderData } from '@remix-run/react'
import { useEffect, useState } from 'react'
import {
  type PaymentResultType,
  finishPayment,
  checkOutgoingPayment, 
  getGrantStatus
} from '~/lib/open-payments.server'
import { destroySession, getSession } from '~/session.server'
import { type ApplicationProps, Button, GridColumn, Layouts, WalletGrid } from '~/components'
import { mergeMeta } from '~/lib/meta'
import { QuickPaySession } from '~/lib/types'
//import {isWalletLayout } from '~/lib/utils'\
import { FinishCheck, FinishError } from '~/components/QuickPay'

export async function loader({ request }: LoaderFunctionArgs) {
  const searchParams = new URL(request.url).searchParams
  const paymentId = searchParams.get('paymentId') || ''
  const hash = searchParams.get('hash') || ''
  const interactRef = searchParams.get('interact_ref') || ''
  const result = searchParams.get('result') || ''
  const session = await getSession(request.headers.get('Cookie'))
  const sessionData = session.get('quickPay')
  const isRequestPayment = sessionData?.isRequestPayment
  const currentGrant = sessionData?.grants[paymentId]
  //const isWalletView = await isWalletLayout(request)

  if (!currentGrant) {
    throw json(
      {
        code: "QUICKPAY_SESSION_ERROR",
        title: "Invalid payment grant."
      },
      { status: 400 }
    )
  }

  return json({
    paymentId,
    hash,
    interactRef,
    result,
    isRequestPayment,
    currentGrant
  })
}

export const handle: ApplicationProps = {
  layout: Layouts.Focus,
  scaffold: {
    header: { title: 'Interledger Pay' }
  }
}

export const meta: MetaFunction = mergeMeta(() => [
  {
    title: 'Interledger Pay'
  }
])

export default function Page() {
  const actionData = useActionData<typeof action>()
  const [ loading, setLoading ] = useState(false)
  const [ message, setMessage ] = useState('') 
  const { paymentId, hash, interactRef, result, currentGrant } = useLoaderData<typeof loader>()
  const fetcher = useFetcher()

  const isLoading = fetcher.state !== 'idle'
  const fetcherData = fetcher.data as
    | { success: true }
    | { success: false; message?: string }
    | undefined

  useEffect(() => {
    if (result !== 'grant_rejected') {
      //waitTime is the duration the grant continuations specifies before calling the next step. In this case the continue.
      const waitTime = currentGrant?.continue?.wait ?? 1
      setTimeout(() => {
        const intent = 'checkIncomingPayment'
        fetcher.submit(
          {
            paymentId,
            hash,
            interactRef,
            intent
          },
          { method: 'post' }
        )
      }, waitTime * 1000)

    } else {
      setMessage('Payment was successfully declined.')
    }
  }, [paymentId, hash, interactRef])

  return (
    <WalletGrid>
      <GridColumn className="col-span-full mt-20 mx-auto text-center max-w-md">
        {isLoading && (
          <>
            <div className="animate-spin h-10 w-10 border-b-2 border-current rounded-full mx-auto mb-6" />
            <div className="text-lg">Checking payment...</div>
          </>
        )}

        <Form method="post">
          {!isLoading && fetcherData?.success && (
            <>
              <div className="text-3xl mb-4">Payment successful</div>
              <div className="mb-10">Your payment was completed.</div>

              <Button type="submit" name="intent" value="finish">Home</Button>
            </>
          )}

          {!isLoading && fetcherData && !fetcherData.success && (
            <>
              <div><FinishError /></div>
              <div className="text-3xl mb-4 text-red-600">Payment failed</div>
              <div className="mb-10">
                {fetcherData.message || message}
              </div>

              <Button type="submit" name="intent" value="finish">Home</Button>
            </>
          )}
        </Form>
      </GridColumn>
    </WalletGrid>
  )

  /*return (
    <>
      <div className="flex justify-center items-center flex-col h-full px-5 gap-8">
        <Loader type="large" />
        <Suspense fallback={<Fallback />}>
          <Await
            resolve={data.checkOutgoingPayment}
            errorElement={<FinishError />}
          >
            {(outgoingPaymentCheck) => (
              <>
                {setLoading(false)}
                {outgoingPaymentCheck.error ? (
                  <>
                    <FinishError />
                    <div className="text-destructive uppercase sm:text-2xl font-medium text-center">
                      {outgoingPaymentCheck.message}
                    </div>
                    <Form method="POST">
                      <Button
                        variant="outline"
                        size="sm"
                        className="wmt-formattable-button"
                        type="submit"
                      >
                        'Home'
                      </Button>
                    </Form>
                  </>
                ) : (
                  <>
                    <FinishCheck color={outgoingPaymentCheck.color} />
                    <div
                      className={cx(
                        'uppercase sm:text-2xl font-medium text-center',
                        outgoingPaymentCheck.color === 'red'
                          ? 'text-destructive'
                          : 'text-green-1'
                      )}
                    >
                      {outgoingPaymentCheck.message}
                    </div>
                    <Form method="POST" {...form.props}>
                      <Button
                        variant="outline"
                        size="sm"
                        className="wmt-formattable-button"
                        type="submit"
                      >
                        'Home'
                      </Button>
                    </Form>
                  </>
                )}
              </>
            )}
          </Await>
        </Suspense>
      </div>
    </>
  )*/
}

export async function action({ request }: ActionFunctionArgs) {
  const session = await getSession(request.headers.get('Cookie'))
  const sessionData: QuickPaySession = session.get('quickPay') || {}
  const formData = Object.fromEntries(await request.formData())
  const intent = formData.intent

  if (intent === 'checkIncomingPayment') {
    const interactRef = formData.interactRef as string
    const walletAddressInfo = sessionData?.validWalletAddress
    const paymentId = String(formData?.paymentId) || ''
    const grant = sessionData?.grants[paymentId]
    const quote = sessionData.quote
    const isRequestPayment = sessionData?.isRequestPayment

    if (!quote || !grant || !walletAddressInfo) {
      throw json(
        {
          code: "QUICKPAY_SESSION_ERROR",
          title: "Payment session expired."
        },
        { status: 400 }
      )
    }

    try {
     /* const finishPaymentResponse = await finishPayment(
        grant,
        quote,
        walletAddressInfo,
        interactRef
      )
      console.log({ finishPaymentResponse })
      const result = await checkOutgoingPayment(
        finishPaymentResponse.url,
        finishPaymentResponse.accessToken,
        quote.incomingPaymentGrantToken,
        quote.receiver,
        isRequestPayment
      )*/
      const grantStatus = await getGrantStatus(grant.continue.access_token.value, grant.continue.uri, interactRef)
      console.log({grantStatus})
      const result = null
      return json(result)
    } catch (err) {
      console.log({ err })

      return json({
        success: false,
        message: 'Internal server error'
      })
    }

  }
  /*const submission = session.get('submission')
  if (submission?.value?.walletAddress) {
    delete submission.value.walletAddress
  }
  const path = `/quick-pay`
 
  return redirect(path, {
    headers: { 'Set-Cookie': await destroySession(session) }
  })*/
}
