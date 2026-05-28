import type { ActionFunctionArgs, LoaderFunctionArgs } from 'react-router'
import { data, href, Form, useFetcher, useLoaderData, useParams } from 'react-router'
import { useCallback } from 'react'
import { GridCard, Switch } from '~/components'
import {
  EnablGatehubBalance,
  EnablePtiWalletBalance,
  EnableXagoWalletBalance,
  GetGatehubWalletBalance,
  GetPtiWalletBalance,
  GetXagoWalletBalance
} from '~/lib/wallet.server'

export async function loader({ request, params }: LoaderFunctionArgs) {
  const xagoBalance = await GetXagoWalletBalance(request, params.id as string)
  const ptiBalance = await GetPtiWalletBalance(request, params.id as string)
  const gatehubBalance = await GetGatehubWalletBalance(
    request,
    params.id as string
  )

  return data({
    zarBalance: xagoBalance,
    usdBalance: ptiBalance,
    eurBalance: gatehubBalance
  })
}

export default function Page() {
  const { zarBalance, usdBalance, eurBalance } = useLoaderData<typeof loader>()
  const { id } = useParams()
  const fetcher = useFetcher()

  const _onChangeSwitch = useCallback<{ (currency: string): void }>(
    (currency: string) => {
      fetcher.submit({ currency }, { method: 'post' })
    },
    [fetcher]
  )

  return (
    <>
      {zarBalance && (
        <GridCard
          className='col-span-full lg:col-span-4'
          title='ZAR Balance'
          options={zarBalance}
        />
      )}
      <Form
        id='features-form'
        action={href('/wallet/:id/balances', { id: id as string })}
        method='post'
        className='hidden'
      />
      <div className='flex w-full items-center justify-between'>
        <dt className='text-xs font-medium capitalize text-weak'>
          Enable ZAR Balance
        </dt>
        <Switch
          checked={!!zarBalance}
          disabled={!!zarBalance}
          onChange={() => _onChangeSwitch('ZAR')}
        />
      </div>
      {usdBalance && (
        <GridCard
          className='col-span-full lg:col-span-4'
          title='USD Balance'
          options={usdBalance}
        />
      )}
      <div className='flex w-full items-center justify-between'>
        <dt className='text-xs font-medium capitalize text-weak'>
          Enable USD Balance
        </dt>
        <Switch
          checked={!!usdBalance}
          disabled={!!usdBalance}
          onChange={() => _onChangeSwitch('USD')}
        />
      </div>
      {eurBalance && (
        <GridCard
          className='col-span-full lg:col-span-4'
          title='EUR Balance'
          options={eurBalance}
        />
      )}
      <div className='flex w-full items-center justify-between'>
        <dt className='text-xs font-medium capitalize text-weak'>
          Enable EUR Balance
        </dt>
        <Switch
          checked={!!eurBalance}
          disabled={!!eurBalance}
          onChange={() => _onChangeSwitch('EUR')}
        />
      </div>
    </>
  )
}

export async function action({ request, params }: ActionFunctionArgs) {
  let form = await request.formData()
  let currency = form.get('currency')

  if (currency == 'ZAR') {
    await EnableXagoWalletBalance(request, params.id as string)
  } else if (currency == 'USD') {
    await EnablePtiWalletBalance(request, params.id as string)
  } else if (currency == 'EUR') {
    await EnablGatehubBalance(request, params.id as string)
  }

  return null
}
