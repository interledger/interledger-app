import type { LoaderFunctionArgs, MetaFunction } from '@remix-run/node'
import { json } from '@remix-run/node'
import { useLoaderData } from '@remix-run/react'
import { route } from 'routes-gen'
import type { ApplicationProps } from '~/components'
import {
  Card,
  CardButton,
  CardContent,
  CardHeader,
  Chip,
  ChipColor,
  Icon,
  Layouts
} from '~/components'
import { Label } from '~/components/Label'
import { mergeMeta } from '~/lib/meta'
import { mockPendingConfirmations } from '~/lib/mocks/confirmations'

export async function loader({ params }: LoaderFunctionArgs) {
  // TODO: Replace with actual API call
  const confirmation = mockPendingConfirmations.find(
    (c) => c.transactionId === params.confirmationId
  )

  if (!confirmation) {
    throw new Response('Not Found', { status: 404 })
  }

  const formattedDate = new Date(
    Number(confirmation.purchaseDate)
  ).toLocaleDateString('en-US', {
    month: 'long',
    day: 'numeric',
    year: 'numeric'
  })

  const formattedTime = new Date(
    Number(confirmation.purchaseDate)
  ).toLocaleTimeString('en-US', {
    hour: 'numeric',
    minute: '2-digit',
    hour12: true
  })

  const formattedAmount = `${parseFloat(confirmation.purchaseAmount).toFixed(
    2
  )} ${confirmation.purchaseCurrency}`

  return json({
    confirmation: {
      ...confirmation,
      formattedDate,
      formattedTime,
      formattedAmount
    }
  })
}

export const handle: ApplicationProps = {
  layout: Layouts.Wallet,
  scaffold: {
    header: {
      back: route('/confirmations'),
      title: 'Confirmation',
      actions: () => ({
        key: 'Pending',
        nodes: <Chip color={ChipColor.orange}>Pending</Chip>
      })
    },
    isNested: true
  }
}

export const meta: MetaFunction<typeof loader> = mergeMeta(({ data }) => [
  {
    title:
      typeof data === 'undefined'
        ? 'Confirmation'
        : `${data.confirmation.formattedAmount} at ${data.confirmation.merchantName}`
  }
])

export default function Page() {
  const { confirmation } = useLoaderData<typeof loader>()

  return (
    <>
      <Card>
        <CardContent>
          <div className='flex items-center justify-between'>
            <h2 className='text-4xl font-medium text-error'>
              {confirmation.formattedAmount}
            </h2>
            <div className='flex flex-col items-end space-y-1'>
              <span className='text-sm font-medium text-medium'>
                {confirmation.formattedDate}
              </span>
              <span className='text-xs text-weak'>
                {confirmation.formattedTime}
              </span>
            </div>
          </div>
        </CardContent>
        <Label className='mt-2'>Merchant</Label>
        <div className='my-1 flex space-x-2 rounded-xl bg-nav p-3'>
          <div className='flex w-full items-center justify-between text-medium'>
            <div className='flex space-x-2'>
              <Icon>store</Icon>
              <span>{confirmation.merchantName}</span>
            </div>
          </div>
        </div>
      </Card>

      <Card>
        <CardHeader>
          <h3 className='text-lg font-medium'>Confirm Purchase</h3>
        </CardHeader>
        <CardContent>
          <span className='text-sm text-medium'>
            Approve or decline this transaction within {confirmation.timeout}{' '}
            seconds.
          </span>
        </CardContent>
        <div className='flex w-full space-x-4'>
          <CardButton className='flex-1 bg-error text-white hover:bg-error/90'>
            <span className='font-medium'>Decline</span>
          </CardButton>
          <CardButton className='bg-success hover:bg-success/90 flex-1 text-white'>
            <span className='font-medium'>Approve</span>
          </CardButton>
        </div>
      </Card>

      <Card>
        <Label>Transaction ID</Label>
        <div className='my-1 flex rounded-xl bg-nav p-3'>
          <span className='break-all text-sm text-medium'>
            {confirmation.transactionId}
          </span>
        </div>
        <CardContent>
          <div className='mt-2 flex w-full justify-between'>
            <span className='text-weak'>Purchase amount</span>
            <span className='text-medium'>{confirmation.formattedAmount}</span>
          </div>
          <div className='mt-2 flex w-full justify-between'>
            <span className='text-weak'>Currency</span>
            <span className='text-medium'>{confirmation.purchaseCurrency}</span>
          </div>
        </CardContent>
      </Card>
    </>
  )
}
