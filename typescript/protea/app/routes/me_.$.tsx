import type { ActionArgs, LoaderArgs } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { Form, useLoaderData } from '@remix-run/react'
import { useState } from 'react'
import { route } from 'routes-gen'
import {
  Button,
  Card,
  Chip,
  ChipColor,
  FynbosIcon,
  Icon,
  Layouts,
  Router,
  Snackbar,
  TwitterIcon
} from '~/components'
import { flowType, requireFlow, updateFlow } from '~/lib/flows.server'
import { hasUserSession } from '~/lib/kratos.server'
import {
  StatusError,
  httpMapping,
  isGrpcError,
  openPaymentsClient
} from '~/lib/proto.server'
import {
  getPublicLinkedIdentities,
  getPublicWalletDetails,
  getWalletPaymentPointer
} from '~/lib/wallet.server'

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

  if (request.headers.get('Content-type') == 'application/json')
    return redirect(paymentPointer.url)

  const wallet = await getPublicWalletDetails(request, paymentPointer.walletID)
  const identities = await getPublicLinkedIdentities(
    request,
    paymentPointer.walletID
  )

  let editable = false
  const isUser = hasUserSession(request)
  if (isUser) {
    const walletPaymentPointer = await getWalletPaymentPointer(request)
    if (walletPaymentPointer.formatted == paymentPointer.formatted)
      editable = true
  }

  return json({
    isUser,
    editable,
    wallet,
    identities,
    paymentPointer,
    paymentPointerParam
  })
}

export const handle = {
  layout: Layouts.Focus
}

export default function Page() {
  const {
    isUser,
    editable,
    wallet,
    identities,
    paymentPointer,
    paymentPointerParam
  } = useLoaderData<typeof loader>()

  const [snackbarState, setSnackbar] = useState<any>({
    message: 'Payment pointer copied to clipboard.',
    icon: 'close',
    show: false
  })
  const [showSnackbar, setShowSnackbar] = useState<boolean>(false)

  return (
    <>
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
          <span className='text-medium'>{paymentPointer.formatted}</span>
          <Icon className='text-medium'>content_copy</Icon>
        </button>
        {identities.map((identity) => (
          <Router
            key={identity.id}
            className='mt-2 flex items-center justify-between rounded-xl bg-nav p-3 text-medium first-of-type:mt-6 hover:bg-nav-hover'
            to={route('/me/identities/:identityId', {
              identityId: identity.signatureHash
            })}
          >
            <div className='flex space-x-3'>
              <TwitterIcon />
              <span>@{identity.identifier}</span>
            </div>
            <div className='flex space-x-3'>
              {identity.state == 'verified' && (
                <Chip color={ChipColor.green}>Verified</Chip>
              )}
              <Icon>navigate_next</Icon>
            </div>
          </Router>
        ))}
      </Card>
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
      <Button form='me' type='submit'>
        Send a payment
      </Button>
      {!isUser && (
        <Card className='space-y-4'>
          <h1 className='font-display text-lg font-medium'>Sign up</h1>
          <div className='flex items-start space-x-4'>
            <div className='flex items-center justify-between rounded-full bg-nav p-5 text-medium'>
              <FynbosIcon />
            </div>
            <div className='flex flex-col space-y-4'>
              <p className='text-sm text-medium'>
                Sign up with Fynbos to reserve your wallet address and start
                transacting.
              </p>
              <Router
                className='text-sm font-medium text-primary'
                to={route('/signup')}
              >
                Sign up now
              </Router>
            </div>
          </div>
        </Card>
      )}
      <Snackbar
        message={snackbarState.message}
        action={snackbarState.action}
        icon={snackbarState.icon}
        show={showSnackbar}
        id='cookie-snackbar'
        dismissAfter={3000}
        onClose={() => setShowSnackbar(false)}
      />
    </>
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
