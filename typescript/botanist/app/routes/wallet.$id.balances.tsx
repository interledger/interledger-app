import type {LoaderArgs} from '@remix-run/node'
import {ActionArgs, json} from '@remix-run/node'
import {Form, useFetcher, useLoaderData, useParams} from '@remix-run/react'
import {
  EnableWalletBalance,
  GetWalletBalance,
} from '~/lib/wallet.server'
import {GridCard, Switch} from '~/components'
import {route} from "routes-gen";
import {useCallback} from "react";

export async function loader({request, params}: LoaderArgs) {
  const balance = await GetWalletBalance(request, params.id as string)

  return json({
    balance,
  })
}

export default function Page() {
  const {balance} = useLoaderData<typeof loader>()
  const {id} = useParams()
  const fetcher = useFetcher()

  const _onChangeSwitch = useCallback<{ (): void }>(
    () => {
      fetcher.submit({}, {method: 'post'})
    },
    [fetcher]
  )

  return (
    <>
      {
        balance &&
          <GridCard
              className='col-span-full lg:col-span-4'
              title="Balance"
              options={balance}
          />
      }
      <Form
        id='features-form'
        action={route('/wallet/:id/balances', {id: id as string})}
        method='post'
        className='hidden'
      />
      <div className='flex w-full items-center justify-between'>
        <dt className='text-xs font-medium capitalize text-weak'>
          Enable ZAR Balance
        </dt>
        <Switch
          checked={!!balance}
          disabled={!!balance}
          onChange={() => _onChangeSwitch()}
        />
      </div>
    </>
  )
}

export async function action({request, params}: ActionArgs) {
  await EnableWalletBalance(request, params.id as string)

  return null
}