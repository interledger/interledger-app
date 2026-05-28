import type { LoaderFunctionArgs } from 'react-router'

import { Grid } from '~/components'
import { data } from 'react-router'
import { useLoaderData } from 'react-router'
import { ListWallets } from '~/lib/wallet.server'
import { grpcClient } from '~/lib/proto.server'

export async function loader({ request }: LoaderFunctionArgs) {
  try {
    const wallets = await ListWallets(request, {
      pageSize: 10000
    })

    console.log('test get wallets :', wallets)
    const signups = await grpcClient.listWaitlistSignups(
      {},
      {
        meta: {
          cookies: String(request.headers.get('cookie')) || ''
        }
      }
    )
    console.log('test signups :', signups)
    return data({
      error: '',
      wallets,
      signups: signups.response.signups
    })
  } catch (e) {
    console.log('Failed to retrieve wallets: ', e)
    return data({
      error: 'there was an error retrieving wallets',
      wallets: { wallets: [] },
      signups: []
    })
  }
}

export default function Page() {
  const { wallets, signups, error } = useLoaderData<typeof loader>()

  return (
    <Grid>
      {!error ? (
        <>
          <div className='col-span-2 flex flex-col rounded-2xl bg-page p-4'>
            <h2 className='font-display text-lg font-medium'>Total wallets</h2>
            <h1 className='mt-2 text-3xl font-medium'>
              {wallets.wallets.length}
            </h1>
          </div>
          <div className='col-span-2 flex flex-col rounded-2xl bg-page p-4'>
            <h2 className='font-display text-lg font-medium'>
              People on waitlist
            </h2>
            <h1 className='mt-2 text-3xl font-medium'>{signups.length}</h1>
          </div>
          <div className='col-span-2 col-start-1 flex flex-col rounded-2xl bg-page p-4'>
            <h2 className='font-display text-lg font-medium'>
              % waitlist can sign up
            </h2>
            <h1 className='mt-2 text-3xl font-medium'>
              {(
                (signups.filter((user) => user.canSignup).length /
                  signups.length) *
                100
              ).toFixed(2)}
            </h1>
          </div>
          <div className='col-span-2 flex flex-col rounded-2xl bg-page p-4'>
            <h2 className='font-display text-lg font-medium'>
              % waitlist beta opt in
            </h2>
            <h1 className='mt-2 text-3xl font-medium'>
              {(
                (signups.filter((user) => user.betaOptIn).length /
                  signups.length) *
                100
              ).toFixed(2)}
            </h1>
          </div>
        </>
      ) : (
        <div className='col-span-4 flex flex-col rounded-2xl bg-page p-4'>
          <h2 className='font-display text-lg font-medium'>
            There was an error retrieving wallets
          </h2>
        </div>
      )}
    </Grid>
  )
}
