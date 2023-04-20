import type {
  LoaderArgs,
  MetaFunction,
  LinksFunction,
  ActionArgs
} from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { Form, useLoaderData, useSubmit } from '@remix-run/react'
import type { InputHTMLAttributes } from 'react'
import { FC, forwardRef, useRef, useState } from 'react'

import styles from '~/styles/VGS.css'
import { Button, Card, Layouts } from '~/components'
import { flowType, requireFlow, updateFlow } from '~/lib/flows.server'
import { getLinkedAccounts, getWalletId } from '~/lib/wallet.server'
import {
  BasisTheoryApiError,
  BasisTheoryProvider,
  BasisTheoryValidationError,
  CardExpirationDateElement,
  CardNumberElement,
  CardVerificationCodeElement,
  useBasisTheory
} from '@basis-theory/basis-theory-react'
import type {
  CardExpirationDateElement as CardExpirationDateElementType,
  CardNumberElement as CardNumberElementType,
  CardVerificationCodeElement as CardVerificationCodeElementType
} from '@basis-theory/basis-theory-react/types'
import { route } from 'routes-gen'
import { flashSnackbar } from '~/lib/snackbar.server'
import clsx from 'clsx'

export async function loader({ request, params }: LoaderArgs) {
  // const linkedAccounts = await getLinkedAccounts(request)
  // await requireFlow(request, flowType.LinkCardAccount)
  // await updateFlow(request, flowType.LinkCardAccount, {
  //   linkedAccountLength: linkedAccounts.linkedAccounts.length
  // })
  //
  // const walletId = await getWalletId(request)

  return json({
    walletId: '',
    token: process.env.BT_TOKEN || 'key_KzDnQyqbd13Lxu8LvmhDyb'
  })
}

export const handle = {
  title: 'Add debit card',
  layout: Layouts.FocusLayout
}

export const meta: MetaFunction = () => {
  return {
    title: 'Add debit card'
  }
}

export const links: LinksFunction = () => {
  return [{ rel: 'stylesheet', href: styles }]
}

export default function Page() {
  const { token } = useLoaderData<typeof loader>()
  const submit = useSubmit()

  const [fieldErrors, setFieldErrors] = useState({
    number: '',
    date: '',
    cvc: ''
  })

  const [cardNumberFocus, setCardNumberFocus] = useState<boolean>(false)
  const [cardExpirationDateFocus, setExpirationDateFocus] =
    useState<boolean>(false)
  const [cardVerificationCodeFocus, setVerificationCodeFocus] =
    useState<boolean>(false)

  const { bt, error } = useBasisTheory(token, { elements: true })

  const cardNumberRef = useRef<CardNumberElementType>(null)
  const cardExpirationDateRef = useRef<CardExpirationDateElementType>(null)
  const cardVerificationCodeRef = useRef<CardVerificationCodeElementType>(null)

  const btSubmit = async () => {
    const cardNumber = cardNumberRef.current
    const cardExpirationDate = cardExpirationDateRef.current
    const cardVerificationCode = cardVerificationCodeRef.current

    try {
      // Figure out the correct typing for tokens
      const tokens = await bt?.tokenize({
        type: 'card',
        data: {
          number: cardNumber,
          expiration_month: cardExpirationDate?.month() ?? 0,
          expiration_year: cardExpirationDate?.year() ?? 0,
          cvc: cardVerificationCode
        }
      })

      let formData = new FormData()
      formData.append('tokenId', tokens?.id)
      // TODO: submit this to the backend.
      submit(formData, {
        action: `/linked-account/card/bt`,
        method: 'post'
      })
    } catch (error) {
      // console.error('BS', error.details)
      if (error instanceof BasisTheoryValidationError) {
        // check error details and setFieldErrors
        console.log('BS:validation', error.details)
      } else if (error instanceof BasisTheoryApiError) {
        // check error data or status
        console.log('BS:API', error)
      } else {
        // handle other errors
        console.log('BS:other', Object.keys(error.details.data))
        setFieldErrors({
          number: error.details.data.number ? 'Card number is invalid' : '',
          date: error.details.data.expiration_year
            ? 'Expiry date is invalid'
            : '',
          cvc: error.details.data.cvc ? 'Security code is invalid' : ''
        })
      }
    }
  }
  // detect if the user is using dark mode
  const isDark =
    typeof window !== 'undefined' &&
    window.matchMedia &&
    window.matchMedia('(prefers-color-scheme: dark)').matches

  const baseStyle = {
    padding: '0.75rem 1rem',
    fontFamily: 'Inter',
    fontStyle: 'normal',
    fontWeight: '400',
    fontSize: '1rem',
    lineHeight: '1.5rem',
    color: isDark ? '#FFFFFF' : '#334155',
    backgroundColor: 'transparent',
    '::placeholder': {
      color: '#64748B'
    },
    '::selection': {
      backgroundColor: 'rgb(244 63 94 / 0.5)'
    }
  }

  const btStyle = {
    fonts: ['https://fonts.googleapis.com/css2?family=Inter&display=swap'],
    base: baseStyle,
    complete: baseStyle,
    empty: baseStyle,
    invalid: baseStyle
  }

  if (error) {
    // initialization error
    throw new Error('BasisTheory initialization error')
  }

  // instance stays undefined during initialization
  if (bt) {
    return (
      <BasisTheoryProvider bt={bt}>
        <Card>
          <p className='text-medium'>Please provide your debit card details.</p>
          <label className='block mt-6'>
            <span className='ml-2 block text-sm font-medium text-medium'>
              Card number
            </span>
            <div
              className={clsx(
                'mt-1 flex h-12 w-full items-center justify-between overflow-hidden pr-4 rounded-xl border-2 border-base focus-within:border-focus focus-within:ring-0',
                cardNumberFocus && 'border-focus ring-0'
              )}
            >
              <div className='block w-full'>
                <CardNumberElement
                  ref={cardNumberRef}
                  onFocus={() => setCardNumberFocus(true)}
                  onBlur={() => setCardNumberFocus(false)}
                  style={btStyle}
                  iconPosition={'right'}
                  placeholder=''
                  id='card-number'
                />
              </div>
            </div>
            <div className='h-7 pt-2 pl-2'>
              {fieldErrors.number && (
                <p className='text-sm text-error'>{fieldErrors.number}</p>
              )}
            </div>
          </label>
          <div className='flex w-full space-x-4 mt-1'>
            <label className='block w-full'>
              <span className='ml-2 block text-sm font-medium text-medium'>
                Expiry date
              </span>
              <div
                className={clsx(
                  'mt-1 flex h-12 w-full items-center justify-between overflow-hidden pr-4 rounded-xl border-2 border-base focus-within:border-focus focus-within:ring-0',
                  cardExpirationDateFocus && 'border-focus ring-0'
                )}
              >
                <div className='block w-full'>
                  <CardExpirationDateElement
                    ref={cardExpirationDateRef}
                    onFocus={() => setExpirationDateFocus(true)}
                    onBlur={() => setExpirationDateFocus(false)}
                    style={btStyle}
                    id='expirationn-date'
                  />
                </div>
              </div>
              <div className='h-7 pt-2 pl-2'>
                {fieldErrors.date && (
                  <p className='text-sm text-error'>{fieldErrors.date}</p>
                )}
              </div>
            </label>
            <label className='block w-full'>
              <span className='ml-2 block text-sm font-medium text-medium'>
                Security code
              </span>
              <div
                className={clsx(
                  'mt-1 flex h-12 w-full items-center justify-between overflow-hidden pr-4 rounded-xl border-2 border-base focus-within:border-focus focus-within:ring-0',
                  cardVerificationCodeFocus && 'border-focus ring-0'
                )}
              >
                <div className='block w-full'>
                  <CardVerificationCodeElement
                    ref={cardVerificationCodeRef}
                    onFocus={() => setVerificationCodeFocus(true)}
                    onBlur={() => setVerificationCodeFocus(false)}
                    style={btStyle}
                    placeholder=''
                    id='card-verification-code'
                  />
                </div>
              </div>
              <div className='h-7 pt-2 pl-2'>
                {fieldErrors.cvc && (
                  <p className='text-sm text-error'>{fieldErrors.cvc}</p>
                )}
              </div>
            </label>
          </div>
        </Card>
        <Button type='submit' onClick={btSubmit} disabled={!bt}>
          Submit
        </Button>
      </BasisTheoryProvider>
    )
  }
}

export async function action({ request }: ActionArgs) {
  const cookie = request.headers.get('Cookie') as string
  const url = new URL(request.url)

  const form = await request.formData()
  const cardToken = form.get('cardToken') as string

  // TOOD: submit cardToken to a grpcClient

  return redirect(route('/'), {
    headers: {
      'Set-Cookie': await flashSnackbar(request, {
        message: 'New card successfully saved.',
        icon: 'close'
      })
    }
  })
}
