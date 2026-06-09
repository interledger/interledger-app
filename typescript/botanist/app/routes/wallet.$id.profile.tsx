import type { ActionFunctionArgs, LoaderFunctionArgs } from 'react-router'

import { data, href, Form, useFetcher, useLoaderData } from 'react-router'
import { useCallback, useEffect, useState } from 'react'
import {
  Autocomplete,
  Button,
  Dialog,
  GridCard,
  Switch,
  TextButton
} from '~/components'
import {
  CheckUserTotpEnabled,
  DeleteUserTotp,
  GetWalletDetails,
  GetWalletFeatures,
  SetWalletFeatures,
  setWalletCountry,
  ListCountries
} from '~/lib/wallet.server'

export async function loader({ request, params }: LoaderFunctionArgs) {
  const wallet = await GetWalletDetails(request, params.id as string)
  const features = await GetWalletFeatures(request, params.id as string)
  const countries = await ListCountries(request)
  const identityId = wallet.users?.[0]?.id
  let hasTotpEnabled = false

  if (identityId) {
    hasTotpEnabled = await CheckUserTotpEnabled(
      request,
      identityId,
      params.id as string
    )
  }

  return data({
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
  const [isTotpEnabled, setIsTotpEnabled] = useState(hasTotpEnabled)
  const [country, setCountry] = useState<{ id: string; name: string }>()
  const [query, setQuery] = useState<string>('')
  const [filteredCountries, setFilteredCountries] = useState<
    { id: string; name: string }[]
  >([])
  const [isResetModalOpen, setIsResetModalOpen] = useState(false)

  const fetcher = useFetcher()
  const resetFetcher = useFetcher<{ success?: boolean; error?: string }>()

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

  const _onConfirmResetAuthenticator = () => {
    if (!identityId || !isTotpEnabled) return

    resetFetcher.submit(
      {
        identityId,
        walletId: wallet.walletID,
        formName: 'deleteTotp'
      },
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

  useEffect(() => {
    if (resetFetcher.state === 'idle' && resetFetcher.data?.success) {
      setIsResetModalOpen(false)
      setIsTotpEnabled(false)
    }
  }, [resetFetcher.state, resetFetcher.data])

  useEffect(() => {
    setIsTotpEnabled(hasTotpEnabled)
  }, [hasTotpEnabled])

  const isResetting = resetFetcher.state !== 'idle'

  return (
    <>
      <Form
        id='features-form'
        action={href('/wallet/:id/profile', { id: wallet.walletID })}
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
      <div className='col-span-full flex h-max max-h-max w-full flex-col space-y-4 rounded-2xl bg-page p-4 lg:col-span-4'>
        <div className='flex items-center justify-between'>
          <div className='space-y-1'>
            <h2 className='font-display text-lg font-medium'>
              Authenticator app (TOTP)
            </h2>
            <p className='text-sm text-weak'>
              {isTotpEnabled
                ? 'User has an authenticator app configured for two-factor login.'
                : 'No authenticator app is currently configured for this user.'}
            </p>
          </div>
          <span
            className={`rounded-full px-2 py-1 text-xs font-medium ${
              isTotpEnabled
                ? 'bg-green-100 text-green-800'
                : 'bg-gray-100 text-gray-600'
            }`}
          >
            {isTotpEnabled ? 'Enabled' : 'Not enabled'}
          </span>
        </div>

        {isTotpEnabled && (
          <Button
            type='button'
            className='h-10 max-w-max rounded-xl bg-red-600 px-6 text-sm hover:enabled:bg-red-500'
            onClick={() => setIsResetModalOpen(true)}
          >
            Reset authenticator
          </Button>
        )}
      </div>

      <Dialog open={isResetModalOpen} setOpen={setIsResetModalOpen}>
        <h3 className='font-display text-lg font-medium'>
          Reset authenticator for this wallet owner?
        </h3>
        <p className='text-sm text-medium'>
          This will immediately invalidate the current TOTP secret. The user
          will be prompted to set up their authenticator app again on next
          login.
        </p>
        <div className='rounded-xl border border-amber-300 bg-amber-50 p-3 text-sm text-amber-900'>
          The wallet owner will be notified by email, and this action is
          recorded in the audit log.
        </div>
        {resetFetcher.data?.error && (
          <p className='text-sm font-medium text-red-600'>
            {resetFetcher.data.error}
          </p>
        )}
        <div className='mt-2 flex justify-end space-x-3'>
          <TextButton
            type='button'
            onClick={() => setIsResetModalOpen(false)}
            disabled={isResetting}
          >
            Cancel
          </TextButton>
          <Button
            type='button'
            className='h-10 w-max rounded-xl bg-red-600 px-6 text-sm hover:enabled:bg-red-500'
            onClick={_onConfirmResetAuthenticator}
            disabled={isResetting}
          >
            {isResetting ? 'Resetting...' : 'Confirm reset'}
          </Button>
        </div>
      </Dialog>
    </>
  )
}

export async function action(args: ActionFunctionArgs) {
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

async function setWalletFeatureAction({ request, params }: ActionFunctionArgs) {
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

async function setWalletCountryAction({ request, params }: ActionFunctionArgs) {
  const form = await request.formData()
  const country = String(form.get('country') || '')
  const walletId = params.id || ''

  await setWalletCountry(request, walletId, country)

  return null
}
//
async function deleteTotpAction({ request, params }: ActionFunctionArgs) {
  const form = await request.formData()
  const walletId = form.get('walletId') as string
  const identityId = form.get('identityId') as string
  const routeWalletId = params.id as string

  if (!identityId) {
    return data(
      { success: false, error: 'Identity ID is required' },
      { status: 400 }
    )
  }

  if (!routeWalletId) {
    return data(
      { success: false, error: 'Wallet ID is required' },
      { status: 400 }
    )
  }

  if (walletId && walletId !== routeWalletId) {
    return data(
      {
        success: false,
        error: 'Wallet ID does not match the requested wallet'
      },
      { status: 400 }
    )
  }

  const wallet = await GetWalletDetails(request, routeWalletId)
  const walletIdentityIds = new Set(wallet.users.map((user) => user.id))

  if (!walletIdentityIds.has(identityId)) {
    return data(
      { success: false, error: 'Identity does not belong to this wallet' },
      { status: 400 }
    )
  }

  try {
    await DeleteUserTotp(request, identityId, routeWalletId)
    return data({ success: true })
  } catch (error) {
    const message =
      error instanceof Error
        ? error.message
        : 'Failed to reset authenticator enrollment'

    return data({ success: false, error: message }, { status: 500 })
  }
}
