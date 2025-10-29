import type { PlainMessage } from '@bufbuild/protobuf/dist/types/message'
import type { LoaderFunctionArgs, MetaFunction } from '@remix-run/node'
import { json } from '@remix-run/node'
import type { UIMatch } from '@remix-run/react'
import { Link, useLoaderData } from '@remix-run/react'
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
  CardCopy,
  CardHeader,
  CardLink,
  Chip,
  ChipColor,
  Dialog,
  DiscordIcon,
  Icon,
  InterledgerIcon,
  Layouts,
  LinkedInIcon,
  SlackIcon,
  TextButton,
  TwitterIcon,
  WebMoLogo
} from '~/components'
import { Label } from '~/components/Label'
import type { PublicWalletInfo } from '~/generated/connect/backend/v1/backend_pb'
import { isConnectError } from '~/lib/error.server'
import { grpc } from '~/lib/grpc.server'
import { mergeMeta } from '~/lib/meta'
import { getPusherArgs } from '~/lib/pusher.server'
import { usePusher } from '~/lib/usePusher'

export async function loader({ request, params }: LoaderFunctionArgs) {
  let senderAccountTitle, receiverAccountTitle

  const transaction = await grpc.lookupTransaction(request, {
    id: params.paymentId as string
  })

  if (isConnectError(transaction)) throw transaction.errorResponse

  if (transaction.type == 'withdrawal' || transaction.type == 'deposit') {
    const linkedAccountsResponse = await grpc.getLinkedAccounts(request, {})
    if (isConnectError(linkedAccountsResponse))
      throw linkedAccountsResponse.error
    senderAccountTitle = linkedAccountsResponse.linkedAccounts.find(
      (account) => account.id == transaction.senderAccountId
    )?.title
    receiverAccountTitle = linkedAccountsResponse.linkedAccounts.find(
      (account) => account.id == transaction.receiverAccountId
    )?.title
  }

  const walletUrl =
    transaction.type == 'sent' ? transaction.destination : transaction.source

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
    senderAccountTitle,
    receiverAccountTitle,
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
              nodes: <Chip color={ChipColor.orange}>Pending</Chip>
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
        : // TODO Fix this for withdrawal
        data.transaction.type == 'sent'
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
      {(transaction.type == 'sent' ||
        transaction.type == 'web_monetization_outgoing') && (
        <Sent openDialog={() => setShowDialog(true)} />
      )}
      {(transaction.type == 'received' ||
        transaction.type == 'web_monetization_incoming') && (
        <Received openDialog={() => setShowDialog(true)} />
      )}
      {transaction.type == 'card_transaction' && <CardTransaction />}
      {transaction.type == 'withdrawal' && <Withdrawal />}
      {transaction.type == 'deposit' && <Deposit />}
      <Dialog open={showDialog} setOpen={setShowDialog}>
        <CardHeader>
          <h1 className='text-xl font-medium'>User information</h1>
        </CardHeader>
        {publicWalletInfo.walletID == 'not-found' && (
          <CardContent>
            <span className='text-medium'>
              This person is not an Interledger Wallet user yet.
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
                  <InterledgerIcon />
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

function Withdrawal() {
  const { senderAccountTitle, receiverAccountTitle, transaction } =
    useLoaderData<typeof loader>()

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
        <Label className='mt-2'>Withdrawal to</Label>
        <div className='my-1 flex space-x-2 rounded-xl bg-nav p-3'>
          <div className='flex w-full items-center justify-between text-medium'>
            <div className='flex space-x-2'>
              <Icon>account_balance</Icon>
              <span>{receiverAccountTitle}</span>
            </div>
          </div>
        </div>
      </Card>
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
            <AlertTitle>Withdrawal unsuccessful</AlertTitle>
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
              <AlertTitle>Withdrawal unsuccessful</AlertTitle>
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
            <CardCopy
              copyContent={transaction.id}
              shareData={{
                title: 'Payment ID',
                text: transaction.id
              }}
              success='Payment ID copied to clipboard.'
              copyError="Couldn't copy to clipboard."
              shareError="Couldn't share Payment ID."
            >
              {transaction.id}
            </CardCopy>
            <CardContent>
              <div className='flex w-full justify-between'>
                <span className='text-weak'>Total fees</span>
                <span className='text-medium'>{transaction.fees}</span>
              </div>
              <div className='mt-2 flex w-full justify-between'>
                <span className='text-weak'>You received</span>
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
          <CardCopy
            copyContent={transaction.id}
            shareData={{
              title: 'Payment ID',
              text: transaction.id
            }}
            success='Payment ID copied to clipboard.'
            copyError="Couldn't copy to clipboard."
            shareError="Couldn't share Payment ID."
          >
            {transaction.id}
          </CardCopy>
          <CardContent>
            <div className='mt-2 flex w-full justify-between'>
              <span className='text-weak'>Withdrawal from</span>
              <span className='text-medium'>{senderAccountTitle}</span>
            </div>
            <div className='mt-2 flex w-full justify-between'>
              <span className='text-weak'>Amount sent</span>
              <span className='text-medium'>{transaction.subtotal}</span>
            </div>
            <div className='mt-2 flex w-full justify-between'>
              <span className='text-weak'>Fees</span>
              <span className='text-medium'>{transaction.fees}</span>
            </div>
            <div className='mt-4 flex w-full justify-between font-medium'>
              <span className='text-medium'>Total amount withdrawn</span>
              <span className='text-medium'>{transaction.fundsReceived}</span>
            </div>
          </CardContent>
        </Card>
      )}

      {transaction.reference && (
        <Card>
          <CardContent>
            <div className='flex w-full flex-col space-y-1'>
              <span className='text-weak'>Withdraw note</span>
              <span className='text-medium'>{transaction.reference}</span>
            </div>
          </CardContent>
        </Card>
      )}
    </>
  )
}

function Deposit() {
  const { receiverAccountTitle, transaction } = useLoaderData<typeof loader>()

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
        <Label className='mt-2'>Deposit to</Label>
        <div className='my-1 flex space-x-2 rounded-xl bg-nav p-3'>
          <div className='flex w-full items-center justify-between text-medium'>
            <div className='flex space-x-2'>
              <InterledgerIcon />
              <span>{receiverAccountTitle}</span>
            </div>
          </div>
        </div>
      </Card>
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
            <AlertTitle>Deposit unsuccessful</AlertTitle>
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
              <AlertTitle>Deposit unsuccessful</AlertTitle>
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
            <CardCopy
              copyContent={transaction.id}
              shareData={{
                title: 'Payment ID',
                text: transaction.id
              }}
              success='Payment ID copied to clipboard.'
              copyError="Couldn't copy to clipboard."
              shareError="Couldn't share Payment ID."
            >
              {transaction.id}
            </CardCopy>
            <CardContent>
              <div className='flex w-full justify-between'>
                <span className='text-weak'>Total fees</span>
                <span className='text-medium'>{transaction.fees}</span>
              </div>
              <div className='mt-2 flex w-full justify-between'>
                <span className='text-weak'>You received</span>
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
          <CardCopy
            copyContent={transaction.id}
            shareData={{
              title: 'Payment ID',
              text: transaction.id
            }}
            success='Payment ID copied to clipboard.'
            copyError="Couldn't copy to clipboard."
            shareError="Couldn't share Payment ID."
          >
            {transaction.id}
          </CardCopy>
          <CardContent>
            <div className='mt-2 flex w-full justify-between'>
              <span className='text-weak'>Fees</span>
              <span className='text-medium'>{transaction.fees}</span>
            </div>
            <div className='mt-4 flex w-full justify-between font-medium'>
              <span className='text-medium'>Total amount deposited</span>
              <span className='text-medium'>{transaction.formattedAmount}</span>
            </div>
            <div className='mt-4 flex w-full justify-between font-medium'>
              <span className='text-medium'>Total amount received</span>
              <span className='text-medium'>{transaction.fundsReceived}</span>
            </div>
          </CardContent>
        </Card>
      )}

      {transaction.reference && (
        <Card>
          <CardContent>
            <div className='flex w-full flex-col space-y-1'>
              <span className='text-weak'>Deposit note</span>
              <span className='text-medium'>{transaction.reference}</span>
            </div>
          </CardContent>
        </Card>
      )}
    </>
  )
}

function Sent({ openDialog }: { openDialog: () => void }) {
  const { transaction } = useLoaderData<typeof loader>()

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
        <CardButton noHover onClick={openDialog}>
          <div className='flex w-full items-center justify-between text-medium'>
            <div className='flex space-x-2'>
              {transaction.destinationIdentityType === 'wallet' && (
                <InterledgerIcon />
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
      {transaction.type == 'web_monetization_outgoing' && (
        <Alert>
          <WebMoLogo />
          <AlertContent>
            <AlertTitle>Web monetization</AlertTitle>
            <AlertBody>
              Payments for the web monetized sites you support are consolidated
              and refreshed daily.
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
            <CardCopy
              copyContent={transaction.id}
              shareData={{
                title: 'Payment ID',
                text: transaction.id
              }}
              success='Payment ID copied to clipboard.'
              copyError="Couldn't copy to clipboard."
              shareError="Couldn't share Payment ID."
            >
              {transaction.id}
            </CardCopy>
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
          <CardCopy
            copyContent={transaction.id}
            shareData={{
              title: 'Payment ID',
              text: transaction.id
            }}
            success='Payment ID copied to clipboard.'
            copyError="Couldn't copy to clipboard."
            shareError="Couldn't share Payment ID."
          >
            {transaction.id}
          </CardCopy>
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
    </>
  )
}

function Received({ openDialog }: { openDialog: () => void }) {
  const { transaction } = useLoaderData<typeof loader>()

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
                <InterledgerIcon />
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
      {transaction.type == 'web_monetization_incoming' && (
        <Alert>
          <WebMoLogo />
          <AlertContent>
            <AlertTitle>Web monetization</AlertTitle>
            <AlertBody>
              Payments for the web monetized sites you support are consolidated
              and refreshed daily.
            </AlertBody>
          </AlertContent>
        </Alert>
      )}
      <Card>
        <Label>Payment ID</Label>
        <CardCopy
          copyContent={transaction.id}
          shareData={{
            title: 'Payment ID',
            text: transaction.id
          }}
          success='Payment ID copied to clipboard.'
          copyError="Couldn't copy to clipboard."
          shareError="Couldn't share Payment ID."
        >
          {transaction.id}
        </CardCopy>
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
    </>
  )
}

function CardTransaction() {
  const { transaction } = useLoaderData<typeof loader>()

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
        <div className='my-1 flex space-x-2 rounded-xl bg-nav p-3'>
          <div className='flex w-full items-center justify-between text-medium'>
            <div className='flex space-x-2'>
              <Icon>
                {transaction.cardTransactionDetails?.type === '0' &&
                  'local_mall'}
                {transaction.cardTransactionDetails?.type === '1' && 'atm'}
              </Icon>
              <span>{transaction.title}</span>
            </div>
          </div>
        </div>
      </Card>
      {transaction.state == 'Failed' && (
        <Card>
          <CardContent>
            <div className='mt-2 flex w-full justify-between'>
              <span className=''>Payment from</span>
              <span className='text-medium'>{transaction.accountTitle}</span>
            </div>
            <div className='mt-2 flex w-full justify-between'>
              <span className='text-weak'>Card</span>
              <span className='text-medium'>
                <Link
                  className='text-primary'
                  to={route('/cards/:cardId', {
                    cardId: transaction.cardTransactionDetails!.cardId // cardId will always be set for a card_transaction
                  })}
                >
                  {transaction.cardTransactionDetails?.cardMaskedPan}{' '}
                  {/* Ditto for cardMaskedPan */}
                </Link>
              </span>
            </div>
            <div className='mt-4 flex w-full justify-between font-medium'>
              <span className='text-medium'>Amount</span>
              <span className='text-medium'>{transaction.fundsReceived}</span>
            </div>
          </CardContent>
        </Card>
      )}
      {transaction.state != 'Failed' && (
        <Card>
          <Label>Payment ID</Label>
          <CardCopy
            copyContent={transaction.id}
            shareData={{
              title: 'Payment ID',
              text: transaction.id
            }}
            success='Payment ID copied to clipboard.'
            copyError="Couldn't copy to clipboard."
            shareError="Couldn't share Payment ID."
          >
            {transaction.id}
          </CardCopy>
          <CardContent>
            <div className='mt-2 flex w-full justify-between'>
              <span className='text-weak'>Payment from</span>
              <span className='text-medium'>{transaction.accountTitle}</span>
            </div>
            <div className='mt-2 flex w-full justify-between'>
              <span className='text-weak'>Card</span>
              <span className='text-medium'>
                <Link
                  className='text-primary'
                  to={route('/cards/:cardId', {
                    cardId: transaction.cardTransactionDetails!.cardId // cardId will always be set for a card_transaction
                  })}
                >
                  {transaction.cardTransactionDetails?.cardMaskedPan}{' '}
                  {/* Ditto for cardMaskedPan */}
                </Link>
              </span>
            </div>
            <div className='mt-4 flex w-full justify-between font-medium'>
              <span className='text-medium'>Amount</span>
              <span className='text-medium'>{transaction.fundsReceived}</span>
            </div>
          </CardContent>
        </Card>
      )}
    </>
  )
}
