import type { Route } from './+types/quick-pay_.finish'
import { data, redirect } from 'react-router'
import { Form, useFetcher, useLoaderData } from 'react-router'
import type { MetaFunction } from 'react-router'
import { useEffect, useState } from 'react'
import { finishPayment, checkOutgoingPayment } from '~/lib/open-payments.server'
import { commitSession, getSession } from '~/session.server'
import { type ApplicationProps, Button, GridColumn, Layouts, WalletGrid } from '~/components'
import { mergeMeta } from '~/lib/meta'
import { BackButton, FinishCheck, FinishError } from '~/components/QuickPay'
import logger from '~/lib/logger.server'
import { routeAllowed } from '~/lib/utils.server'
import { QuickPaySession } from '~/lib/types'

export type FinishActionData = {
  message?: string
  error?: boolean
}

export async function loader({ request }: Route.LoaderArgs) {
  routeAllowed('OP_INTPAY_ENABLED')
  const searchParams = new URL(request.url).searchParams
  const paymentId = searchParams.get('paymentId') || ''
  const hash = searchParams.get('hash') || ''
  const interactRef = searchParams.get('interact_ref') || ''
  const result = searchParams.get('result') || ''
  const session = await getSession(request.headers.get('Cookie'))
  const sessionData = session.get('quickPay')
  const isRequestPayment = sessionData?.isRequestPayment
  const currentGrant = sessionData?.grants[paymentId]

  if (!currentGrant) {
    throw data(
      {
        code: "QUICKPAY_SESSION_ERROR",
        title: "Invalid payment grant."
      },
      { status: 400 }
    )
  }

  return data({
    paymentId,
    hash,
    interactRef,
    result,
    isRequestPayment,
    currentGrant
  })
}

export const handle: ApplicationProps = {
  layout: (_match, context) => context?.isUser ? Layouts.Wallet : Layouts.Marketing,
  scaffold: {
    header: { title: 'Interledger Pay' }
  }
}

export const meta: MetaFunction = mergeMeta(() => [
  {
    title: 'Interledger Pay - Final step: Confirmation'
  }
])

export default function Page() {
  const { paymentId, hash, interactRef, isRequestPayment, result, currentGrant } = useLoaderData<typeof loader>()
  const fetcher = useFetcher()
  const fetcherData = fetcher.data as unknown as FinishActionData
  const [loading, setLoading] = useState(true)
  const [statusAndMessage, setStatusAndMessage] = useState({ error: false, message: '' })
  const clearSessionKeys: Array<keyof QuickPaySession> = ['receiverAddress', 'quote']

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
            <div className="fixed inset-0 z-[60] flex w-full h-full bg-page">
            </div>
            <div className="z-[70]">
              <div className="animate-spin h-10 w-10 border-b-2 border-current rounded-full mx-auto mb-6" />
              <div className="text-lg">Checking payment...</div>
            </div>
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
                <BackButton title="Back" to="/quick-pay/pay"/>
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

export async function action({ request }: Route.ActionArgs) {
  const session = await getSession(request.headers.get('Cookie'))
  let sessionData: QuickPaySession = session.get('quickPay') || {}
  const formData = Object.fromEntries(await request.formData())
  const intent = formData.intent

  if (intent === 'checkIncomingPayment') {
    const interactRef = formData.interactRef as string
    const walletAddressInfo = sessionData?.senderAddress
    const paymentId = String(formData?.paymentId) || ''
    const grants = sessionData?.grants || {}
    const grant = grants[paymentId]
    const quote = sessionData.quote
    const isRequestPayment = !!sessionData?.request

    if (!quote || !grant || !walletAddressInfo) {
      throw data(
        {
          code: "QUICKPAY_SESSION_ERROR",
          title: "Payment session expired."
        },
        { status: 400 }
      )
    }

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
        String(quote.incomingPaymentGrantToken),
        quote.receiver,
        isRequestPayment
      )
      return data(result)
    } catch (err) {
      logger.error({ err }, 'Open payments response error.')

      return data({
        error: true,
        message: 'Internal server error'
      })
    }

  }

  if (intent === 'finish') {
    sessionData = {}
    session.set('quickPay', sessionData)
    return redirect('/quick-pay', {
      headers: { 'Set-Cookie': await commitSession(session) }
    })
  }

}
