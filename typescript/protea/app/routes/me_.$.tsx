import { useState } from 'react'
import type { LoaderArgs, ActionArgs } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { Form, useLoaderData } from '@remix-run/react'
import { route } from 'routes-gen'
import { Button, Card, Icon, Layouts, Router, Snackbar } from '~/components'
import { hasUserSession } from '~/lib/kratos.server'
import {
  getPublicWalletDetails,
  getWalletPaymentPointer
} from '~/lib/wallet.server'
import {
  httpMapping,
  isGrpcError,
  openPaymentsClient,
  StatusError
} from '~/lib/proto.server'
import { flowType, requireFlow, updateFlow } from '~/lib/flows.server'

export async function loader({ request, params }: LoaderArgs) {
  const paymentPointerParam = params['*'] as string

  const response = await openPaymentsClient
    .getPaymentPointer({ url: paymentPointerParam })
    .then((v) => v)
    .catch(StatusError)
  if (isGrpcError(response)) {
    throw json({}, httpMapping(response.code))
  }

  // TODO once conditional auth implemented.
  // const canSendResponse = await openPaymentsClient
  //   .canSendToPaymentPointer(
  //     { paymentPointer: response.response.url },
  //     {
  //       meta: {
  //         cookies: String(request.headers.get('cookie')) || ''
  //       }
  //     }
  //   )
  //   .then((v) => v)
  //   .catch(StatusError)
  //
  // console.log('canSendResponse', canSendResponse)
  // if (isGrpcError(canSendResponse)) {
  //   throw json({}, httpMapping(canSendResponse.code))
  // }

  const paymentPointer = response.response

  const wallet = await getPublicWalletDetails(request, paymentPointer.walletID)

  if (request.headers.get('Content-type') == 'application/json')
    return redirect(paymentPointer.url)

  let editable = false
  const isUser = hasUserSession(request)
  if (isUser) {
    const walletPaymentPointer = await getWalletPaymentPointer(request)
    if (walletPaymentPointer.formatted == paymentPointer.formatted)
      editable = true
  }

  return json({
    editable,
    wallet,
    paymentPointer,
    paymentPointerParam
  })
}

export const handle = {
  layout: Layouts.FocusLayout
}

export default function Page() {
  const { editable, wallet, paymentPointer, paymentPointerParam } =
    useLoaderData<typeof loader>()

  const [snackbarState, setSnackbar] = useState<any>({
    message: 'Payment pointer copied to clipboard.',
    icon: 'close',
    show: false
  })
  const [showSnackbar, setShowSnackbar] = useState<boolean>(false)

  return (
    <Card>
      <h1 className='flex items-center justify-between font-display text-2xl font-medium'>
        <span>{wallet.publicName}</span>
        {editable && (
          <Router to={route('/settings/profile-public')}>
            <Icon>edit</Icon>
          </Router>
        )}
      </h1>
      <button
        type='button'
        onClick={async () => {
          if (typeof navigator.clipboard == 'undefined') {
            setSnackbar({
              message: "Couldn't copy to clipboard.",
              icon: 'close',
              show: true
            })
            setShowSnackbar(true)
          } else
            navigator.clipboard.writeText(paymentPointer.formatted).then(
              () => {
                setSnackbar({
                  message: 'Payment pointer copied to clipboard.',
                  icon: 'close',
                  show: true
                })
                setShowSnackbar(true)
              },
              () => {
                setSnackbar({
                  message: "Couldn't copy to clipboard.",
                  icon: 'close',
                  show: true
                })
                setShowSnackbar(true)
              }
            )
        }}
        className='mt-4 flex flex items-center justify-between rounded-xl bg-nav p-4 hover:bg-nav-hover'
      >
        <span className='font-medium text-medium'>
          {paymentPointer.formatted}
        </span>
        <Icon className='text-medium'>content_copy</Icon>
      </button>

      <Form
        id='me'
        action={`/me/${paymentPointerParam}`}
        method='post'
        className='hidden'
      />
      <input
        form='me'
        value={paymentPointer.formatted}
        name='paymentPointer'
        type='hidden'
      />
      <div className='mt-12'>
        <Button form='me' type='submit'>
          Pay
        </Button>
      </div>
      <Snackbar
        message={snackbarState.message}
        action={snackbarState.action}
        icon={snackbarState.icon}
        show={showSnackbar}
        id='cookie-snackbar'
        dismissAfter={3000}
        onClose={() => setShowSnackbar(false)}
      />
    </Card>
  )
}

export async function action({ request, params }: ActionArgs) {
  const paymentPointerParam = params['*'] as string
  const response = await openPaymentsClient
    .getPaymentPointer({ url: paymentPointerParam })
    .then((v) => v)
    .catch(StatusError)
  if (isGrpcError(response)) {
    throw json({}, httpMapping(response.code))
  }

  // We can redirect immediately because /pay/amount will handle un-authed calls appropriately.
  await requireFlow(request, flowType.Pay, {
    startRoute: route('/pay/amount'),
    data: {
      paymentPointer: { ...response.response }
    },
    returnTo: '/'
  })

  await updateFlow(request, flowType.Pay, {
    paymentPointer: { ...response.response }
  })
  return redirect(route('/pay/amount'))
}
