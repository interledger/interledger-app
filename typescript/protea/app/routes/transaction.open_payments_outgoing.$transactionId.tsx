import type { LoaderArgs } from '@remix-run/node'
import { json } from '@remix-run/node'
import { useLoaderData } from '@remix-run/react'
import { AnchorRouter, Chip, ChipColor, Icon, Layouts } from '~/components'
import { requireUserSession } from '~/lib/kratos.server'
import { getTransaction } from '~/lib/wallet.server'
import { route } from 'routes-gen'
import {
  isGrpcError,
  openPaymentsClient,
  StatusError
} from '~/lib/proto.server'

export async function loader({ request, params }: LoaderArgs) {
  const session = await requireUserSession(request)
  const transaction = await getTransaction(
    request,
    params.type as string,
    params.transactionId as string
  )

  const cookie = String(request.headers.get('cookie'))
  const outgoingPayment = await openPaymentsClient
    .lookupOutgoingPayment(
      { id: transaction.foreignId },
      {
        meta: {
          cookies: cookie || ''
        }
      }
    )
    .then((v) => v)
    .catch(StatusError)

  if (isGrpcError(outgoingPayment)) {
    throw new Error('')
  }

  // For now handle just incoming & outgoing payments
  let beneficiaryName = ''
  const response = await openPaymentsClient
    .getPaymentPointer({ url: outgoingPayment.response.toPaymentPointer })
    .then((v) => v)
    .catch(StatusError)

  // Silently fail if it can't be found for now
  if (!isGrpcError(response)) {
    beneficiaryName = response.response.legalName
  }

  return json({
    // Always go to /transactions with back button even if we've just done a payment flow
    backTo: route('/transactions'),
    transaction,
    beneficiaryName,
    paymentPointer: outgoingPayment.response.toPaymentPointer,
    note: outgoingPayment.response.description,
    traits: session.identity.traits
  })
}

export const handle = {
  layout: Layouts.FocusLayout
}

export default function Page() {
  const { transaction, beneficiaryName, paymentPointer, note } =
    useLoaderData<typeof loader>()

  return (
    <>
      <div className='flex w-full flex-col rounded-2xl bg-page p-4 pb-8'>
        <div className='flex justify-between'>
          <span className='font-display text-2xl font-medium capitalize'>
            Sent
          </span>
          {transaction.status != 'Pending' && (
            <Chip color={ChipColor.green}>Complete</Chip>
          )}
          {transaction.status == 'Pending' && (
            <Chip color={ChipColor.yellow}>Pending</Chip>
          )}
        </div>
        <div className='mt-6 flex w-full flex-col space-y-1'>
          <span className='text-sm'>To</span>
          <span className='text-sm text-strong'>{paymentPointer}</span>
        </div>
        {beneficiaryName != '' && (
          <div className='mt-6 flex w-full flex-col space-y-1'>
            <span className='text-sm'>Beneficiary name</span>
            <span className='text-sm text-strong'>{beneficiaryName}</span>
          </div>
        )}
        <div className='mt-6 flex w-full justify-between'>
          <span className='text-sm'>You pay</span>
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
          <span className='text-sm'>They receive</span>
          <span className='text-sm text-2xl font-medium text-strong'>
            {transaction.total || '$ 0.00'}
          </span>
        </div>
        <div className='mt-6 flex w-full flex-col space-y-1'>
          <span className='text-sm'>Payment date</span>
          <span className='text-sm text-strong'>{transaction.date}</span>
        </div>
        {note && (
          <div className='mt-6 flex w-full flex-col space-y-2'>
            <span className='text-sm'>Note</span>
            <span className='text-sm text-strong'>{note}</span>
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
      </div>
    </>
  )
}
