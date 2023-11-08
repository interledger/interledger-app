import type { PlainMessage } from '@bufbuild/protobuf'
import type { LoaderFunctionArgs, MetaFunction } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { useLoaderData } from '@remix-run/react'
import { useState } from 'react'
import { route } from 'routes-gen'
import type { ApplicationProps } from '~/components'
import {
  Alert,
  AlertContent,
  Card,
  CardButton,
  CardContent,
  CardHeader,
  CardLink,
  CardTitle,
  Chip,
  ChipColor,
  Dialog,
  DiscordIcon,
  FynbosIcon,
  Icon,
  Layouts,
  LinkedInIcon,
  Router,
  SlackIcon,
  TextButton,
  TwitterIcon
} from '~/components'
import { Label } from '~/components/Label'
import type { PublicWalletInfo } from '~/generated/connect/backend/v1/backend_pb'
import { isConnectError } from '~/lib/error.server'
import { grpc } from '~/lib/grpc.server'
import { mergeMeta } from '~/lib/meta'

export async function loader({ request, params }: LoaderFunctionArgs) {
  const url = new URL(request.url)
  const token = url.searchParams.get('token') || ''
  let link = await grpc.introspect(request, { token })
  if (isConnectError(link)) {
    throw link.errorResponse
  }

  if (!link.completed) {
    return redirect(route('/collect/card'))
  }

  let publicWalletInfo: PlainMessage<PublicWalletInfo>
  const publicWalletInfoResponse = await grpc.getPublicWalletInfo(request, {
    walletAddress: link.senderWalletUrl
  })

  if (isConnectError(publicWalletInfoResponse)) {
    publicWalletInfo = {
      walletID: 'not-found',
      address: link.senderWalletUrl,
      shortAddress: '',
      publicName: '',
      identities: [],
      canReceive: false
    }
  } else publicWalletInfo = publicWalletInfoResponse

  return json({
    senderWalletUrl: link.senderWalletUrl,
    formattedAmount: link.formattedAmount,
    formattedDate: link.formattedDate,
    formattedTime: link.formattedTime,
    note: link.note,
    mask: link.mask,
    publicWalletInfo
  })
}

export const handle: ApplicationProps = {
  layout: Layouts.Focus,
  scaffold: {
    header: {
      title: 'Success'
    }
  }
}

export const meta: MetaFunction = mergeMeta(() => [
  {
    title: 'Collected payment'
  }
])

export default function Page() {
  const {
    formattedAmount,
    formattedTime,
    formattedDate,
    mask,
    note,
    publicWalletInfo
  } = useLoaderData<typeof loader>()
  const [showDialog, setShowDialog] = useState<boolean>(false)

  return (
    <>
      <Alert color={ChipColor.green}>
        <AlertContent>
          You have collected a payment of {formattedAmount} to your card ending{' '}
          {mask}.
        </AlertContent>
      </Alert>
      <Card>
        <CardContent>
          <div className='flex items-center justify-between'>
            <h2 className='text-4xl font-medium'>{formattedAmount}</h2>
            <div className='flex flex-col items-end space-y-1'>
              <span className='text-sm font-medium text-medium'>
                {formattedDate}
              </span>
              <span className='text-xs text-weak'>{formattedTime}</span>
            </div>
          </div>
        </CardContent>
        <Label>Payment from</Label>
        <CardButton
          noHover
          onClick={() => {
            setShowDialog(true)
          }}
        >
          <div className='flex w-full items-center justify-between text-medium'>
            <div className='flex space-x-2'>
              <FynbosIcon />
              <span>{publicWalletInfo.publicName}</span>
            </div>
            <Icon>navigate_next</Icon>
          </div>
        </CardButton>

        {note && (
          <CardContent className='mt-2'>
            <div className='flex w-full flex-col space-y-1'>
              <span className='text-weak'>Payment note</span>
              <span className='text-medium'>{note}</span>
            </div>
          </CardContent>
        )}
      </Card>
      <Card>
        <CardHeader>
          <CardTitle>Sign up to Fynbos</CardTitle>
        </CardHeader>
        <CardContent className='flex space-x-4'>
          <div className='flex h-16 w-16 items-center justify-center rounded-full bg-nav'>
            <FynbosIcon />
          </div>
          <div className='flex-1 flex-col space-y-4'>
            <p className='text-medium'>
              Sign up to Fynbos and get rewarded for sending and receiving
              payments.
            </p>
            <Router
              className='text-sm font-medium text-primary'
              to={route('/signup')}
            >
              Sign up
            </Router>
          </div>
        </CardContent>
      </Card>
      <Dialog
        open={showDialog}
        setOpen={() => {
          setShowDialog(true)
        }}
      >
        <CardHeader>
          <h1 className='text-xl font-medium'>User information</h1>
        </CardHeader>
        {publicWalletInfo.walletID == 'not-found' && (
          <CardContent>
            <span className='text-medium'>
              This person is not a Fynbos user yet.
            </span>
          </CardContent>
        )}
        {publicWalletInfo.walletID !== 'not-found' && (
          <>
            <CardContent>
              <span className='text-medium'>
                You are viewing public information about the person you intend
                to pay.
              </span>
            </CardContent>
            <Label className='mt-4'>Public name</Label>
            <div className='mt-1 flex rounded-xl bg-nav p-3 text-medium'>
              <span className=''>{publicWalletInfo.publicName}</span>
            </div>
            <Label className='mt-2'>Wallet address</Label>
            <CardLink className='flex w-full' to={publicWalletInfo.address}>
              <div className='flex w-full items-center justify-between text-medium'>
                <div className='flex space-x-2'>
                  <FynbosIcon />
                  <span>{publicWalletInfo.shortAddress}</span>
                </div>
                <Icon>navigate_next</Icon>
              </div>
            </CardLink>
          </>
        )}

        {publicWalletInfo.identities.map((identity) => (
          <div key={identity.id} className='contents'>
            <Label className='mt-2 capitalize'>{identity.platform}</Label>
            <CardLink className='flex w-full' to={publicWalletInfo.address}>
              <div className='flex w-full items-center justify-between text-medium'>
                <div className='flex space-x-2'>
                  {identity.platform == 'twitter' && <TwitterIcon />}
                  {identity.platform == 'linkedin' && <LinkedInIcon />}
                  {identity.platform == 'discord' && <DiscordIcon />}
                  {identity.platform == 'slack' && <SlackIcon />}
                  <span>{identity.identifier}</span>
                </div>
                {identity.state == 'verified' && (
                  <Chip color={ChipColor.green}>Verified</Chip>
                )}
              </div>
            </CardLink>
          </div>
        ))}

        <CardContent className='flex w-full justify-end space-x-6'>
          <TextButton type='button' onClick={() => setShowDialog(false)}>
            Close
          </TextButton>
        </CardContent>
      </Dialog>
    </>
  )
}
