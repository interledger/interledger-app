import type { ActionArgs, LoaderArgs, MetaFunction } from '@remix-run/node'
import { json } from '@remix-run/node'
import {
  useActionData,
  useLoaderData,
  useNavigation,
  useSubmit
} from '@remix-run/react'
import { useEffect, useRef, useState } from 'react'

import {
  BasisTheoryProvider,
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
import clsx from 'clsx'
import { route } from 'routes-gen'
import type { ApplicationProps } from '~/components'
import { Button, Card, CardContent, Layouts } from '~/components'
import { jsonWithCSRF, validateCSRFToken } from '~/lib/csrf.server'
import { redirectWithSnackbar } from '~/lib/snackbar.server'
import { useScaffoldStore } from '~/lib/useScaffoldStore'
import { createCard, getWalletId } from '~/lib/wallet.server'

export async function loader({ request, params }: LoaderArgs) {
  const walletId = await getWalletId(request)
  return jsonWithCSRF(request, {
    walletId,
    token: process.env.BT_TOKEN || ''
  })
}

export const handle: ApplicationProps = {
  layout: Layouts.Focus,
  scaffold: {
    header: {
      back: route('/accounts'),
      title: 'Connect card'
    }
  }
}

export const meta: MetaFunction = () => {
  return {
    title: 'Connect card'
  }
}

export default function Page() {
  const { walletId, token, csrfToken } = useLoaderData<typeof loader>()
  const submit = useSubmit()
  const actionData = useActionData<typeof action>()

  const navigation = useNavigation()

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

  const [loading, setLoading] = useState<boolean>(true)
  const setScaffoldLoading = useScaffoldStore((state) => state.setLoading)

  useEffect(() => {
    setScaffoldLoading(loading)
  }, [loading, setScaffoldLoading])

  useEffect(() => {
    if (loading && navigation.location?.pathname.includes('accounts')) {
      setLoading(false)
    }
  }, [loading, navigation])

  useEffect(() => {
    if (actionData && actionData.error) {
      setLoading(false)
      let errorMessage: string
      switch (actionData.error) {
        case 'Failed precondition: ErrUnsupportedCard':
          errorMessage =
            'Your card is unsupported and cannot be connected to Fynbos.'
          break
        case 'Failed precondition: ErrUnsupportedCountry':
          errorMessage =
            'Your country is unsupported and your card cannot be connected to Fynbos.'
          break
        case 'Already exists: ErrDuplicateCard':
          errorMessage = 'Your card is already connected to Fynbos.'
          break
        case 'Failed precondition: ErrMaxCardsAdded':
          errorMessage =
            'You have connected the maximum number of cards to Fynbos.'
          break
        case 'Unavailable: ErrMultiStatus':
          errorMessage =
            'We did not receive a response from our card processor.'
          break
        default:
          errorMessage = 'There was an error connecting your card.'
      }
      setFieldErrors({
        number: errorMessage,
        cvc: '',
        date: ''
      })
    }
  }, [actionData, setLoading])

  const btSubmit = async () => {
    setLoading(true)
    const cardNumber = cardNumberRef.current
    const cardExpirationDate = cardExpirationDateRef.current
    const cardVerificationCode = cardVerificationCodeRef.current

    if (!bt) {
      throw new Error('BasisTheory not initialized')
    }

    await bt
      .tokenize({
        type: 'card',
        data: {
          number: cardNumber,
          expiration_month: cardExpirationDate?.month() ?? 0,
          expiration_year: cardExpirationDate?.year() ?? 0,
          cvc: cardVerificationCode
        },
        metadata: {
          wallet_id: walletId
        },
        deduplicate_token: true,
        fingerprint_expression: '{{ metadata.wallet_id }}{{ data.number }}'
      })
      .then((token) => {
        let formData = new FormData()
        if (!Array.isArray(token)) {
          formData.append('tokenId', token.id as string)
          formData.append('csrfToken', csrfToken)
          submit(formData, {
            action: `/connect/card`,
            method: 'post'
          })
        }
      })
      .catch((error) => {
        setLoading(false)
        setFieldErrors({
          number: error.details.data.number ? 'Card number is invalid.' : '',
          date: error.details.data.expiration_year
            ? 'Expiry date is invalid.'
            : '',
          cvc: error.details.data.cvc ? 'Security code is invalid.' : ''
        })
      })
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
          <CardContent>
            <p className='text-medium'>
              Debit cards enable both sending and receiving money, while credit
              cards only allow receiving money.
            </p>
          </CardContent>
          <CardContent>
            <p className='text-medium'>
              We currently only support Visa and Mastercard cards.
            </p>
          </CardContent>
          <label className='mt-2 block'>
            <span className='ml-2 block text-sm font-medium text-medium'>
              Card number
            </span>
            <div
              className={clsx(
                'mt-1 flex h-12 w-full items-center justify-between overflow-hidden rounded-xl border-2 border-base pr-4 focus-within:border-focus focus-within:ring-0',
                cardNumberFocus && 'border-focus ring-0'
              )}
            >
              <div className='block w-full'>
                <CardNumberElement
                  ref={cardNumberRef}
                  onReady={() => setLoading(false)}
                  onFocus={() => setCardNumberFocus(true)}
                  onBlur={() => setCardNumberFocus(false)}
                  style={btStyle}
                  iconPosition={'right'}
                  placeholder=''
                  id='card-number'
                />
              </div>
            </div>
            <div className='h-7 pl-2 pt-2'>
              {fieldErrors.number && (
                <p className='text-sm text-error'>{fieldErrors.number}</p>
              )}
            </div>
          </label>
          <div className='mt-4 flex w-full space-x-4'>
            <label className='block w-full'>
              <span className='ml-2 block text-sm font-medium text-medium'>
                Expiry date
              </span>
              <div
                className={clsx(
                  'mt-1 flex h-12 w-full items-center justify-between overflow-hidden rounded-xl border-2 border-base pr-4 focus-within:border-focus focus-within:ring-0',
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
              <div className='h-7 pl-2 pt-2'>
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
                  'mt-1 flex h-12 w-full items-center justify-between overflow-hidden rounded-xl border-2 border-base pr-4 focus-within:border-focus focus-within:ring-0',
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
              <div className='h-7 pl-2 pt-2'>
                {fieldErrors.cvc && (
                  <p className='text-sm text-error'>{fieldErrors.cvc}</p>
                )}
              </div>
            </label>
          </div>
        </Card>
        <Button type='submit' onClick={btSubmit} disabled={!bt || loading}>
          Submit
        </Button>
      </BasisTheoryProvider>
    )
  }
}

export async function action({ request }: ActionArgs) {
  const form = await request.formData()
  const cardToken = form.get('tokenId') as string

  await validateCSRFToken(request, form)

  let resp = await createCard(request, cardToken)
  if (resp.httpMapping?.status == 409 || resp.httpMapping?.status == 400) {
    return json(resp, resp.httpMapping?.status)
  } else if (resp.httpMapping?.status != 200) {
    throw json({}, resp.httpMapping)
  }

  return redirectWithSnackbar(
    request,
    route('/accounts/:accountId', { accountId: resp.linkedAccountID }),
    {
      message: 'New card successfully saved.',
      icon: 'close'
    }
  )
}
