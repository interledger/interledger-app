import type { LoaderArgs } from '@remix-run/node'
import { json } from '@remix-run/node'
import { useLoaderData } from '@remix-run/react'
import {
  AnchorRouter,
  Card,
  Chip,
  ChipColor,
  Icon,
  Layouts
} from '~/components'
import { getUserSession } from '~/lib/kratos.server'
import { getTransaction, getWalletId } from '~/lib/wallet.server'
import { route } from 'routes-gen'
import {
  httpMapping,
  isGrpcError,
  openPaymentsClient,
  StatusError
} from '~/lib/proto.server'
import { usePusher } from '~/lib/pusher'

export async function loader({ request, params }: LoaderArgs) {
  const session = await getUserSession(request)
  const transaction = await getTransaction(
    request,
    params.type as string,
    params.transactionId as string
  )

  const walletID = await getWalletId(request)

  const cookie = String(request.headers.get('cookie'))
  const incomingPayment = await openPaymentsClient
    .lookupIncomingPayment(
      { id: transaction.foreignId },
      {
        meta: {
          cookies: cookie || ''
        }
      }
    )
    .then((v) => v)
    .catch(StatusError)

  if (isGrpcError(incomingPayment)) {
    throw json({}, httpMapping(incomingPayment.code))
  }

  // For now handle just incoming & outgoing payments
  let beneficiaryName = ''
  const response = await openPaymentsClient
    .getPaymentPointer({ url: incomingPayment.response.fromPaymentPointer })
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
    walletID,
    beneficiaryName,
    paymentPointer: incomingPayment.response.fromPaymentPointer,
    note: incomingPayment.response.externalRef,
    traits: session.identity.traits
  })
}

export const handle = {
  layout: Layouts.FocusLayout
}

export default function Page() {
  const { transaction, walletID, beneficiaryName, paymentPointer, note } =
    useLoaderData<typeof loader>()

  usePusher(walletID, ['transaction', 'kyc'])
  return (
    <>
      <Card>
        <div className='flex justify-between'>
          <span className='font-display text-2xl font-medium capitalize'>
            Received
          </span>
        </div>
        <Card.Item className='mt-6' variant='col'>
          <span className='text-sm'>From</span>
          <span className='text-sm text-strong'>{paymentPointer}</span>
        </Card.Item>
        {beneficiaryName != '' && (
          <Card.Item className='mt-6' variant='col'>
            <span className='text-sm'>Sender name</span>
            <span className='text-sm text-strong'>{beneficiaryName}</span>
          </Card.Item>
        )}
        <Card.Item className='mt-6'>
          <span className='text-sm'>They sent</span>
          <span className='text-sm font-medium text-strong'>
            {transaction.subTotal || '$ 0.00'}
          </span>
        </Card.Item>
        <Card.Item className='mt-2'>
          <span className='text-sm'>Total fees</span>
          <span className='text-sm font-medium text-strong'>
            {transaction.fees || '$ 0.00'}
          </span>
        </Card.Item>
        <Card.Item className='mt-2'>
          <span className='text-sm'>You received</span>
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
