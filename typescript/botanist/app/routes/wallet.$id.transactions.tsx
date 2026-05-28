import type { LoaderFunctionArgs } from 'react-router'
import {
  data,
  href,
  isRouteErrorResponse,
  NavLink,
  Outlet,
  useLoaderData,
  useLocation,
  useRouteError
} from 'react-router'
import { GetWalletDetails, GetWalletTransactions } from '~/lib/wallet.server'
import type { FC } from 'react'
import clsx from 'clsx'
import { Error, Chip, ChipColor } from '~/components'

export async function loader({ request, params }: LoaderFunctionArgs) {
  const wallet = await GetWalletDetails(request, params.id as string)
  const transactions = await GetWalletTransactions(request, params.id as string)
  return data({
    transactions,
    wallet
  })
}

interface Transaction {
  walletID: string
  id: string
  type: string
  status: string
  date: string
  amount: string
  source: string
  destination: string
}

const ListItem: FC<Transaction> = ({
  walletID,
  id,
  type,
  status,
  date,
  amount,
  source,
  destination
}) => {
  return (
    <NavLink
      preventScrollReset={true}
      prefetch='none'
      className='flex'
      to={href('/wallet/:id/transactions/:transactionId', {
        id: walletID,
        transactionId: id
      })}
    >
      {({ isActive }) => (
        <li
          className={`flex w-full flex-col items-center space-y-2 rounded-lg p-3 hover:bg-slate-50 ${
            isActive ? 'bg-container-hover' : 'hover:bg-container'
          }`}
        >
          <div className='flex w-full items-center justify-between'>
            <span className='text-medium'>{source}</span>
            <Chip color={ChipColor.purple}>{type}</Chip>
            <span className='text-medium'>{destination}</span>
          </div>
          <div className='flex w-full items-center justify-between'>
            <span className='font-medium text-medium'>{amount}</span>
            <span className='text-xs text-medium'>{date}</span>
            <Chip
              color={status == 'Complete' ? ChipColor.green : ChipColor.orange}
            >
              {status}
            </Chip>
          </div>
        </li>
      )}
    </NavLink>
  )
}

export default function Page() {
  const { transactions } = useLoaderData<typeof loader>()
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
            <Chip color={ChipColor.purple}>Type</Chip>
            <span className='text-medium'>Destination</span>
          </div>
          <div className='flex w-full items-center justify-between'>
            <span className='font-medium text-medium'>Amount</span>
            <span className='text-xs text-medium'>Date - Time</span>
            <Chip color={ChipColor.green}>Status</Chip>
          </div>
        </li>
        {/*<ListItem*/}
        {/*  key='{transaction.id}'*/}
        {/*  {...{*/}
        {/*    walletID: '6be0c94a-f461-48c8-822c-8ead5930342b',*/}
        {/*    id: 'transaction.id',*/}
        {/*    status: 'Complete',*/}
        {/*    type: 'Send',*/}
        {/*    date: DateTime.now().toFormat('dd MMM yyyy - HH:mm'),*/}
        {/*    amount: 'USD 100.00',*/}
        {/*    source: 'transaction.source',*/}
        {/*    destination: 'transaction.destination'*/}
        {/*  }}*/}
        {/*/>*/}
        {transactions.map((transaction) => (
          <ListItem key={transaction.id} {...transaction} />
        ))}
      </div>
      <Outlet />
    </>
  )
}

export function ErrorBoundary() {
  const error = useRouteError()

  if (isRouteErrorResponse(error)) {
    return (
      <div className='sticky top-4 col-span-full flex h-max max-h-max flex-col space-y-4 rounded-2xl bg-page p-4 lg:col-span-6'>
        <h3 className='text-5xl font-medium text-error'>{error.status}</h3>
        <h3 className='text-lg font-medium text-medium'>{error.statusText}</h3>
        <h3 className='text-lg leading-6 text-strong'>
          {JSON.stringify(error.data)}
        </h3>
      </div>
    )
  }
  return <Error data={{ title: 'error.data.message' }} />
}
