import type { LoaderArgs, MetaFunction } from '@remix-run/node'
import { json, LinksFunction } from '@remix-run/node'
import { useLoaderData, useNavigate, useParams } from '@remix-run/react'
import { useCallback, useEffect, useState } from 'react'

import styles from '~/styles/VGS.css'
import { Button, Card, Layouts, Shape } from '~/components'
import { route } from 'routes-gen'
import { flowType, requireFlow, updateFlow } from '~/lib/flows.server'
import { getLinkedAccounts } from '~/lib/wallet.server'
import { loadVGSCollect } from '@vgs/collect-js'
import {
  useVGSCollectResponse,
  useVGSCollectState,
  VGSCollectFocusEventData,
  VGSCollectForm,
  VGSCollectFormState,
  VGSCollectKeyboardEventData,
  VGSCollectStateParams,
  VGSCollectVaultEnvironment
} from '@vgs/collect-js-react'
import CardSecurityCodeField = VGSCollectForm.CardSecurityCodeField
import CardExpirationDateField = VGSCollectForm.CardExpirationDateField
import CardNumberField = VGSCollectForm.CardNumberField

export async function loader({ request, params }: LoaderArgs) {
  const linkedAccounts = await getLinkedAccounts(request)
  await requireFlow(request, flowType.LinkCardAccount)
  await updateFlow(request, flowType.LinkCardAccount, {
    linkedAccountLength: linkedAccounts.linkedAccounts.length
  })

  // TODO: RPC endpoint to get VGS token

  return json({
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
  const { vaultId, environment, version } = useLoaderData<typeof loader>()
  const [isVGSCollectScriptLoaded, setCollectScriptLoaded] = useState(false)
  const [fieldErrors, setFieldErrors] = useState({
    cardNumber: '',
    expDate: '',
    cardSecurityCode: ''
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

  /**
   * VGS Collect state hook to retrieve the form state
   */
  const [state] = useVGSCollectState()

  /**
   * VGS Collect submit hook to retrieve the form response
   */
  const [response] = useVGSCollectResponse()

  const VGSCollectFieldStyles = {
    color: '#1b1d1f',
    width: '100%',
    height: '100%',
    paddingRight: '4rem',
    paddingLeft: '1rem',
    boxSizing: 'border-box',
    border: 'solid 2px #CBD5E1',
    borderRadius: '0.75rem',
    '&:focus': {
      borderColor: '#2563EB'
    }
  }

  useEffect(() => {
    /**
     * Track form state
     */
    console.log('State hook changed', state)
  }, [state])

  useEffect(() => {
    /**
     * Track response from the VGS Collect form
     */
    console.log('Response hook changed', response)
  }, [response])

  const onSubmitCallback = (status: VGSCollectHttpStatusCode, resp: any) => {
    /**
     * Receive information about HTTP request
     */
    console.log('Submitted', status)
  }

  const onUpdateCallback = (state: VGSCollectFormState) => {
    console.log('Updated', state)
    // isDirty: boolean;
    // isTouched: boolean;
    // isFocused: boolean;
    // isValid: boolean;
    // isEmpty: boolean;
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

    setFieldErrors(fieldErrors)
  }

  const onErrorCallback = (errors: VGSCollectFormState) => {
    const fieldErrors = {
      cardNumber: '',
      expDate: '',
      cardSecurityCode: ''
    }
    if (errors && errors['card-number'] && errors['card-number'].errors) {
      switch (errors['card-number'].errors[0].code) {
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

    if (errors && errors['exp-date'] && errors['exp-date'].errors) {
      switch (errors['exp-date'].errors[0].code) {
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

    if (
      errors &&
      errors['card-security-code'] &&
      errors['card-security-code'].errors
    ) {
      switch (errors['card-security-code'].errors[0].code) {
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
    console.log('errors', errors)
    console.log('fieldErrors', fieldErrors)
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
            action='/post'
            submitParameters={{}}
            onUpdateCallback={onUpdateCallback}
            onSubmitCallback={onSubmitCallback}
            onErrorCalback={onErrorCallback}
          >
            <label className='block mt-6'>
              <span className='ml-2 block text-sm font-medium text-medium'>
                Card number
              </span>
              <CardNumberField
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
            <div className='flex space-x-2'>
              <label className='block mt-6'>
                <span className='ml-2 block text-sm font-medium text-medium'>
                  Expiry date
                </span>
                <CardExpirationDateField
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
              <label className='block mt-6'>
                <span className='ml-2 block text-sm font-medium text-medium'>
                  Security code
                </span>
                <CardSecurityCodeField
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
