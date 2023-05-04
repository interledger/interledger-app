import type { LoaderArgs } from '@remix-run/node'
import { json } from '@remix-run/node'
import { NavLink, Outlet, useLoaderData, useLocation } from '@remix-run/react'
import { GetWalletDetails, GetWalletTransactions } from '~/lib/wallet.server'
import { DateTime } from 'luxon'
import { FC } from 'react'
import { route } from 'routes-gen'
import clsx from 'clsx'
import { Chip, ChipColor } from '~/components/Chip'

export async function loader({ request, params }: LoaderArgs) {
  const wallet = await GetWalletDetails(request, params.id as string)
  const transactions = await GetWalletTransactions(request, params.id as string)
  return json({
    transactions,
    wallet: {
      ...wallet,
      dateOfBirth: DateTime.fromSeconds(
        parseInt(wallet.dateOfBirth?.seconds ?? '')
      ).toFormat('dd MMM yyyy')
    }
  })
}

type TabItemProps = {
  to: string
  title: string
  date: string
  amount: string
  status: string
}

const ListItem: FC<TabItemProps> = ({ title, date, amount, status, to }) => {
  return (
    <NavLink preventScrollReset={true} prefetch='none' className='flex' to={to}>
      {({ isActive }) => (
        <li
          className={`flex w-full justify-between hover:bg-slate-50 items-center rounded-lg p-3 ${
            isActive && 'bg-slate-100'
          }`}
        >
          <div className='flex flex-col space-y-1'>
            <dt className='text-medium'>{title}</dt>
            <dd className='text-xs text-medium'>{date || '-'}</dd>
          </div>
          <div className='flex flex-col items-end space-y-1'>
            <dt className='text-medium font-medium'>{amount}</dt>
            <Chip color={ChipColor.green}>{status}</Chip>
          </div>
        </li>
      )}
    </NavLink>
  )
}

export default function Page() {
  const { wallet, transactions } = useLoaderData<typeof loader>()
  let location = useLocation()
  return (
    <>
      <dl
        className={clsx(
          location.pathname.endsWith('transactions')
            ? 'flex'
            : 'hidden lg:flex',
          'col-span-full max-h-max h-max lg:col-span-6 flex-col space-y-4 rounded-2xl bg-page p-4'
        )}
      >
        <ListItem
          to={route('/wallet/:id/transactions/:transactionId', {
            id: wallet.walletID,
            transactionId: 'asd'
          })}
          title='First name'
          date='2020 Friday somethisn'
          amount='- 5.00 USD'
          status='Complete'
        />
        {transactions.map((transaction) => (
          <ListItem
            to={route('/wallet/:id/transactions/:transactionId', {
              id: wallet.walletID,
              transactionId: transaction.id
            })}
            key={transaction.id}
            title={transaction.source}
            date={transaction.date}
            amount={transaction.amount}
            status={transaction.status}
          />
        ))}
      </dl>
      <Outlet />
    </>
  )
}
