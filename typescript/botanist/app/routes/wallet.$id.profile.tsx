import type { LoaderArgs } from '@remix-run/node'

import { Grid } from '~/components'
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

function ListItem({ title, body }: { title: string; body: string }) {
  return (
    <div className='flex flex-col space-y-1'>
      <dt className='text-sm font-medium text-weak'>{title}</dt>
      <dd className='text-strong'>{body || '-'}</dd>
    </div>
  )
}

export default function Page() {
  const { wallet } = useLoaderData<typeof loader>()

  return (
    <>
      <dl className='col-span-full flex lg:col-span-6 flex-col space-y-4 rounded-2xl bg-page p-4 pb-6'>
        <ListItem title='First name' body={wallet.firstName} />
        <ListItem title='Last name' body={wallet.lastName} />
        <ListItem title='Email' body={wallet.users[0].email} />
        <ListItem title='Phone number' body={wallet.users[0].phoneNumber} />
        <ListItem title='Country' body={wallet.countryCode} />
        <ListItem title='KYC status' body={wallet.countryCode} />
      </dl>
    </>
  )
}
