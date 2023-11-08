import type { PlainMessage } from '@bufbuild/protobuf'
import type { LoaderFunctionArgs, MetaFunction } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { useLoaderData, useNavigate } from '@remix-run/react'
import { useEffect, useState } from 'react'
import { route } from 'routes-gen'
import type { ApplicationProps } from '~/components'
import {
  Alert,
  AlertBody,
  AlertContent,
  AlertTitle,
  Button,
  Card,
  CardButton,
  CardContent,
  CardHeader,
  CardLink,
  Chip,
  ChipColor,
  Dialog,
  DiscordIcon,
  FynbosIcon,
  Icon,
  Layouts,
  LinkedInIcon,
  OutlineButton,
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
import { useScaffoldStore } from '~/lib/useScaffoldStore'

export async function loader({ request, params }: LoaderFunctionArgs) {
  if (process.env.FYNBOS_ENV == 'prod') {
    return redirect(route('/payments'))
  }

  const transaction = await grpc.lookupTransaction(request, {
    id: params.paymentId as string
  })

  if (isConnectError(transaction)) throw transaction.errorResponse

  const walletUrl = transaction.source
  let publicWalletInfo: PlainMessage<PublicWalletInfo>

  const publicWalletInfoResponse = await grpc.getPublicWalletInfo(request, {
    walletAddress: walletUrl
  })

  if (isConnectError(publicWalletInfoResponse)) {
    publicWalletInfo = {
      walletID: 'not-found',
      address: walletUrl,
      shortAddress: '',
      publicName: '',
      identities: [],
      canReceive: false
    }
  } else publicWalletInfo = publicWalletInfoResponse

  let rpc = await grpc.createPaymentLink(request, {
    transactionId: transaction.id
  })
  if (isConnectError(rpc)) throw rpc.errorResponse

  return json({
    transaction,
    shareUrl: rpc.url,
    publicWalletInfo
  })
}

export const handle: ApplicationProps = {
  layout: Layouts.Focus,
  scaffold: {
    header: {
      back: route('/payments'),
      title: 'Share Payment'
    }
  }
}

export const meta: MetaFunction = mergeMeta(() => [
  {
    title: 'Share Payment'
  }
])

export default function Page() {
  const [showDetails, setShowDetails] = useState<Boolean>(false)
  const nav = useNavigate()
  const { transaction, shareUrl } = useLoaderData<typeof loader>()
  const [pushSnackbar] = useScaffoldStore((state) => [state.pushSnackbar])

  useEffect(() => {
    navigator.clipboard.writeText(shareUrl).then(
      () => {
        pushSnackbar({
          id: 'copy-payment-link-success',
          message: 'The link has been copied to your clipboard.',
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
  }, [pushSnackbar, shareUrl])

  return (
    <>
      {showDetails && <ShareDetails />}
      {!showDetails && (
        <>
          <ShareSummary />
          <div className='flex w-full justify-end space-x-6 pt-2'>
            <OutlineButton
              type='button'
              className='basis-1/2'
              onClick={() =>
                nav(
                  route('/payments/:paymentId', { paymentId: transaction.id })
                )
              }
            >
              Cancel
            </OutlineButton>
            <Button
              type='submit'
              onClick={() => {
                setShowDetails(true)
              }}
            >
              View payment details
            </Button>
          </div>
        </>
      )}
    </>
  )
}

function ShareDetails() {
  const { transaction, publicWalletInfo, shareUrl } =
    useLoaderData<typeof loader>()
  const [showDialog, setShowDialog] = useState<boolean>(false)
  const [pushSnackbar] = useScaffoldStore((state) => [state.pushSnackbar])

  return (
    <>
      <Card>
        <CardContent>
          <div className='flex items-center justify-between'>
            <h2 className='text-4xl font-medium text-error'>
              {transaction.subtotal}
            </h2>
            <div className='flex flex-col items-end space-y-1'>
              <span className='text-sm font-medium text-medium'>
                {transaction.formattedDate}
              </span>
              <span className='text-xs text-weak'>
                {transaction.formattedTime}
              </span>
            </div>
          </div>
        </CardContent>
        <Label className='mt-2 text-weak'>Payment to</Label>
        <div className='my-1 flex space-x-2 rounded-xl bg-nav p-3'>
          <Icon>account_circle</Icon>
          <span>{transaction.destination}</span>
        </div>
      </Card>
      <Card>
        <Label className='text-weak'>Payment from</Label>
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
      </Card>
      <Card>
        <CardContent>
          <div className='flex w-full justify-between'>
            <span className='text-weak'>Payment ID</span>
            <span className='text-medium'>{transaction.id}</span>
          </div>
          <div className='mt-2 flex w-full justify-between'>
            <span className='text-weak'>Payment from</span>
            <span className='text-medium'>{transaction.accountTitle}</span>
          </div>
          <div className='mt-2 flex w-full justify-between'>
            <span className='text-weak'>Amount sent</span>
            <span className='text-medium'>{transaction.subtotal}</span>
          </div>
          <div className='mt-2 flex w-full justify-between'>
            <span className='text-weak'>Fees</span>
            <span className='text-medium'>{transaction.fees}</span>
          </div>
          {transaction.hasPaymentProtection && (
            <div className='mt-2 flex w-full justify-between'>
              <span className='text-weak'>Payment protection (3%)</span>
              <span className='text-medium'>
                {transaction.paymentProtectionAmount}
              </span>
            </div>
          )}
          <div className='mt-4 flex w-full justify-between font-medium'>
            <span className='text-medium'>Total amount to debit</span>
            <span className='text-medium'>{transaction.formattedAmount}</span>
          </div>
        </CardContent>
      </Card>
      {transaction.reference && (
        <Card>
          <CardContent>
            <div className='flex w-full flex-col space-y-1'>
              <span className='text-weak'>Payment note</span>
              <span className='text-medium'>{transaction.reference}</span>
            </div>
          </CardContent>
        </Card>
      )}
      {transaction.hasPaymentProtection && (
        <Card>
          <CardContent>
            <div className='flex w-full flex-col space-y-1'>
              <span className='text-weak'>Payment protection</span>
              <Router
                to='/payment-protection'
                className='font-medium text-primary'
              >
                Eligible
              </Router>
            </div>
          </CardContent>
        </Card>
      )}
      <Card>
        <CardContent className='space-y-2'>
          <div className='flex w-full flex-col space-y-1'>
            <span className='text-weak'>Share payment</span>
            <CardButton
              noHover
              type='button'
              onClick={() => {
                if (typeof navigator.share == 'undefined') {
                  navigator.clipboard.writeText(shareUrl).then(
                    () => {
                      pushSnackbar({
                        id: 'copy-payment-link-success',
                        message: 'The link has been copied to your clipboard.',
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
                } else navigator.share({ url: shareUrl })
              }}
              className='items-center justify-between'
            >
              <span className='text-left font-medium text-medium'>
                {shareUrl}
              </span>
              <Icon className='text-medium'>share</Icon>
            </CardButton>
          </div>
          <p className='text-weak'>
            Anyone with the link can collect the payment, therefore only share
            it with the intended receiver.
          </p>
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

function ShareSummary() {
  const { transaction, shareUrl } = useLoaderData<typeof loader>()
  const [pushSnackbar] = useScaffoldStore((state) => [state.pushSnackbar])

  return (
    <>
      <Card>
        <CardContent>
          <div className='flex items-center justify-between'>
            <h2 className='text-4xl font-medium text-error'>
              {transaction.subtotal}
            </h2>
            <div className='flex flex-col items-end space-y-1'>
              <span className='text-sm font-medium text-medium'>
                {transaction.formattedDate}
              </span>
              <span className='text-xs text-weak'>
                {transaction.formattedTime}
              </span>
            </div>
          </div>
        </CardContent>
        <Label className='mt-2 text-weak'>Share payment</Label>
        <CardButton
          noHover
          type='button'
          onClick={() => {
            if (typeof navigator.share == 'undefined') {
              navigator.clipboard.writeText(shareUrl).then(
                () => {
                  pushSnackbar({
                    id: 'copy-payment-link-success',
                    message: 'The link has been copied to your clipboard.',
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
            } else navigator.share({ url: shareUrl })
          }}
          className='items-center justify-between'
        >
          <span className='text-left font-medium text-medium'>{shareUrl}</span>
          <Icon className='text-medium'>share</Icon>
        </CardButton>
      </Card>
      <Alert>
        <Icon>notification_important</Icon>
        <AlertContent>
          <AlertTitle>
            Only share the payment with the intended receiver
          </AlertTitle>
          <AlertBody>
            Anyone with the link can collect the payment, therefore only share
            it with the intended receiver.
          </AlertBody>
        </AlertContent>
      </Alert>
    </>
  )
}
