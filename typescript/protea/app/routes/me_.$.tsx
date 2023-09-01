import type { ActionArgs, LoaderArgs } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import {
  Form,
  isRouteErrorResponse,
  useLoaderData,
  useParams,
  useRouteError
} from '@remix-run/react'
import { captureRemixErrorBoundaryError } from '@sentry/remix'
import clsx from 'clsx'
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
  TwitterIcon
} from '~/components'
import { Label } from '~/components/Label'
import { hasUserSession } from '~/lib/kratos.server'
import { getPerson } from '~/lib/marketing.server'
import {
  StatusError,
  httpMapping,
  isGrpcError,
  openPaymentsClient
} from '~/lib/proto.server'
import { PayStep } from '~/lib/usePayStore'
import { useScaffoldStore } from '~/lib/useScaffoldStore'
import {
  getPublicLinkedIdentities,
  getPublicWalletDetails,
  getWalletInfo
} from '~/lib/wallet.server'

export async function loader({ request, params }: LoaderArgs) {
  const walletAddressParam = params['*'] as string
  let profilePicture = null

  if (walletAddressParam.includes('fynbos.me/adrian')) {
    profilePicture = await getPerson({
      filter: { name: { eq: 'Adrian' } }
    })
  }
  if (walletAddressParam.includes('fynbos.me/matt')) {
    profilePicture = await getPerson({
      filter: { name: { eq: 'Matt' } }
    })
  }

  const response = await openPaymentsClient
    .getPaymentPointer({ url: walletAddressParam })
    .then((v) => v)
    .catch(StatusError)
  if (isGrpcError(response)) {
    throw json({}, httpMapping(response.code))
  }

  const walletAddress = response.response

  if (request.headers.get('Content-type') == 'application/json')
    return redirect(walletAddress.url)

  const wallet = await getPublicWalletDetails(request, walletAddress.walletID)
  const identities = await getPublicLinkedIdentities(
    request,
    walletAddress.walletID
  )

  let editable = false
  const isUser = hasUserSession(request)
  if (isUser) {
    const walletPaymentPointer = await getWalletInfo(request)
    if (walletPaymentPointer.formattedURL == walletAddress.formatted)
      editable = true
  }

  return json({
    profilePicture,
    isUser,
    editable,
    wallet,
    identities,
    walletAddress: walletAddress.url.replace(/(http(s)?:\/\/)/i, ''),
    paymentPointer: walletAddress,
    paymentPointerParam: walletAddressParam
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

  const [pushSnackbar] = useScaffoldStore((state) => [state.pushSnackbar])

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
        </CardContent>
        <Label className='mt-2'>Wallet address</Label>
        <CardButton
          noHover
          type='button'
          onClick={() => {
            if (typeof navigator.clipboard == 'undefined') {
              pushSnackbar({
                id: 'copy-to-clipboard-fail',
                message: "Couldn't copy to clipboard.",
                icon: 'close',
                canShow: true
              })
            } else
              navigator.clipboard.writeText(walletAddress).then(
                () => {
                  pushSnackbar({
                    id: 'copy-wallet-address-success',
                    message: 'Wallet address copied to clipboard.',
                    icon: 'close',
                    canShow: true
                  })
                },
                () => {
                  pushSnackbar({
                    id: 'copy-to-clipboard-fail',
                    message: "Couldn't copy to clipboard.",
                    icon: 'close',
                    canShow: true
                  })
                }
              )
          }}
          className='items-center justify-between'
        >
          <span className='text-medium'>{walletAddress}</span>
          <Icon className='text-medium'>content_copy</Icon>
        </CardButton>
        {identities.twitter && (
          <>
            <Label className='mt-4'>Twitter</Label>
            {identities.twitter.map((identity) => (
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
          </>
        )}
        {identities.domain && (
          <>
            <Label className='mt-4'>Domain</Label>
            {identities.domain.map((identity) => (
              <CardLink
                key={identity.id}
                className='items-center justify-between'
                to={route('/me/identities/:identityId', {
                  identityId: identity.signatureHash
                })}
              >
                <div className='flex space-x-3'>
                  <Icon>captive_portal</Icon>
                  <span>{identity.identifier}</span>
                </div>
                <div className='flex space-x-3'>
                  {identity.state == 'verified' && (
                    <Chip color={ChipColor.green}>Verified</Chip>
                  )}
                  <Icon>navigate_next</Icon>
                </div>
              </CardLink>
            ))}
          </>
        )}
        {paymentPointerParam.includes('fynbos.me/adrian') && (
          <>
            <Label className='mt-4'>LinkedIn</Label>
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
            <Label className='mt-4'>LinkedIn</Label>
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
      {/* TODO Disable this if can't send to this user too */}
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
    </>
  )
}

export function ErrorBoundary() {
  const error = useRouteError()
  const params = useParams()

  if (isRouteErrorResponse(error)) {
    if (error.status == 404) {
      captureRemixErrorBoundaryError(error)
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
            <div className='flex items-center justify-between rounded-xl bg-nav p-3'>
              {params['*'] && (params['*'] as string)}
              <Chip color={ChipColor.green}>Available</Chip>
            </div>
          </Card>
          {/* TODO This should prefill the /wallet-address page for the user with the current address*/}
          <ButtonRouter to={route('/signup')}>
            Claim wallet address
          </ButtonRouter>
        </>
      )
    }
  }

  throw error
}

export async function action({ request, params }: ActionArgs) {
  const walletAddressParam = params['*'] as string
  return redirect(`/pay?address=${walletAddressParam}&start=${PayStep.AMOUNT}`)
}
