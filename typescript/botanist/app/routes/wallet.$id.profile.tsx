import type { ActionArgs, LoaderArgs } from '@remix-run/node'

import { json } from '@remix-run/node'
import { Form, useFetcher, useLoaderData } from '@remix-run/react'
import {
  GetWalletDetails,
  GetWalletFeatures,
  SetWalletFeatures
} from '~/lib/wallet.server'
import { GridCard, Switch } from '~/components'
import { useCallback } from 'react'
import { route } from 'routes-gen'

export async function loader({ request, params }: LoaderArgs) {
  const wallet = await GetWalletDetails(request, params.id as string)
  const features = await GetWalletFeatures(request, params.id as string)

  return json({
    wallet,
    features
  })
}

export default function Page() {
  const { wallet, features } = useLoaderData<typeof loader>()

  const fetcher = useFetcher()

  const _onChangeSwitch = useCallback<{ (key: string, val: boolean): void }>(
    (key, val) => {
      fetcher.submit({ key, val: val.toString() }, { method: 'post' })
    },
    [fetcher]
  )

  return (
    <>
      <Form
        id='features-form'
        action={route('/wallet/:id/profile', { id: wallet.walletID })}
        method='post'
        className='hidden'
      />
      <GridCard
        className='col-span-full lg:col-span-4'
        title='Profile'
        options={wallet}
      />
      <div className='col-span-full flex h-max max-h-max w-full flex-col space-y-4 rounded-2xl bg-page p-4 lg:col-span-4'>
        <h2 className='font-display text-lg font-medium'>Features</h2>

        {Object.entries(features).map(([key, value]) => {
          if (key == 'walletID') return null
          return (
            <div key={key} className='flex w-full items-center justify-between'>
              <dt className='text-xs font-medium capitalize text-weak'>
                {key}
              </dt>
              <Switch
                checked={value as boolean}
                disabled={false}
                onChange={(val: any) => _onChangeSwitch(key, val)}
              />
            </div>
          )
        })}
      </div>
    </>
  )
}

export async function action({ request, params }: ActionArgs) {
  const form = await request.formData()
  const feature = form.get('key') as string
  const val = form.get('val') as string

  const currentFeatures = await GetWalletFeatures(request, params.id as string)

  await SetWalletFeatures(request, {
    ...currentFeatures,
    [feature]: val == 'true'
  })

  return null
}
