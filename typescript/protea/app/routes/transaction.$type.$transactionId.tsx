import type { LoaderArgs } from '@remix-run/node'
import { json } from '@remix-run/node'
import { useLoaderData, useParams } from '@remix-run/react'
import { AnchorRouter, Chip, ChipColor, Icon, Layouts } from '~/components'
import { requireUserSession } from '~/lib/kratos.server'
import { getTransaction } from '~/lib/wallet.server'
import { route } from 'routes-gen'

export async function loader({ request, params }: LoaderArgs) {
  const session = await requireUserSession(request)
  const transaction = await getTransaction(
    request,
    params.type as string,
    params.transactionId as string
  )
  return json({
    // Always go to /transactions with back button even if we've just done a payment flow
    backTo: route('/transactions'),
    transaction,
    traits: session.identity.traits
  })
}

export const handle = {
  layout: Layouts.FocusLayout
}

export default function Page() {
  const { transaction } = useLoaderData<typeof loader>()
  const params = useParams()

  return (
    <>
      <div className='flex w-full flex-col rounded-2xl bg-page p-4 pb-8'>
        <div className='flex justify-between'>
          <span className='font-display text-2xl font-medium capitalize'>
            {params.type == 'outgoing' && 'Sent'}
            {params.type == 'incoming' && 'Received'}
          </span>
          {transaction.status != 'Pending' && (
            <Chip color={ChipColor.green}>Complete</Chip>
          )}
          {transaction.status == 'Pending' && (
            <Chip color={ChipColor.yellow}>Pending</Chip>
          )}
        </div>
        <div className='mt-6 flex w-full flex-col space-y-1'>
          <span className='text-sm'>
            {params.type == 'outgoing' && 'To'}
            {params.type == 'incoming' && 'From'}
          </span>
          <span className='text-sm text-strong'>
            {transaction.paymentPointer}
          </span>
        </div>
        <div className='mt-6 flex w-full justify-between'>
          <span className='text-sm'>
            {params.type == 'outgoing' && 'You pay'}
            {params.type == 'incoming' && 'They sent'}
          </span>
          <span className='text-sm font-medium text-strong'>
            {transaction.subTotal || '$ 0.00'}
          </span>
        </div>
        <div className='mt-2 flex w-full justify-between'>
          <span className='text-sm'>Total fees</span>
          <span className='text-sm font-medium text-strong'>
            {transaction.fees || '$ 0.00'}
          </span>
        </div>
        <div className='mt-2 flex w-full justify-between'>
          <span className='text-sm'>
            {params.type == 'outgoing' && 'They receive'}
            {params.type == 'incoming' && 'You receive'}
          </span>
          <span className='text-sm text-2xl font-medium text-strong'>
            {transaction.total || '$ 0.00'}
          </span>
        </div>
        <div className='mt-6 flex w-full flex-col space-y-1'>
          <span className='text-sm'>Payment date</span>
          <span className='text-sm text-strong'>{transaction.date}</span>
        </div>
        {transaction.note && (
          <div className='mt-6 flex w-full flex-col space-y-2'>
            <span className='text-sm'>Note</span>
            <span className='text-sm text-strong'>{transaction.note}</span>
          </div>
        )}
        <div className='mt-6 flex w-full flex-col space-y-1'>
          <span className='text-sm'>Transaction ID</span>
          <span className='text-sm text-strong'>{transaction.id}</span>
        </div>
      </div>
      <div className='mt-6 flex w-full flex-col rounded-2xl bg-page p-4 pb-8'>
        <h2 className='font-display font-medium text-strong'>Support</h2>
        <span className='mt-4 text-sm'>
          Our telephone support lines are open Monday to Friday between 9am and
          5pm PST.
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
          The banking services of the Fynbos are powered by Machnet. Machnet is
          a financial technology company and not a bank. Banking services are
          provided by Machnet's partner banks who are Member FDIC. Machnet
          provides the Bank services through its banking software provider,
          Synapse. To report a complaint relating to the bank services,
          email&nbsp;
          <AnchorRouter
            to='mailto:help@synapsefi.com'
            className='text-sm text-primary'
          >
            help@synapsefi.com
          </AnchorRouter>
          .
        </span>
      </div>
    </>
  )
}
