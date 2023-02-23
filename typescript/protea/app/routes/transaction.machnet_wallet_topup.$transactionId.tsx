import type { LoaderArgs, MetaFunction } from '@remix-run/node'
import { json } from '@remix-run/node'
import { useLoaderData } from '@remix-run/react'
import {
  AnchorRouter,
  Chip,
  Card,
  ChipColor,
  Icon,
  Layouts
} from '~/components'
import { getLinkedAccount, getTransaction } from '~/lib/wallet.server'
import { route } from 'routes-gen'
import { usePusher } from '~/lib/usePusher'
import { getPusherArgs } from '~/lib/pusher.server'

export async function loader({ request, params }: LoaderArgs) {
  const transaction = await getTransaction(
    request,
    params.type as string,
    params.transactionId as string
  )

  const pusherArgs = await getPusherArgs(request)

  let linkedAccountName = ''
  if (transaction.transfers.length > 0) {
    const linkedAccountId = transaction.transfers.filter(
      (trf) => trf.type == 'debit_card'
    )[0]?.linkedAccountId
    if (linkedAccountId) {
      const linkedAccount = await getLinkedAccount(request, linkedAccountId)
      linkedAccountName = linkedAccount.name
    }
  }

  return json({
    // Always go to /transactions with back button even if we've just done a payment flow
    backTo: route('/transactions'),
    transaction,
    pusherArgs,
    linkedAccountName
  })
}

export const handle = {
  layout: Layouts.FocusLayout
}

export const meta: MetaFunction = () => {
  return {
    title: 'Transaction | Top up'
  }
}

export default function Page() {
  const { transaction, pusherArgs, linkedAccountName } =
    useLoaderData<typeof loader>()

  usePusher(pusherArgs, ['transaction', 'kyc'])

  return (
    <>
      <Card>
        <div className='flex justify-between'>
          <span className='font-display text-2xl font-medium capitalize'>
            Top up
          </span>
        </div>
        <Card.Item variant='col' className='mt-8'>
          <span className='text-sm text-medium'>Top up from</span>
          <span className='text-sm text-strong'>{linkedAccountName}</span>
        </Card.Item>
        <Card.Item className='mt-2'>
          <span className='text-sm text-medium'>To</span>
          <span className='text-sm text-strong'>Cash balance</span>
        </Card.Item>
        <Card.Item className='mt-6'>
          <span className='text-sm text-medium'>Top up amount</span>
          <span className='text-sm font-medium text-strong'>
            {transaction.subTotal || '$ 0.00'}
          </span>
        </Card.Item>
        <Card.Item className='mt-2'>
          <span className='text-sm text-medium'>Total fees</span>
          <span className='text-sm font-medium text-strong'>
            {transaction.fees || '$ 0.00'}
          </span>
        </Card.Item>
        <Card.Item className='mt-2'>
          <span className='text-sm text-medium'>You receive</span>
          <span className='text-sm text-2xl font-medium text-strong'>
            {transaction.total || '$ 0.00'}
          </span>
        </Card.Item>
      </Card>
      <Card className='mt-6'>
        <Card.Item>
          <span className='text-sm text-medium'>Payment date</span>
          <span className='text-sm text-strong'>
            {transaction.date || 'Pending'}
          </span>
        </Card.Item>
        <Card.Item className='mt-4 items-center'>
          <span className='text-sm text-medium'>Status</span>
          {transaction.status == 'Completed' && (
            <Chip color={ChipColor.green}>Complete</Chip>
          )}
          {transaction.status == 'Pending' && (
            <Chip color={ChipColor.yellow}>Pending</Chip>
          )}
          {transaction.status == 'Failed' && (
            <Chip color={ChipColor.orange}>Failed</Chip>
          )}
        </Card.Item>
      </Card>
      <Card className='mt-6'>
        <Card.Item variant='col'>
          <span className='text-sm text-medium'>Transaction ID</span>
          <span className='text-sm text-strong'>{transaction.id}</span>
        </Card.Item>
      </Card>
      <Card className='mt-6'>
        <h2 className='font-display font-medium text-strong'>Support</h2>
        <span className='mt-4 text-sm'>
          If you have any questions, issues, or complaints, please first contact
          Fynbos at:
        </span>
        <div className='mt-3 flex items-center space-x-2 text-medium'>
          <Icon>call</Icon>
          <AnchorRouter
            to='tel:+1 (856) 249-3067'
            className='text-sm text-primary'
          >
            +1 (856) 249-3067
          </AnchorRouter>
        </div>
        <div className='mt-2 flex items-center space-x-2 text-medium'>
          <Icon>mail</Icon>
          <AnchorRouter
            to='mailto:support@fynbos.app'
            className='text-sm text-primary'
          >
            support@fynbos.app
          </AnchorRouter>
        </div>
        <span className='mt-4 text-sm'>
          Our contact hours are Monday to Friday between 9am and 5pm EST.
        </span>
        <span className='mt-4 text-sm'>
          In case your grievances are not addressed by Fynbos or for any
          escalation purposes, please contact Machnet either through email&nbsp;
          <AnchorRouter
            to='mailto:help@machnetinc.com'
            className='text-sm text-primary'
          >
            help@machnetinc.com
          </AnchorRouter>
          &nbsp;or call{' '}
          <AnchorRouter
            to='tel:+1 (408) 539-6455'
            className='text-sm text-primary'
          >
            +1 (408) 539-6455
          </AnchorRouter>
          .
        </span>
        <span className='mt-4 text-sm'>
          Bank services are provided by Evolve Bank & Trust, Member FDIC,
          through our banking software provider, SynapseFI. To report a
          complaint relating to the bank services, email&nbsp;
          <AnchorRouter
            to='mailto:help@synapsefi.com'
            className='text-sm text-primary'
          >
            help@synapsefi.com
          </AnchorRouter>
          .
        </span>
      </Card>
    </>
  )
}
