import type { PlainMessage } from '@bufbuild/protobuf/dist/types/message'
import type { LoaderFunctionArgs, MetaFunction } from '@remix-run/node'
import { json } from '@remix-run/node'
import type { UIMatch } from '@remix-run/react'
import { useLoaderData } from '@remix-run/react'
import { useState } from 'react'
import { route } from 'routes-gen'
import type { ApplicationProps } from '~/components'
import {
  Alert,
  AlertBody,
  AlertContent,
  AlertTitle,
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
import { getPusherArgs } from '~/lib/pusher.server'
import { usePusher } from '~/lib/usePusher'
import { useScaffoldStore } from '~/lib/useScaffoldStore'

export async function loader({ request, params }: LoaderFunctionArgs) {
  const transaction = await grpc.lookupTransaction(request, {
    id: params.paymentId as string
  })

  if (isConnectError(transaction)) throw transaction.errorResponse

  const walletUrl = transaction.type.includes('outgoing')
    ? transaction.destination
    : transaction.source

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

  const pusherArgs = await getPusherArgs(request)

  return json({
    publicWalletInfo,
    transaction,
    pusherArgs
  })
}

export enum TransactionRefundState {
  NA = 0,
  PENDING = 1,
  COMPLETED = 2
}

export const handle: ApplicationProps = {
  layout: Layouts.Wallet,
  scaffold: {
    header: {
      back: route('/payments'),
      title: 'Payment',
      actions: (match: UIMatch<typeof loader>) => {
        if (
          match.data.transaction.refundState == TransactionRefundState.PENDING
        ) {
          return {
            key: 'Pending refund',
            nodes: <Chip color={ChipColor.red}>Pending refund</Chip>
          }
        } else if (
          match.data.transaction.refundState == TransactionRefundState.COMPLETED
        ) {
          return {
            key: 'Refunded',
            nodes: <Chip color={ChipColor.red}>Refunded</Chip>
          }
        }
        switch (match.data.transaction.state) {
          case 'Completed':
            return {
              key: 'Complete',
              nodes: <Chip color={ChipColor.green}>Complete</Chip>
            }
          case 'Pending':
            return {
              key: 'Pending',
              nodes: (
                <Chip color={ChipColor.orange}>
                  Pending
                  {match.data.transaction.hasPaymentLink && ' collection'}
                </Chip>
              )
            }
          case 'Failed':
            return {
              key: 'Failed',
              nodes: <Chip color={ChipColor.red}>Failed</Chip>
            }
          default:
            return null
        }
      }
    },
    isNested: true
  }
}

export const meta: MetaFunction<typeof loader> = mergeMeta(({ data }) => [
  {
    title:
      typeof data == 'undefined'
        ? 'Payment'
        : data.transaction.type.includes('outgoing')
        ? `${data.transaction.subtotal} to ${data.transaction.title}`
        : `${data.transaction.formattedAmount} from ${data.transaction.title}`
  }
])

export default function Page() {
  const { transaction, publicWalletInfo, pusherArgs } =
    useLoaderData<typeof loader>()
  const [showDialog, setShowDialog] = useState<boolean>(false)
  usePusher(pusherArgs, ['transaction', 'kyc'])

  return (
    <>
      {transaction.type.includes('outgoing') && (
        <Outgoing openDialog={() => setShowDialog(true)} />
      )}
      {transaction.type.includes('incoming') && (
        <Incoming openDialog={() => setShowDialog(true)} />
      )}
      <Dialog open={showDialog} setOpen={setShowDialog}>
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

function Outgoing({ openDialog }: { openDialog: () => void }) {
  const { transaction } = useLoaderData<typeof loader>()

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
        <Label className='mt-2'>Payment to</Label>
        {transaction.destinationIdentityType !== 'Unknown' && (
          <CardButton noHover onClick={openDialog}>
            <div className='flex w-full items-center justify-between text-medium'>
              <div className='flex space-x-2'>
                {transaction.destinationIdentityType === 'wallet' && (
                  <FynbosIcon />
                )}
                {transaction.destinationIdentityType === 'linkedin' && (
                  <TwitterIcon />
                )}
                {transaction.destinationIdentityType === 'twitter' && (
                  <LinkedInIcon />
                )}
                {transaction.destinationIdentityType === 'discord' && (
                  <DiscordIcon />
                )}
                {transaction.destinationIdentityType === 'slack' && (
                  <SlackIcon />
                )}
                <span>{transaction.title}</span>
              </div>
              <Icon>navigate_next</Icon>
            </div>
          </CardButton>
        )}
        {transaction.destinationIdentityType === 'Unknown' && (
          <CardButton noHover onClick={openDialog} disabled>
            <div className='flex w-full items-center justify-between text-medium'>
              <div className='flex space-x-2'>
                <Icon className=''>account_circle</Icon>
                <span>{transaction.title}</span>
              </div>
            </div>
          </CardButton>
        )}
      </Card>
      {transaction.state == 'Pending' && transaction.hasPaymentLink && (
        <Alert>
          <Icon>schedule</Icon>
          <AlertContent>
            <AlertTitle>
              Payment expires {transaction.formattedPaymentLinkExpiryDate}
            </AlertTitle>
            <AlertBody>
              Uncollected payments will automatically be refunded on the date
              indicated.
            </AlertBody>
          </AlertContent>
        </Alert>
      )}
      {transaction.refundState == TransactionRefundState.PENDING && (
        <Alert>
          <Icon>error</Icon>
          <AlertContent>
            <AlertTitle>Pending refund</AlertTitle>
            <AlertBody>
              Any money, including fees, debited from your account will be
              returned.
            </AlertBody>
          </AlertContent>
        </Alert>
      )}
      {transaction.refundState == TransactionRefundState.COMPLETED && (
        <Alert>
          <Icon>error</Icon>
          <AlertContent>
            <AlertTitle>Payment unsuccessful</AlertTitle>
            <AlertBody>
              All money, including fees, debited from your account has been
              returned.
            </AlertBody>
          </AlertContent>
        </Alert>
      )}
      {transaction.refundState == TransactionRefundState.NA &&
        transaction.state == 'Failed' && (
          <Alert>
            <Icon>error</Icon>
            <AlertContent>
              <AlertTitle>Payment unsuccessful</AlertTitle>
              <AlertBody>
                No need to worry. Nothing has been debited from your account.
              </AlertBody>
            </AlertContent>
          </Alert>
        )}
      {transaction.refundState != TransactionRefundState.NA &&
        transaction.state == 'Failed' && (
          <Card>
            <Label>Payment ID</Label>
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
                  navigator.clipboard.writeText(transaction.id).then(
                    () => {
                      pushSnackbar({
                        id: 'copy-wallet-address-success',
                        message: 'Payment ID copied to clipboard.',
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
              <span className='text-left font-medium text-medium'>
                {transaction.id}
              </span>
              <Icon className='text-medium'>content_copy</Icon>
            </CardButton>
            <CardContent>
              <div className='flex w-full justify-between'>
                <span className='text-weak'>Total fees</span>
                <span className='text-medium'>{transaction.fees}</span>
              </div>
              <div className='mt-2 flex w-full justify-between'>
                <span className='text-weak'>You paid</span>
                <span className='text-medium'>
                  {transaction.formattedAmount}
                </span>
              </div>
              {transaction.refundState == TransactionRefundState.COMPLETED && (
                <div className='mt-2 flex w-full justify-between'>
                  <span className='text-weak'>Your refund</span>
                  <span className='text-medium'>
                    {transaction.formattedAmount}
                  </span>
                </div>
              )}
            </CardContent>
          </Card>
        )}
      {transaction.state != 'Failed' && (
        <Card>
          <Label>Payment ID</Label>
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
                navigator.clipboard.writeText(transaction.id).then(
                  () => {
                    pushSnackbar({
                      id: 'copy-wallet-address-success',
                      message: 'Payment ID copied to clipboard.',
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
            <span className='text-left font-medium text-medium'>
              {transaction.id}
            </span>
            <Icon className='text-medium'>content_copy</Icon>
          </CardButton>
          <CardContent>
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
      )}

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
      {transaction.state == 'Pending' && transaction.hasPaymentLink && (
        <>
          <Card>
            <Label>Share payment</Label>
            <CardButton
              noHover
              type='button'
              onClick={() => {
                if (typeof navigator.share == 'undefined') {
                  navigator.clipboard
                    .writeText(transaction.paymentLinkUrl)
                    .then(
                      () => {
                        pushSnackbar({
                          id: 'copy-wallet-address-success',
                          message:
                            'The link has been copied to your clipboard.',
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
                } else navigator.share({ url: transaction.paymentLinkUrl })
              }}
              className='items-center justify-between'
            >
              <span className='truncate text-left font-medium text-medium'>
                {transaction.paymentLinkUrl}
              </span>
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
                Anyone with the link can collect the payment, therefore only
                share it with the intended receiver.
              </AlertBody>
            </AlertContent>
          </Alert>
        </>
      )}
    </>
  )
}

function Incoming({ openDialog }: { openDialog: () => void }) {
  const { transaction } = useLoaderData<typeof loader>()

  const [pushSnackbar] = useScaffoldStore((state) => [state.pushSnackbar])

  return (
    <>
      <Card>
        <CardContent>
          <div className='flex items-center justify-between'>
            <h2 className='text-4xl font-medium'>{transaction.subtotal}</h2>
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
        <Label className='mt-2'>Payment from</Label>
        <CardButton noHover onClick={openDialog}>
          <div className='flex w-full items-center justify-between text-medium'>
            <div className='flex space-x-2'>
              {transaction.destinationIdentityType === 'wallet' && (
                <FynbosIcon />
              )}
              {transaction.destinationIdentityType === 'linkedin' && (
                <TwitterIcon />
              )}
              {transaction.destinationIdentityType === 'twitter' && (
                <LinkedInIcon />
              )}
              {transaction.destinationIdentityType === 'discord' && (
                <DiscordIcon />
              )}
              {transaction.destinationIdentityType === 'slack' && <SlackIcon />}
              <span>{transaction.title}</span>
            </div>
            <Icon>navigate_next</Icon>
          </div>
        </CardButton>
      </Card>
      <Card>
        <Label>Payment ID</Label>
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
              navigator.clipboard.writeText(transaction.id).then(
                () => {
                  pushSnackbar({
                    id: 'copy-wallet-address-success',
                    message: 'Payment ID copied to clipboard.',
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
          <span className='text-left font-medium text-medium'>
            {transaction.id}
          </span>
          <Icon className='text-medium'>content_copy</Icon>
        </CardButton>
        <CardContent>
          <div className='mt-2 flex justify-between'>
            <span className='text-weak'>Paid to</span>
            <span className='text-medium'>{transaction.accountTitle}</span>
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
            <div className='mb-4 flex w-full justify-between'>
              <span className='text-weak'>Payment protection</span>
              <Router
                to='/payment-protection'
                className='font-medium text-primary'
              >
                Eligible
              </Router>
            </div>
            <span className='text-medium'>
              For any disputes, please{' '}
              <Router to='/support' className='font-medium text-primary'>
                contact support
              </Router>
              .
            </span>
          </CardContent>
        </Card>
      )}
    </>
  )
}
