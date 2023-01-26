import type { LoaderArgs } from '@remix-run/node'

import { WalletGrid } from '~/components'
import { json } from '@remix-run/node'
import { useFetcher, useLoaderData } from '@remix-run/react'
import { ListWallets } from '~/lib/wallet.server'
import { grpcClient } from '~/lib/proto.server'

export async function loader({ request }: LoaderArgs) {
  const wallets = await ListWallets(request, {
    page: 1,
    pageSize: 10000
  })
  const signups = await grpcClient.listWaitlistSignups(
    {},
    {
      meta: {
        cookies: String(request.headers.get('cookie')) || ''
      }
    }
  )

  return json({
    wallets,
    signups: signups.response.signups
  })
}

export default function Page() {
  const { wallets, signups } = useLoaderData<typeof loader>()

  return (
    <WalletGrid>
      <div className='col-span-4 col-start-2 flex flex-col rounded-2xl bg-page p-4 pb-6'>
        <h2 className='font-display text-lg font-medium'>Total wallets</h2>
        <h1 className='mt-2 text-3xl font-medium'>{wallets.wallets.length}</h1>
      </div>
      <div className='col-span-4 flex flex-col rounded-2xl bg-page p-4 pb-6'>
        <h2 className='font-display text-lg font-medium'>People on waitlist</h2>
        <h1 className='mt-2 text-3xl font-medium'>{signups.length}</h1>
      </div>
      <div className='col-span-4 col-start-2 flex flex-col rounded-2xl bg-page p-4 pb-6'>
        <h2 className='font-display text-lg font-medium'>
          % waitlist can sign up
        </h2>
        <h1 className='mt-2 text-3xl font-medium'>
          {(
            (signups.filter((user) => user.canSignup).length / signups.length) *
            100
          ).toFixed(2)}
        </h1>
      </div>
      <div className='col-span-4 flex flex-col rounded-2xl bg-page p-4 pb-6'>
        <h2 className='font-display text-lg font-medium'>
          % waitlist beta opt in
        </h2>
        <h1 className='mt-2 text-3xl font-medium'>
          {(
            (signups.filter((user) => user.betaOptIn).length / signups.length) *
            100
          ).toFixed(2)}
        </h1>
      </div>
    </WalletGrid>
  )
}
