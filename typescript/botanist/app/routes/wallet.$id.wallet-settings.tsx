import type { ActionFunctionArgs, LoaderFunctionArgs } from 'react-router'
import { data, useFetcher, useLoaderData } from 'react-router'
import { useCallback } from 'react'
import { Switch } from '~/components'
import { GetWalletConfs, SetWalletConf } from '~/lib/wallet.server'

export async function loader({ request, params }: LoaderFunctionArgs) {
  const walletConfs = await GetWalletConfs(request, params.id as string)
  return data({ walletConfs })
}

export default function Page() {
  const { walletConfs } = useLoaderData<typeof loader>()
  const fetcher = useFetcher()

  const _onChangeConfSwitch = useCallback<{ (key: string, val: boolean): void }>(
    (key, val) => {
      fetcher.submit(
        { key, boolValue: val.toString() },
        { method: 'post' }
      )
    },
    [fetcher]
  )

  return (
    <div className='col-span-full flex h-max max-h-max w-full flex-col space-y-4 rounded-2xl bg-page p-4 lg:col-span-4'>
      <h2 className='font-display text-lg font-medium'>Wallet Settings</h2>

      {walletConfs.confs.map((conf) => (
        <div
          key={conf.key}
          className='flex w-full items-center justify-between'
        >
          <dt
            className='text-xs font-medium text-weak'
            title={conf.description}
          >
            {conf.displayName || conf.key}
          </dt>
          {conf.type === 'bool' ? (
            <Switch
              checked={conf.boolValue}
              disabled={false}
              onChange={(val: any) => _onChangeConfSwitch(conf.key, val)}
            />
          ) : (
            <span className='text-xs text-weak'>
              {conf.type === 'int' ? conf.intValue : conf.stringValue}
            </span>
          )}
        </div>
      ))}
    </div>
  )
}

export async function action({ request, params }: ActionFunctionArgs) {
  const form = await request.formData()
  const key = form.get('key') as string
  const boolValue = form.get('boolValue') as string

  await SetWalletConf(request, params.id as string, key, {
    boolValue: boolValue === 'true'
  })

  return null
}
