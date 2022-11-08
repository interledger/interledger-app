import type { LoaderArgs } from '@remix-run/node'
import { json } from '@remix-run/node'
import { useLoaderData } from '@remix-run/react'
import { AnchorRouter, ButtonRouter, Icon, Layouts } from '~/components'
import { route } from 'routes-gen'
import { requireUserSession } from '~/lib/kratos.server'

export async function loader({ request }: LoaderArgs) {
  const session = await requireUserSession(request)
  const transaction = {
    displaySendAmount: '$ 42.00',
    displayReceiveAmount: '$ 42.00',
    receivePaymentPointer: '$ 42.00',
    sendPaymentPointer: '$ 42.00',
    note: 'Some note'
  }
  // TODO fetch transaction information properly
  return json({
    transaction,
    traits: session.identity.traits
  })
}

export const handle = {
  layout: Layouts.FocusLayout
}

export default function Page() {
  const { transaction } = useLoaderData<typeof loader>()

  return (
    <>
      <div className='flex w-full flex-col rounded-2xl bg-page p-4 pb-8'>
        <h1 className='mb-6 font-display text-2xl font-medium'>Receipt</h1>
        <span>Please check the details and confirm the payment.</span>
        <div className='mt-6 flex w-full justify-between'>
          <span className='text-sm'>You pay</span>
          <span className='text-sm font-medium text-strong'>
            {transaction.displaySendAmount || '$ 0.00'}
          </span>
        </div>
        <div className='mt-2 flex w-full justify-between'>
          <span className='text-sm'>Total fees</span>
          <span className='text-sm font-medium text-strong'>$ 0.00</span>
        </div>
        <div className='mt-2 flex w-full justify-between'>
          <span className='text-sm'>They receive</span>
          <span className='text-sm text-2xl font-medium text-strong'>
            {transaction.displayReceiveAmount || '$ 0.00'}
          </span>
        </div>
        {transaction.note && (
          <div className='mt-8 flex w-full flex-col space-y-2'>
            <span className='text-sm'>Note</span>
            <span className='text-sm text-strong'>{transaction.note}</span>
          </div>
        )}

        <div className='mt-8 flex w-full justify-between'>
          <span className='text-sm'>To</span>
          <span className='text-sm font-medium text-strong'>
            {transaction.receivePaymentPointer}
          </span>
        </div>

        <div className='mt-8 flex w-full justify-between'>
          <span className='text-sm'>From</span>
          <span className='text-sm font-medium text-strong'>
            {transaction.sendPaymentPointer}
          </span>
        </div>
        <div className='mt-2 flex w-full justify-end'>
          <span className='text-sm font-medium text-strong'>
            TODO: card information
          </span>
        </div>

        <div className='mt-6'>
          <ButtonRouter to={route('/')}>Close</ButtonRouter>
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
