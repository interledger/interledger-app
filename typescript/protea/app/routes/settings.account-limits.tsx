import type { LoaderArgs } from '@remix-run/node'
import { json } from '@remix-run/node'
import { useLoaderData } from '@remix-run/react'
import { AnchorRouter, Card, Layouts, WalletGrid } from '~/components'
import { getAccountLimits } from '~/lib/wallet.server'
import type { FC } from 'react'

export async function loader({ request }: LoaderArgs) {
  let limits = await getAccountLimits(request)
  return json({
    limits
  })
}

export const handle = {
  layout: Layouts.WalletLayout
}

export default function Page() {
  const { limits } = useLoaderData<typeof loader>()
  return (
    <WalletGrid>
      <Card className='col-span-full sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <h1 className='font-display text-2xl font-medium'>Account limits</h1>
        <span className='text-sm text-medium mt-4'>
          To increase your account limits, please contact{' '}
          <AnchorRouter className='text-primary' to='mailto:support@fynbos.dev'>
            support@fynbos.dev
          </AnchorRouter>
        </span>
        <h2 className='font-display font-medium text-sm mt-6 -mb-1'>
          Transfers
        </h2>
        <Limit title='Daily' amount={limits.transfer?.daily} />
        <Limit title='Monthly' amount={limits.transfer?.monthly} />
        <Limit title='Annually' amount={limits.transfer?.annual} />
      </Card>
      <Card className='col-span-full sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <h2 className='font-display font-medium text-sm -mb-1'>Top ups</h2>
        <Limit title='Daily' amount={limits.fundWallet?.daily} />
        <Limit title='Monthly' amount={limits.fundWallet?.monthly} />
        <Limit title='Annually' amount={limits.fundWallet?.annual} />
      </Card>
      <Card className='col-span-full sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <h2 className='font-display font-medium text-sm -mb-1'>Withdrawals</h2>
        <Limit title='Daily' amount={limits.withdrawal?.daily} />
        <Limit title='Monthly' amount={limits.withdrawal?.monthly} />
        <Limit title='Annually' amount={limits.withdrawal?.annual} />
      </Card>
      <Card className='col-span-full sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <h2 className='font-display font-medium text-sm'>Cash balance</h2>
        <span className='text-sm text-medium mt-2'>
          You are permitted to hold a maximum of{' '}
          {limits.fundWallet?.walletHold?.total.replace(' ', '\u00a0')} in your
          cash balance at any time.
        </span>
      </Card>
    </WalletGrid>
  )
}

type LimitProps = {
  title: string
  amount?: LimitAmount
}

type LimitAmount = {
  remaining?: string
  total?: string
  percentage?: number
}
const Limit: FC<LimitProps> = ({ title, amount }) => {
  return (
    <div className='flex flex-col mt-3'>
      <div className='flex justify-between'>
        <span className='text-medium text-sm'>{title}</span>
        <div className='flex space-x-1 items-center'>
          <span className='text-purple-500 text-xs font-medium'>
            {amount?.remaining} remaining
          </span>
          <span className='text-disabled text-sm'>/</span>
          <span className='text-medium text-sm'>{amount?.total}</span>
        </div>
      </div>
      <div className='mt-1 w-full bg-slate-200 h-1'>
        <div
          style={{ width: `${amount?.percentage}%` }}
          className='h-full bg-purple-500'
        />
      </div>
    </div>
  )
}
