import type { ActionArgs, LoaderArgs } from '@remix-run/node'

import { json } from '@remix-run/node'
import { Form, useFetcher, useLoaderData, useParams } from '@remix-run/react'
import { useCallback, useEffect, useState } from 'react'
import { route } from 'routes-gen'
import { Autocomplete, GridCard, Switch } from '~/components'
import {
  GetWalletDetails,
  GetWalletFeatures,
  SetWalletFeatures,
  setWalletCountry,
  ListCountries,
  DeleteUserTotp,
  CheckUserTotpEnabled
} from '~/lib/wallet.server'

export async function loader({ request, params }: LoaderArgs) {
  const wallet = await GetWalletDetails(request, params.id as string)
  const features = await GetWalletFeatures(request, params.id as string)
  const countries = await ListCountries(request)
  const hasTotpEnabled = wallet.users?.[0]?.id
    ? await CheckUserTotpEnabled(request, wallet.users[0].id)
    : false
  const identityId = wallet.users?.[0]?.id

  return json({
    wallet,
    features,
    hasTotpEnabled,
    countries: countries.countries.map((value) => ({
      id: value.code,
      name: value.name
    })),
    identityId
  })
}

export default function Page() {
  const { wallet, features, countries, hasTotpEnabled, identityId } =
    useLoaderData<typeof loader>()
  const [country, setCountry] = useState<{ id: string; name: string }>()
  const [query, setQuery] = useState<string>('')
  const [filteredCountries, setFilteredCountries] = useState<
    { id: string; name: string }[]
  >([])
  const params = useParams()

  const fetcher = useFetcher()

  const _onChangeFeatureSwitch = useCallback<{
    (key: string, val: boolean): void
  }>(
    (key, val) => {
      fetcher.submit(
        { key, val: val.toString(), formName: 'features' },
        { method: 'post' }
      )
    },
    [fetcher]
  )

  const _onCountryChange = (value: { id: string; name: string }) => {
    setCountry(value)
    fetcher.submit(
      { country: value.id, formName: 'country' },
      { method: 'post' }
    )
  }

  const _onDeleteTotp = () => {
    if (!identityId) return
    if (
      !confirm('Are you sure you want to delete TOTP enrollment for the selected user? He will have to re-enable TOTP.')
    )
      return

    fetcher.submit(
      { identityId: identityId, formName: 'deleteTotp' },
      { method: 'post' }
    )
  }

  useEffect(() => {
    if (query === '') setFilteredCountries(countries)
    else {
      setFilteredCountries(
        countries.filter((country) => {
          return (
            country.name
              .toLowerCase()
              .replace(/\s+/g, '')
              .includes(query.toLowerCase().replace(/\s+/g, '')) ||
            country.id
              .toLowerCase()
              .replace(/\s+/g, '')
              .includes(query.toLowerCase().replace(/\s+/g, ''))
          )
        })
      )
    }
  }, [query, countries])

  useEffect(() => {
    let walletCountry = countries.find((ctry) => ctry.id == wallet.countryCode)
    setCountry(walletCountry)
  }, [countries, wallet.countryCode])

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
                onChange={(val: any) => _onChangeFeatureSwitch(key, val)}
              />
            </div>
          )
        })}
      </div>
      <div className='col-span-full flex h-max max-h-max w-full flex-col space-y-4 rounded-2xl bg-page p-4 lg:col-span-4'>
        <h2 className='font-display text-lg font-medium'>Country</h2>

        <Autocomplete
          id='country'
          value={country}
          onChange={_onCountryChange}
          onQuery={setQuery}
          options={filteredCountries}
          className='mt-4'
          aria-invalid={Boolean(fetcher.data?.errors?.country) || undefined}
          aria-describedby={
            fetcher.data?.errors?.country ? 'country-error' : undefined
          }
          errorMessage={fetcher.data?.errors?.country}
        />
      </div>
      {hasTotpEnabled && (
        <div className='col-span-full flex h-max max-h-max w-full flex-col space-y-4 rounded-2xl bg-page p-4 lg:col-span-4'>
          <h2 className='font-display text-lg font-medium'>Security</h2>

          <div className='flex w-full items-center justify-between'>
            <dt className='text-xs font-medium text-weak'>TOTP Enrollment</dt>
            <button
              type='button'
              onClick={_onDeleteTotp}
              disabled={fetcher.state !== 'idle' || !wallet.users?.[0]?.id}
              className='rounded bg-red-600 px-3 py-1.5 text-xs font-medium text-white hover:bg-red-700 disabled:opacity-50 disabled:cursor-not-allowed'
            >
              {fetcher.state !== 'idle' ? 'Loading...' : 'Delete TOTP'}
            </button>
          </div>
          {fetcher.data?.success && (
            <p className='text-xs text-green-600'>
              TOTP enrollment deleted successfully
            </p>
          )}
          {fetcher.data?.error && (
            <p className='text-xs text-red-600'>{fetcher.data.error}</p>
          )}
        </div>
      )}
    </>
  )
}

export async function action(args: ActionArgs) {
  const formName = (await args.request.clone().formData()).get(
    'formName'
  ) as string

  if (formName == 'features') {
    return setWalletFeatureAction(args)
  }

  if (formName == 'deleteTotp') {
    return deleteTotpAction(args)
  }

  return setWalletCountryAction(args)
}

async function setWalletFeatureAction({ request, params }: ActionArgs) {
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

async function setWalletCountryAction({ request, params }: ActionArgs) {
  const form = await request.formData()
  const country = String(form.get('country') || '')
  const walletId = params.id || ''

  await setWalletCountry(request, walletId, country)

  return null
}

async function deleteTotpAction({ request, params }: ActionArgs) {
  const form = await request.formData()
  const identityId = form.get('identityId') as string

  if (!identityId) {
    return json({ error: 'Identity ID is required' }, { status: 400 })
  }

  const deleteResponse = await DeleteUserTotp(request, identityId)

  return json(deleteResponse)
}
