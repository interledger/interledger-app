import type { LoaderArgs, MetaFunction } from '@remix-run/node'
import { json } from '@remix-run/node'
import { useLoaderData } from '@remix-run/react'
import { useState } from 'react'
import { route } from 'routes-gen'
import type { ApplicationProps } from '~/components'
import {
  Card,
  CardButton,
  CardContent,
  CardHeader,
  CardLink,
  Chip,
  ChipColor,
  Dialog,
  FynbosIcon,
  Icon,
  Layouts,
  LinkedInIcon,
  TextButton,
  TwitterIcon
} from '~/components'
import { Label } from '~/components/Label'
import { getPusherArgs } from '~/lib/pusher.server'
import { usePusher } from '~/lib/usePusher'
import { getPublicWalletInfo, getTransaction } from '~/lib/wallet.server'

export async function loader({ request, params }: LoaderArgs) {
  const transaction = await getTransaction(
    request,
    params.transactionId as string
  )

  const publicWalletInfo = await getPublicWalletInfo(
    request,
    transaction.walletUrl
  )

  const pusherArgs = await getPusherArgs(request)

  return json({
    publicWalletInfo,
    transaction,
    pusherArgs
  })
}

export const handle: ApplicationProps = {
  layout: Layouts.Wallet,
  scaffold: {
    header: {
      back: route('/transactions'),
      title: 'Sent payment',
      actions: [
        {
          type: 'chip',
          content: (match) => {
            switch (match.data.transaction.status) {
              case 'Completed':
                return <Chip color={ChipColor.green}>Complete</Chip>
              case 'Pending':
                return <Chip color={ChipColor.orange}>Pending</Chip>
              case 'Failed':
                return <Chip color={ChipColor.red}>Failed</Chip>
              default:
                return null
            }
          }
        }
      ]
    },
    isNested: true
  }
}

export const meta: MetaFunction = () => {
  return {
    title: 'Transaction | Outgoing'
  }
}

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
        <CardContent>
          <span className='text-medium'>
            You are viewing public information about the person you intend to
            pay.
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
        {publicWalletInfo.identities.map((identity) => (
          <div key={identity.id} className='contents'>
            <Label className='mt-2 capitalize'>{identity.platform}</Label>
            <CardLink className='flex w-full' to={publicWalletInfo.address}>
              <div className='flex w-full items-center justify-between text-medium'>
                <div className='flex space-x-2'>
                  {identity.platform == 'twitter' && <TwitterIcon />}
                  {identity.platform == 'linkedin' && <LinkedInIcon />}
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
  return (
    <>
      <Card>
        <CardContent>
          <div className='flex items-center justify-between'>
            <h2 className='text-4xl font-medium text-error'>
              - {transaction.total}
            </h2>
            {transaction.icon === 'wallet' && <FynbosIcon height='h-12' />}
            {transaction.icon === 'twitter' && <TwitterIcon height='h-12' />}
          </div>
        </CardContent>
        <Label className='mt-2'>Payment to</Label>
        <CardButton onClick={openDialog}>
          <div className='flex w-full items-center justify-between text-medium'>
            <span>{transaction.title}</span>
            <Icon>navigate_next</Icon>
          </div>
        </CardButton>
      </Card>
      <Card>
        <CardContent>
          <div className='flex w-full justify-between'>
            <span className='text-weak'>Total fees</span>
            <span className='text-medium'>
              Free <sup>*</sup>
            </span>
          </div>
          <div className='mt-2 flex w-full justify-between'>
            <span className='text-weak'>They receive</span>
            <span className='text-medium'>{transaction.total}</span>
          </div>
          <div className='mt-4 flex w-full space-x-2'>
            <span className='text-xs text-medium'>*</span>
            <span className='text-xs text-medium'>
              For a limited time, Fynbos will absorb the fees associated with
              making a payment.
            </span>
          </div>
        </CardContent>
      </Card>
      <Card>
        <CardContent>
          <div className='flex w-full flex-col justify-between space-y-1'>
            <span className='text-weak'>Source</span>
            <span className='text-medium'>{transaction.accountTitle}</span>
          </div>
          {transaction.reference && (
            <div className='mt-4 flex w-full flex-col space-y-1'>
              <span className='text-weak'>Reference</span>
              <span className='text-medium'>{transaction.reference}</span>
            </div>
          )}
          <div className='mt-4 flex w-full flex-col justify-between space-y-1'>
            <span className='text-weak'>Date</span>
            <span className='text-medium'>
              {transaction.date} at {transaction.time}
            </span>
          </div>
          <div className='mt-4 flex w-full flex-col justify-between space-y-1'>
            <span className='text-weak'>Transaction ID</span>
            <span className='text-medium'>{transaction.id}</span>
          </div>
        </CardContent>
      </Card>
    </>
  )
}

function Incoming({ openDialog }: { openDialog: () => void }) {
  const { transaction } = useLoaderData<typeof loader>()
  return (
    <>
      <Card>
        <CardContent>
          <div className='flex items-center justify-between'>
            <h2 className='text-4xl font-medium text-strong'>
              {transaction.total}
            </h2>
            {transaction.icon === 'wallet' && <FynbosIcon height='h-12' />}
            {transaction.icon === 'twitter' && <TwitterIcon height='h-12' />}
          </div>
        </CardContent>
        <Label className='mt-2'>Payment from</Label>
        <CardButton onClick={openDialog}>
          <div className='flex w-full items-center justify-between text-medium'>
            <span>{transaction.title}</span>
            <Icon>navigate_next</Icon>
          </div>
        </CardButton>
      </Card>
      <Card>
        <CardContent>
          <div className='flex w-full flex-col justify-between space-y-1'>
            <span className='text-weak'>Paid to</span>
            <span className='text-medium'>{transaction.accountTitle}</span>
          </div>
          {transaction.reference && (
            <div className='mt-4 flex w-full flex-col space-y-1'>
              <span className='text-weak'>Their reference</span>
              <span className='text-medium'>{transaction.reference}</span>
            </div>
          )}
          <div className='mt-4 flex w-full flex-col justify-between space-y-1'>
            <span className='text-weak'>Date</span>
            <span className='text-medium'>
              {transaction.date} at {transaction.time}
            </span>
          </div>
          <div className='mt-4 flex w-full flex-col justify-between space-y-1'>
            <span className='text-weak'>Transaction ID</span>
            <span className='text-medium'>{transaction.id}</span>
          </div>
        </CardContent>
      </Card>
    </>
  )
}
