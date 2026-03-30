import type {
  ActionFunctionArgs,
  LoaderFunctionArgs,
  MetaFunction
} from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { Form, useFetcher, useLoaderData } from '@remix-run/react'
import { useEffect, useState } from 'react'
import { finishPayment, checkOutgoingPayment } from '~/lib/open-payments.server'
import { commitSession, getSession } from '~/session.server'
import { type ApplicationProps, Button, GridColumn, Layouts, WalletGrid } from '~/components'
import { mergeMeta } from '~/lib/meta'
import { QuickPaySession } from '~/lib/types'
//import {isWalletLayout } from '~/lib/utils'\
import { FinishCheck, FinishError } from '~/components/QuickPay'

export type FinishActionData = {
  message?: string
  error?: boolean
}

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
  const { paymentId, hash, interactRef, isRequestPayment, result, currentGrant } = useLoaderData<typeof loader>()
  const fetcher = useFetcher()
  const fetcherData = fetcher.data as unknown as FinishActionData
  const [loading, setLoading] = useState(true)
  const [statusAndMessage, setStatusAndMessage] = useState({ error: false, message: '' })

  useEffect(() => {
    if (fetcherData && fetcherData.message) {
      setStatusAndMessage({ error: fetcherData.error || false, message: fetcherData.message })
      setLoading(false)
    }
  }, [fetcherData])

  useEffect(() => {
    if (result === 'grant_rejected') {
      setStatusAndMessage({ error: true, message: 'Payment was successfully declined.' })
      setLoading(false)
    } else if (!fetcherData) {
      //waitTime is the duration the grant continuations specifies before calling the next step. In this case the continue.
      const waitTime = currentGrant?.continue?.wait ?? 1
      const timer = setTimeout(() => {
        const intent = 'checkIncomingPayment'
        setLoading(true)
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
      return () => clearTimeout(timer)
    }
  }, [paymentId, hash, interactRef, currentGrant])

  return (
    <WalletGrid>
      <GridColumn className="col-span-full mt-20 mx-auto text-center max-w-md">
        {loading ? (
          <>
            <div className="animate-spin h-10 w-10 border-b-2 border-current rounded-full mx-auto mb-6" />
            <div className="text-lg">Checking payment...</div>
          </>
        ) : (
          <Form method="post">
            {!statusAndMessage.error ? (
              <>
                <div className="flex justify-center mb-6"><FinishCheck className="w-16 h-16" /></div>
                <div className="text-3xl mb-4">Payment successful</div>
                <div className="mb-10">Your payment was completed.</div>

                <Button type="submit" name="intent" value="finish">Home</Button>
              </>
            ) : (
              <>
                <div className="flex justify-center mb-6"><FinishError className="w-16 h-16" /></div>
                <div className="text-3xl mb-4 text-red-600">Payment failed</div>
                <div className="mb-10">
                  {statusAndMessage.message}
                </div>

                <Button type="submit" name="intent" value="finish">Home</Button>
              </>
            )}
          </Form>)}
      </GridColumn>
    </WalletGrid>
  )
}

export async function action({ request }: ActionFunctionArgs) {
  const session = await getSession(request.headers.get('Cookie'))
  let sessionData: QuickPaySession = session.get('quickPay') || {}
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

    //Error is generatede on next line
    try {
      const finishPaymentResponse = await finishPayment(
        grant,
        quote,
        walletAddressInfo,
        interactRef
      )
      const result = await checkOutgoingPayment(
        finishPaymentResponse.url,
        finishPaymentResponse.accessToken,
        quote.incomingPaymentGrantToken,
        quote.receiver,
        isRequestPayment
      )
      console.log({ result })
      return json(result)
    } catch (err) {
      console.log({ err })

      return json({
        error: true,
        message: 'Internal server error'
      })
    }

  }

  if (intent === 'finish') {
    //Reset session info
    sessionData = {}
    session.set('quickPay', sessionData)
    return redirect('/quick-pay', {
      headers: { 'Set-Cookie': await commitSession(session) }
    })
  }

}
