import type { ActionArgs, LoaderArgs } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import {
  Form,
  isRouteErrorResponse,
  useLoaderData,
  useParams,
  useRouteError
} from '@remix-run/react'
import clsx from 'clsx'
import { useState } from 'react'
import type { ResponsiveImageType } from 'react-datocms'
import { Image } from 'react-datocms'
import { route } from 'routes-gen'
import type { ApplicationProps } from '~/components'
import {
  AnchorRouter,
  Button,
  ButtonRouter,
  Card,
  CardButton,
  CardContent,
  CardHeader,
  CardIcon,
  CardLink,
  CardTitle,
  Chip,
  ChipColor,
  FynbosIcon,
  Icon,
  Layouts,
  LinkedInIcon,
  Router,
  Snackbar,
  TwitterIcon
} from '~/components'
import { Label } from '~/components/Label'
import { flowType, requireFlow, updateFlow } from '~/lib/flows.server'
import { hasUserSession } from '~/lib/kratos.server'
import { getPerson } from '~/lib/marketing.server'
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
  let profilePicture = null

  if (paymentPointerParam.includes('fynbos.me/adrian')) {
    profilePicture = await getPerson({
      filter: { name: { eq: 'Adrian' } }
    })
  }
  if (paymentPointerParam.includes('fynbos.me/matt')) {
    profilePicture = await getPerson({
      filter: { name: { eq: 'Matt' } }
    })
  }

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
    profilePicture,
    isUser,
    editable,
    wallet,
    identities,
    walletAddress: paymentPointer.url.replace(/(http(s)?:\/\/)/i, ''),
    paymentPointer,
    paymentPointerParam
  })
}

export const handle: ApplicationProps = {
  layout: Layouts.Focus,
  scaffold: {
    header: {}
  }
}

export default function Page() {
  const {
    profilePicture,
    isUser,
    wallet,
    identities,
    walletAddress,
    paymentPointer,
    paymentPointerParam
  } = useLoaderData<typeof loader>()

  const [snackbarState, setSnackbar] = useState<any>({
    message: 'Wallet address copied to clipboard.',
    icon: 'close',
    show: false
  })
  const [showSnackbar, setShowSnackbar] = useState<boolean>(false)

  return (
    <>
      <Card>
        <CardContent>
          <div className='flex w-full items-center justify-center'>
            {profilePicture && (
              <Image
                pictureClassName='m-0'
                className='aspect-square'
                data={
                  profilePicture?.person?.avatar
                    ?.responsiveImage as ResponsiveImageType
                }
              />
            )}
          </div>

          <h1
            className={clsx(
              profilePicture && 'mt-6',
              'text-center text-2xl font-medium'
            )}
          >
            {wallet.publicName}
            {/*TODO Possibly put this edit button in the header*/}
            {/*{editable && (*/}
            {/*  <Router to={route('/settings/profile-public')}>*/}
            {/*    <Icon>edit</Icon>*/}
            {/*  </Router>*/}
            {/*)}*/}
          </h1>
          <Label className='-mb-5 mt-6'>Wallet address</Label>
        </CardContent>
        <CardButton
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
              navigator.clipboard.writeText(walletAddress).then(
                () => {
                  setSnackbar({
                    message: 'Wallet address copied to clipboard.',
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
          className='items-center justify-between'
        >
          <span className='text-medium'>{walletAddress}</span>
          <Icon className='text-medium'>content_copy</Icon>
        </CardButton>
        {identities.length > 0 && (
          <CardContent>
            <Label className='-mb-5'>Twitter</Label>
          </CardContent>
        )}
        {identities.map((identity) => (
          <CardLink
            key={identity.id}
            className='items-center justify-between'
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
          </CardLink>
        ))}
        {paymentPointerParam.includes('fynbos.me/adrian') && (
          <>
            <p className='ml-2 mt-6 text-sm font-medium text-medium'>
              LinkedIn
            </p>
            <AnchorRouter
              className='mt-2 flex items-center justify-between rounded-xl bg-nav p-3 text-medium hover:bg-nav-hover'
              to='https://www.linkedin.com/in/adrianhopebailie/'
            >
              <div className='flex space-x-3'>
                <LinkedInIcon />
                <span>Adrian Hope-Bailie</span>
              </div>
              <div className='flex space-x-3'>
                <Icon>navigate_next</Icon>
              </div>
            </AnchorRouter>
          </>
        )}
        {paymentPointerParam.includes('fynbos.me/matt') && (
          <>
            <p className='ml-2 mt-6 text-sm font-medium text-medium'>
              LinkedIn
            </p>
            <AnchorRouter
              className='mt-2 flex items-center justify-between rounded-xl bg-nav p-3 text-medium hover:bg-nav-hover'
              to='https://www.linkedin.com/in/matthew-de-haast-aa448884/'
            >
              <div className='flex space-x-3'>
                <LinkedInIcon />
                <span>Matthew de Haast</span>
              </div>
              <div className='flex space-x-3'>
                <Icon>navigate_next</Icon>
              </div>
            </AnchorRouter>
          </>
        )}
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
      <Button disabled={!isUser} form='me' type='submit'>
        Send a payment
      </Button>
      {!isUser && (
        <p className='-mt-2 text-center text-xs text-medium'>
          Payments are currently in beta and are only enabled for certain users.
        </p>
      )}
      {!isUser && (
        <Card>
          <CardHeader>
            <CardTitle>Join the waitlist</CardTitle>
          </CardHeader>
          <CardContent>
            <div className='flex items-start space-x-4'>
              <CardIcon>
                <FynbosIcon />
              </CardIcon>
              <div className='flex flex-col space-y-4'>
                <p className='text-sm text-medium'>
                  For a secure, programmable digital wallet that connects all
                  your accounts, join the waitlist now.
                </p>
                <Router
                  className='text-sm font-medium text-primary'
                  to={route('/waitlist')}
                >
                  Join the waitlist
                </Router>
              </div>
            </div>
          </CardContent>
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

export function ErrorBoundary() {
  const error = useRouteError()
  const params = useParams()

  if (isRouteErrorResponse(error)) {
    if (error.status == 404)
      return (
        <>
          <Card>
            <CardHeader>
              <CardTitle>Available wallet address</CardTitle>
            </CardHeader>
            <CardContent>
              <p className='text-medium'>
                This is not yet a registered wallet address.
              </p>
            </CardContent>
            <div className='m-2 mt-0 flex items-center justify-between rounded-xl bg-nav p-3'>
              {params['*'] && (params['*'] as string)}
              <Chip color={ChipColor.green}>Available</Chip>
            </div>
          </Card>
          <ButtonRouter to={route('/signup')}>
            Claim wallet address
          </ButtonRouter>
        </>
      )
  }

  throw error
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
