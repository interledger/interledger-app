import type { Route } from './+types/quick-pay_.pay'
import { data, redirect } from 'react-router'
import { Form, useActionData, useLoaderData, useRouteLoaderData, useNavigation } from 'react-router'
import type { MetaFunction } from 'react-router'
import { useEffect, useState } from 'react'
import { z } from 'zod'
import { type ApplicationProps, Button, GridColumn, Layouts, TextField, WalletGrid } from '~/components'
import { AmountDisplay, QuoteDialog, PayWithInterledgerMark } from '~/components/QuickPay/'
import { mergeMeta } from '~/lib/meta'
import { fetchRequestQuote, getRequestPaymentDetails, getValidWalletAddress, initializePayment } from '~/lib/open-payments.server'
import { requestSchema, formatAmount, formatDate, createError, routeAllowed } from '~/lib/utils.server'
import type { RootLoaderData } from '~/root'
import { commitSession, getSession } from '~/session.server'
import { type ActionData, QuickPaySession } from '~/lib/types'
import { formatError } from '~/lib/helpers'
import logger from '~/lib/logger.server'

export async function loader({ request }: Route.LoaderArgs) {
    routeAllowed('OP_INTPAY_ENABLED')
    const searchParams = new URL(request.url).searchParams
    const isQuote = searchParams.get('quote') || false
    const paymentId = searchParams.get('url') || ''
    let receiver = searchParams.get('receiver') || ''
    let isValidPaymentUrl = true
    let paymentDetails
    let receiverWalletAddress
    let quoteData

    if (isQuote) {
        const session = await getSession(request.headers.get('Cookie'))
        const sessionData = session.get('quickPay')
        const quote = sessionData?.quote
        receiver = sessionData.receiverAddress

        if (quote === undefined) {
            throw data(
                {
                    code: "QUICKPAY_SESSION_ERROR",
                    title: "Payment session expired."
                },
                { status: 400 }
            )
        }
        quoteData = {
            
        }
    }

    try {
        receiverWalletAddress = await getValidWalletAddress(receiver)
        paymentDetails = await getRequestPaymentDetails(paymentId, receiver)
    } catch (err) {
        logger.error({ err }, 'Invalid payment link.')
        isValidPaymentUrl = false
    }

    if (!isValidPaymentUrl || paymentDetails === undefined || paymentDetails.incomingAmount === undefined) {
        throw data(
            {
                code: "QUICKPAY_SESSION_ERROR",
                title: "Invalid payment link."
            },
            { status: 400 }
        )
    }

    const paymentData = {
        amount: formatAmount(paymentDetails.incomingAmount),
        date: formatDate({ date: paymentDetails.createdAt }),
        receiverName: receiverWalletAddress?.publicName ? receiverWalletAddress.publicName : '',
        note: paymentDetails?.metadata?.description ? paymentDetails.metadata.description : '',
        receiverWalletAddress: paymentDetails.walletAddress
    }

    const assetCode = receiverWalletAddress?.assetCode
    return data({
        paymentData,
        quoteData,
        isQuote: isQuote,
        assetCode,
        paymentId
    } as const)
}

export const handle: ApplicationProps = {
    layout: (_match, context) => context?.isUser ? Layouts.Wallet : Layouts.Marketing,
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
    const { assetCode, isQuote, quoteData, paymentData, paymentId } = useLoaderData<typeof loader>()
    const { walletAddress } = useRouteLoaderData('root') as RootLoaderData
    const actionData = useActionData<ActionData>()
    const navigation = useNavigation()
    const isSubmitting = navigation.state === "submitting"
    const [modalOpen, setModalOpen] = useState(false)
    const errors = actionData?.errors

    useEffect(() => {
        setModalOpen(Boolean(isQuote))
    }, [isQuote])

    return (

        <WalletGrid>
            <GridColumn
                className='col-span-full mt-20 mx-auto'
            >
                <AmountDisplay displayAmount={String(paymentData.amount.amount)} assetCode={assetCode} />
                <div className="mx-auto w-full max-w-sm">
                    <Form method="POST">
                        <input
                            type="hidden"
                            name="incomingPaymentUrl"
                            value={paymentId}
                        />
                        <div className="flex flex-col gap-4">
                            <TextField
                                label="Amount requested"
                                defaultValue={paymentData.amount.amountWithCurrency}
                                disabled
                            />
                            <TextField
                                label="Pay into Wallet Address"
                                defaultValue={paymentData.receiverWalletAddress}
                                disabled
                            />
                            <TextField
                                label="Date requested"
                                defaultValue={paymentData.date}
                                disabled
                            />
                            <TextField
                                label="Payment note"
                                name="note"
                                placeholder="Note"
                                errorMessage={formatError(errors?.note)}
                            />
                            <TextField
                                type="text"
                                label="Pay from"
                                placeholder="Wallet address"
                                name="senderAddress"
                                defaultValue={walletAddress ?? ''}
                                errorMessage={formatError(errors?.senderAddress)}
                            />
                            <div className="flex justify-center">
                                <Button
                                    aria-label="pay"
                                    type="submit"
                                    name="intent"
                                    value="pay"
                                    disabled={isSubmitting}
                                >
                                    {isSubmitting ? (
                                        <span className="animate-pulse">Processing...</span>
                                    ) : (
                                        <>
                                            <span className="text-md">Pay with</span>
                                            <PayWithInterledgerMark className="h-8 w-40 mx-2" />
                                        </>
                                    )}
                                </Button>
                            </div>
                        </div>
                    </Form>
                </div>
                <QuoteDialog
                    showDialog={modalOpen}
                    setShowDialog={setModalOpen}
                    receiverName={paymentData.receiverName}
                    receiveAmount={String(paymentData.amount.amount) || ''}
                    debitAmount={String(paymentData.amount.amount) || ''}
                />
            </GridColumn>
        </WalletGrid >
    )
}

export async function action({ request }: Route.ActionArgs) {
    const searchParams = new URL(request.url).searchParams
    const receiver = searchParams.get('receiver') || ''

    const session = await getSession(request.headers.get('Cookie'))
    const sessionData: QuickPaySession = session.get('quickPay') || {}

    const formData = Object.fromEntries(await request.formData())
    const intent = formData.intent
    let path = `/quick-pay/payment?url=${formData.incomingPaymentUrl}&receiver=${receiver}`

    if (intent !== 'pay' && intent !== 'confirm') {
        sessionData.quote = undefined
        session.set('quickPay', sessionData)
        return redirect(`/quick-pay/pay`, {
            headers: { 'Set-Cookie': await commitSession(session) }
        })
    }

    if (intent === 'pay') {
        const result = requestSchema.safeParse(formData)
    
        if (!result.success) {
          const errors = z.treeifyError(result.error).properties
          return data({
            errors
          })
        }
    
        let senderAddress
        try {
          senderAddress = await getValidWalletAddress(result.data.senderAddress)
          sessionData.validWalletAddress = senderAddress
          session.set('quickPay', sessionData)
        } catch (err) {
          return data({ errors: createError("senderAddress", "Your wallet address is not valid.") })
        }
    
        try {
          const quote = await fetchRequestQuote(result.data)
          console.log({quote})
          sessionData.quote = quote
          session.set('quickPay', sessionData)
        } catch (err) {
          logger.error({ err }, 'Error getting quote.')
          return data({ errors: createError("actionError", "An error occurred, please try again.") })
        }
 path += "&quote=true"       
        return redirect(path, {
          headers: { 'Set-Cookie': await commitSession(session) }
        })
      }
}