import { useState } from 'react'
import type { LoaderArgs, ActionArgs } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { Form, useLoaderData } from '@remix-run/react'
import { route } from 'routes-gen'
import { Button, HomeShapes, Icon, Layouts, Snackbar } from '~/components'
import { hasUserSession } from '~/lib/kratos.server'
import { getWalletPaymentPointer } from '~/lib/wallet.server'
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
  const paymentPointer = response.response

  if (request.headers.get('Content-type') == 'application/json')
    return redirect(paymentPointer.url)

  let editable = false
  const isUser = await hasUserSession(request)
  if (isUser) {
    const walletPaymentPointer = await getWalletPaymentPointer(request)
    if (walletPaymentPointer.formatted == paymentPointer.formatted)
      editable = true
  }

  return json({
    editable,
    paymentPointer,
    paymentPointerParam
  })
}

export const handle = {
  layout: Layouts.FocusLayout
}

export default function Page() {
  const { paymentPointer, paymentPointerParam } = useLoaderData<typeof loader>()

  const [snackbarState, setSnackbar] = useState<any>({
    message: 'Payment pointer copied to clipboard.',
    icon: 'close',
    show: false
  })
  const [showSnackbar, setShowSnackbar] = useState<boolean>(false)

  return (
    <div className='flex w-full flex-col rounded-2xl bg-page p-4 pb-8'>
      <div className='mt-2'>
        <HomeShapes />
      </div>
      <h1 className='mt-6 flex items-center justify-between font-display text-2xl font-medium'>
        <span>{paymentPointer.alias}</span>
        {/*TODO: Edit button once we can edit the alias */}
        {/*{!editable && <IconButton>edit</IconButton>}*/}
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
        className='mt-4 flex flex items-center justify-between rounded-xl bg-container p-4 hover:bg-container-hover'
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
    </div>
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

  await requireFlow(request, flowType.Pay, {
    startRoute: route('/pay'),
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
