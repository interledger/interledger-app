import type { LoaderArgs, MetaFunction } from '@remix-run/node'
import { json, LinksFunction } from '@remix-run/node'
import { useLoaderData } from '@remix-run/react'
import { useEffect, useState } from 'react'

import styles from '~/styles/VGS.css'
import { Button, Card, Layouts, Shape } from '~/components'
import { flowType, requireFlow, updateFlow } from '~/lib/flows.server'
import { getLinkedAccounts, getWalletId } from '~/lib/wallet.server'
import { loadVGSCollect } from '@vgs/collect-js'
import {
  ICollectFormPayloadStructure,
  VGSCollectForm,
  VGSCollectFormState,
  VGSCollectVaultEnvironment
} from '@vgs/collect-js-react'

export async function loader({ request, params }: LoaderArgs) {
  const linkedAccounts = await getLinkedAccounts(request)
  await requireFlow(request, flowType.LinkCardAccount)
  await updateFlow(request, flowType.LinkCardAccount, {
    linkedAccountLength: linkedAccounts.linkedAccounts.length
  })

  const walletId = await getWalletId(request)

  // TODO: RPC endpoint to get VGS token

  return json({
    walletId,
    vaultId: process.env.VGS_VAULT_ID || 'tntqofj9utu',
    environment: process.env.VGS_ENVIRONMENT || 'sandbox',
    version: process.env.VGS_COLLECT_VERSION || '2.18.3'
  })
}

export const handle = {
  layout: Layouts.FocusLayout
}

export const meta: MetaFunction = () => {
  return {
    title: 'Add linked account'
  }
}

export const links: LinksFunction = () => {
  return [{ rel: 'stylesheet', href: styles }]
}

class VGSCollectHttpStatusCode {}

export default function Page() {
  const { vaultId, environment, version, walletId } =
    useLoaderData<typeof loader>()
  const [isVGSCollectScriptLoaded, setCollectScriptLoaded] = useState(false)
  const [fieldErrors, setFieldErrors] = useState({
    cardNumber: '',
    expDate: '',
    cardSecurityCode: ''
  })
  const [cardInfo, setCardInfo] = useState({
    last4: '',
    cardType: ''
  })

  useEffect(() => {
    loadVGSCollect({
      vaultId,
      environment,
      version
    }).then(() => {
      setCollectScriptLoaded(true)
    })
  }, [environment, vaultId, version])

  const VGSCollectFieldStyles = {
    color: '#334155',
    width: '100%',
    height: '100%',
    paddingRight: '4rem',
    paddingLeft: '1rem',
    boxSizing: 'border-box',
    border: 'solid 2px #CBD5E1',
    borderRadius: '0.75rem',
    '&:focus': {
      borderColor: '#2563EB'
    },
    '&.invalid.dirty:not(:focus)': {
      borderColor: '#B91C1C'
    },
    '@font-face': {
      fontFamily: 'Inter',
      fontStyle: 'normal',
      fontWeight: '400',
      fontDisplay: 'swap',
      src: "url('https://cdn.fynbos.app/fonts/inter/v12/cyrillic-ext.woff2') format('woff2')",
      unicodeRange:
        'U+0460-052F, U+1C80-1C88, U+20B4, U+2DE0-2DFF, U+A640-A69F, U+FE2E-FE2F'
    },
    fontSize: '1rem',
    fontFamily: 'Inter, sans-serif'
  }

  const onSubmitCallback = (status: VGSCollectHttpStatusCode, resp: any) => {
    // TODO handle response from form and continue with flow.
    console.log('Submitted', status, resp)
  }

  const handleFormStateChange = (state: VGSCollectFormState) => {
    const fieldErrors = {
      cardNumber: '',
      expDate: '',
      cardSecurityCode: ''
    }
    if (
      state &&
      state['card-number'] &&
      state['card-number'].errors &&
      state['card-number'].isDirty &&
      !state['card-number'].isFocused &&
      state['card-number'].errors.length > 0
    ) {
      switch (state['card-number'].errors[0].code) {
        case 1001:
          fieldErrors.cardNumber = 'Card number is required.'
          break
        case 1011:
        case 1005:
        case 1004:
        case 1020:
        case 1010:
          fieldErrors.cardNumber = 'Card number is invalid.'
          break
      }
    }

    if (
      state &&
      state['card-security-code'] &&
      state['card-security-code'].errors &&
      state['card-security-code'].isDirty &&
      !state['card-security-code'].isFocused &&
      state['card-security-code'].errors.length > 0
    ) {
      switch (state['card-security-code'].errors[0].code) {
        case 1001:
          fieldErrors.cardSecurityCode = 'CVV is required.'
          break
        case 1017:
        case 1020:
        case 1010:
          fieldErrors.cardSecurityCode = 'CVV is invalid.'
          break
      }
    }

    if (
      state &&
      state['exp-date'] &&
      state['exp-date'].errors &&
      state['exp-date'].isDirty &&
      !state['exp-date'].isFocused &&
      state['exp-date'].errors.length > 0
    ) {
      switch (state['exp-date'].errors[0].code) {
        case 1001:
          fieldErrors.expDate = 'Expiration is required.'
          break
        case 1015:
        case 1020:
        case 1010:
          fieldErrors.expDate = 'Expiration is invalid.'
          break
      }
    }

    if (state && state['card-number']) {
      setCardInfo({
        last4: state && (state['card-number'].last4 as string),
        cardType: state && (state['card-number'].cardType as string)
      })
    }

    setFieldErrors(fieldErrors)
  }

  return (
    <Card>
      <div className='flex justify-between'>
        <h1 className='font-display text-2xl font-medium'>Debit card</h1>
        <div className='hidden sm:flex'>
          <Shape
            width={'w-8'}
            radius={'rounded-tl-full'}
            color={'bg-lime-500'}
          />
          <Shape
            width={'w-8'}
            radius={'rounded-tl-full'}
            color={'bg-slate-600'}
          />
        </div>
      </div>
      <p className='mt-6 text-medium'>
        Please provide your debit card details.
      </p>
      {isVGSCollectScriptLoaded && (
        <>
          {/**
           * VGS Collect form wrapper element. Abstraction over the VGSCollect.create()
           * https://www.verygoodsecurity.com/docs/api/collect/#api-vgscollectcreate
           */}
          <VGSCollectForm
            vaultId={vaultId as string}
            environment={environment as VGSCollectVaultEnvironment}
            // TODO Set this to the endpoint configured on the backend
            action='/webhooks/verygoodsecurity/card'
            submitParameters={{
              // JSON request body generated on the form submission including custom parameters
              // https://www.verygoodsecurity.com/docs/vgs-collect/js/integration#form-submit
              data: (fields: ICollectFormPayloadStructure) => {
                return {
                  ...fields,
                  walletId,
                  last4: cardInfo.last4,
                  cardType: cardInfo.cardType
                }
              }
            }}
            onUpdateCallback={handleFormStateChange}
            onSubmitCallback={onSubmitCallback}
            onErrorCalback={handleFormStateChange}
          >
            <label className='block mt-6'>
              <span className='ml-2 block text-sm font-medium text-medium'>
                Card number
              </span>
              <VGSCollectForm.CardNumberField
                placeholder=''
                className='w-full'
                validations={['required', 'validCardNumber']}
                showCardIcon={{
                  right: '1rem'
                }}
                name='card-number'
                css={VGSCollectFieldStyles}
              />
              <div className='h-7 pt-2 pl-2'>
                {fieldErrors.cardNumber && (
                  <p className='text-sm text-error'>{fieldErrors.cardNumber}</p>
                )}
              </div>
            </label>
            <div className='flex space-x-4'>
              <label className='block mt-1'>
                <span className='ml-2 block text-sm font-medium text-medium'>
                  Expiry date
                </span>
                <VGSCollectForm.CardExpirationDateField
                  placeholder='MM / YY'
                  className='flex w-full'
                  validations={['required', 'validCardExpirationDate']}
                  yearLength={2}
                  css={{ ...VGSCollectFieldStyles, paddingRight: '0' }}
                  name='exp-date'
                />
                <div className='h-7 pt-2 pl-2'>
                  {fieldErrors.expDate && (
                    <p className='text-sm text-error'>{fieldErrors.expDate}</p>
                  )}
                </div>
              </label>
              <label className='block mt-1'>
                <span className='ml-2 block text-sm font-medium text-medium'>
                  Security code
                </span>
                <VGSCollectForm.CardSecurityCodeField
                  placeholder=''
                  className='w-full'
                  name='card-security-code'
                  validations={['required', 'validCardSecurityCode']}
                  css={VGSCollectFieldStyles}
                  showCardIcon={{
                    right: '1rem'
                  }}
                />
                <div className='h-7 pt-2 pl-2'>
                  {fieldErrors.cardSecurityCode && (
                    <p className='text-sm text-error'>
                      {fieldErrors.cardSecurityCode}
                    </p>
                  )}
                </div>
              </label>
            </div>

            <Button className='mt-12' type='submit'>
              Submit
            </Button>
          </VGSCollectForm>
        </>
      )}
    </Card>
  )
}
