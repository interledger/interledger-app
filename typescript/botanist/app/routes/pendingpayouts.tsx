import type { LoaderFunctionArgs } from 'react-router'
import { data } from 'react-router'
import { Outlet, useLoaderData, useLocation } from 'react-router'
import clsx from 'clsx'
import { DateTime } from 'luxon'
import type { FC } from 'react'
import {
  Chip,
  ChipColor,
  LinkedInIcon,
  SlackIcon,
  TwitterIcon
} from '~/components'
import type { Payment } from '~/generated/protobuf-ts/backend/admin/v1/backend'
import { GetPendingPayouts } from '~/lib/wallet.server'

export async function loader({ request, params }: LoaderFunctionArgs) {
  const payments = await GetPendingPayouts(request)

  return data({ payments })
}

const ListItem: FC<Payment> = ({
  id,
  publicID,
  note,
  receiverIdentity,
  receiverIdentityType,
  receiverWalletUrl,
  senderAccount,
  senderAmount,
  requiredActions,
  state,
  updatedAt,
  senderWalletUrl
}) => {
  return (
    <li className='flex w-full flex-col items-center space-y-2 rounded-lg p-3'>
      <div className='flex w-full items-center justify-between'>
        <span className='text-medium'>{senderWalletUrl}</span>
        <div className='flex space-x-2'>
          {receiverIdentityType == 'twitter' && <TwitterIcon />}
          {receiverIdentityType == 'linkedin' && <LinkedInIcon />}
          {receiverIdentityType == 'slack' && <SlackIcon />}
          <span>{receiverIdentity}</span>
        </div>
      </div>
      <div className='flex w-full items-center justify-between'>
        <span className='font-medium text-medium'>{senderAmount}</span>
        <span className='text-xs text-medium'>
          {DateTime.fromSeconds(
            parseInt(updatedAt?.seconds ?? '')
          ).toLocaleString(DateTime.DATETIME_FULL)}
        </span>
        <Chip color={status == 'Complete' ? ChipColor.green : ChipColor.orange}>
          {status}
        </Chip>
      </div>
    </li>
  )
}

export default function Page() {
  const { payments } = useLoaderData<typeof loader>()
  let location = useLocation()
  return (
    <>
      <div
        className={clsx(
          location.pathname.endsWith('transactions')
            ? 'flex'
            : 'hidden lg:flex',
          'col-span-full h-max max-h-max flex-col space-y-4 rounded-2xl bg-page p-4 lg:col-span-6'
        )}
      >
        <li className='flex w-full flex-col items-center space-y-2 rounded-lg border-2 border-base p-3'>
          <div className='flex w-full items-center justify-between'>
            <span className='text-medium'>Source</span>
            <span className='text-medium'>Destination</span>
          </div>
          <div className='flex w-full items-center justify-between'>
            <span className='font-medium text-medium'>Amount</span>
            <span className='text-xs text-medium'>Date - Time</span>
            <Chip color={ChipColor.green}>Status</Chip>
          </div>
        </li>
        {payments.map((payment) => (
          <ListItem key={payment.id} {...payment} />
        ))}
      </div>
      <Outlet />
    </>
  )
}
