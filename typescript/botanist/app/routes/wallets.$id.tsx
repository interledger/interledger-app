import type { LoaderArgs } from '@remix-run/node'

import { WalletGrid } from '~/components'
import { json } from '@remix-run/node'
import { useLoaderData } from '@remix-run/react'
import { GetWalletDetails } from '~/lib/wallet.server'
import { DateTime } from 'luxon'

export async function loader({ request, params }: LoaderArgs) {
  const wallet = await GetWalletDetails(request, params.id as string)

  return json({
    // TODO: Refactor formatting into wallet.server
    wallet: {
      ...wallet,
      gender:
        wallet.gender == 0
          ? 'Unknown'
          : wallet.gender == 1
          ? 'Male'
          : wallet.gender == 2
          ? 'Female'
          : 'Other',
      dateOfBirth: DateTime.fromSeconds(
        parseInt(wallet.dateOfBirth?.seconds ?? '')
      ).toFormat('dd MMM yyyy')
    }
  })
}

/**
 * RPC endpoints required for the admin panel
 *
 * - Transactions
 *     - List
 *     - Get detailed
 * - Identities
 *     - List
 *     - Get Detailed
 * - KYC Status
 * - Linked Accounts
 *     - List
 *     - Detailed
 * - Audit
 *     - List
 */

function ListItem({ title, body }: { title: string; body: string }) {
  return (
    <div className='py-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:py-5 sm:px-6'>
      <dt className='text-sm font-medium text-gray-500'>{title}</dt>
      <dd className='mt-1 text-sm text-gray-900 sm:col-span-2 sm:mt-0'>
        {body}
      </dd>
    </div>
  )
}

export default function Page() {
  const { wallet } = useLoaderData<typeof loader>()

  return (
    <WalletGrid>
      <div className='col-span-full lg:col-span-6 flex flex-col rounded-2xl bg-page p-4 pb-6'>
        <div>
          <h3 className='text-lg font-medium leading-6 text-gray-900'>
            Wallet profile
          </h3>
          <p className='mt-1 mb-5 max-w-2xl text-sm text-gray-500'>
            Personal details and wallet information of a specified wallet.
          </p>
        </div>
        <div className='border-t border-gray-200'>
          <dl className='sm:divide-y sm:divide-gray-200'>
            <ListItem title='Wallet ID' body={wallet.walletID} />
            <ListItem
              title='Full name'
              body={wallet.firstName + ' ' + wallet.lastName}
            />
            <ListItem title='Email' body={wallet.users[0].email} />
            <ListItem title='Phone number' body={wallet.users[0].phoneNumber} />
            <ListItem title='Country' body={wallet.countryCode} />
          </dl>
        </div>
      </div>
    </WalletGrid>
  )
}
