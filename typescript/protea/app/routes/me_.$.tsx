import type {
  ActionFunctionArgs,
  LoaderFunctionArgs,
  MetaFunction
} from '@remix-run/node'
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
  Button,
  ButtonRouter,
  Card,
  CardContent,
  CardCopy,
  CardHeader,
  CardIcon,
  CardLink,
  CardTitle,
  Chip,
  ChipColor,
  DiscordIcon,
  InterledgerIcon,
  Icon,
  Layouts,
  Router,
  SlackIcon,
  TwitterIcon
} from '~/components'
import { Label } from '~/components/Label'
import { getPublicIdentities } from '~/data/identity.server'
import { getPublicWalletDetails } from '~/data/wallet.server'
import type { Query } from '~/generated/dato-cms-graphql'
import { isConnectError } from '~/lib/error.server'
import { grpc } from '~/lib/grpc.server'
import { getClientIP } from '~/lib/ip.server'
import { hasUserSession } from '~/lib/kratos.server'
import { mergeMeta } from '~/lib/meta'

export async function loader({ request, params }: LoaderFunctionArgs) {
  const walletAddressParam = params['*'] as string
  let profilePicture: { person: Query['person'] } | { person: null } = {
    person: null
  }

  const response = await grpc.getPublicWalletInfo(request, {
    walletAddress: walletAddressParam
  })

  if (isConnectError(response)) throw response.errorResponse

  const walletAddress = response

  if (request.headers.get('Content-type') == 'application/json')
    return redirect(walletAddress.address)

  const wallet = await getPublicWalletDetails(request, walletAddress.walletID)
  const identities = await getPublicIdentities(request, walletAddress.walletID)

  let canSendToAddress = false
  const isUser = hasUserSession(request)
  if (isUser) {
    const response = await grpc.getPaymentAddress(request, {
      address: walletAddressParam
    })
    if (!isConnectError(response)) {
      canSendToAddress = response.canSendToAddress
    }
  }

  return json({
    profilePicture,
    isUser,
    canSendToAddress,
    wallet,
    identities,
    walletAddress: walletAddress,
    paymentPointerParam: walletAddressParam
  })
}

export const handle: ApplicationProps = {
  layout: Layouts.Focus,
  scaffold: {
    header: {}
  }
}

export const meta: MetaFunction<typeof loader> = mergeMeta(
  ({ data, location }) => [
    {
      tagName: 'link',
      rel: 'monetization',
      href: data?.walletAddress.address
    }
  ]
)

export default function Page() {
  const {
    profilePicture,
    isUser,
    wallet,
    identities,
    canSendToAddress,
    walletAddress,
    paymentPointerParam
  } = useLoaderData<typeof loader>()

  return (
    <>
      <Card>
        <CardContent>
          <div className='flex w-full items-center justify-center'>
            {profilePicture.person && (
              <Image
                pictureClassName='m-0'
                className='aspect-square'
                data={
                  profilePicture.person?.avatar
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
          </h1>
        </CardContent>
        <Label className='mt-2'>Wallet address</Label>
        <CardCopy
          copyContent={walletAddress.address}
          shareData={{
            title: 'Wallet address',
            text: 'You can pay me using Fynbos with my wallet address.',
            url: walletAddress.address
          }}
          success='Wallet address copied to clipboard.'
          copyError="Couldn't copy to clipboard."
          shareError="Couldn't share wallet address."
        >
          {walletAddress.shortAddress}
        </CardCopy>
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
        {identities.discord && (
          <>
            <Label className='mt-4'>Discord</Label>
            {identities.discord.map((identity) => (
              <CardLink
                key={identity.id}
                className='items-center justify-between'
                to={route('/me/identities/:identityId', {
                  identityId: identity.signatureHash
                })}
              >
                <div className='flex space-x-3'>
                  <DiscordIcon />
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
        {identities.slack && (
          <>
            <Label className='mt-4'>Slack</Label>
            {identities.slack.map((identity) => (
              <CardLink
                key={identity.id}
                className='items-center justify-between'
                to={route('/me/identities/:identityId', {
                  identityId: identity.signatureHash
                })}
              >
                <div className='flex space-x-3'>
                  <SlackIcon />
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
      </Card>
      <Form
        id='me'
        action={`/me/${paymentPointerParam}`}
        method='post'
        className='hidden'
      />
      <input
        form='me'
        value={walletAddress.shortAddress}
        name='paymentPointer'
        type='hidden'
      />
      <Button disabled={!isUser || !canSendToAddress} form='me' type='submit'>
        Send a payment
      </Button>
      {!canSendToAddress && isUser && (
        <p className='-mt-2 text-center text-xs text-medium'>
          You can't send payments to {wallet.publicName}.
        </p>
      )}
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
                <InterledgerIcon />
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

export async function action({ request, params }: ActionFunctionArgs) {
  // TODO: create payment here and redirect to /pay/:paymentId
  const walletAddressParam = params['*'] as string

  const clientIpAddress = getClientIP(request)

  const response = await grpc.getPaymentAddress(request, {
    address: walletAddressParam
  })

  if (isConnectError(response)) throw response.errorResponse

  const payment = await grpc.createPayment(request, {
    receiverIdentity: response.walletUrl,
    receiverIdentityType: 3,
    ipAddress: clientIpAddress
  })
  if (isConnectError(payment)) throw payment.errorResponse

  return redirect(route('/pay/:paymentId', { paymentId: payment.id }))
}
