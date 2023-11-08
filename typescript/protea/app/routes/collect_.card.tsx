import {
  redirect,
  type ActionFunctionArgs,
  type LoaderFunctionArgs,
  type MetaFunction
} from '@remix-run/node'
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
  TextElement,
  useBasisTheory
} from '@basis-theory/basis-theory-react'
import type {
  CardExpirationDateElement as CardExpirationDateElementType,
  CardNumberElement as CardNumberElementType,
  CardVerificationCodeElement as CardVerificationCodeElementType
} from '@basis-theory/basis-theory-react/types'
import { Code } from '@bufbuild/connect'
import clsx from 'clsx'
import { AnimatePresence, motion } from 'framer-motion'
import { route } from 'routes-gen'
import type { ApplicationProps } from '~/components'
import { Button, Card, CardContent, Layouts } from '~/components'
import { jsonWithCSRF, validateCSRFToken } from '~/lib/csrf.server'
import { isConnectError } from '~/lib/error.server'
import { grpc } from '~/lib/grpc.server'
import { mergeMeta } from '~/lib/meta'
import { redirectWithSnackbar } from '~/lib/snackbar.server'
import { useScaffoldStore } from '~/lib/useScaffoldStore'

export async function loader({ request }: LoaderFunctionArgs) {
  const url = new URL(request.url)
  const token = url.searchParams.get('token') || ''
  let link = await grpc.introspect(request, { token })

  // TODO: exhaustive precondition failure handling
  if (isConnectError(link)) {
    throw link.errorResponse
  }

  if (link.completed || link.expired) {
    return redirect(route('/collect/:linkId', { linkId: link.id }))
  }

  return jsonWithCSRF(request, {
    walletId: link.receiverWalletId,
    btToken: process.env.BT_TOKEN || '',
    fynbosEnv: process.env.FYNBOS_ENV,
    paymentLinkToken: token
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

export const meta: MetaFunction = mergeMeta(() => [
  {
    title: 'Connect card'
  }
])

export default function Page() {
  const { walletId, btToken, paymentLinkToken, csrfToken, fynbosEnv } =
    useLoaderData<typeof loader>()
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

  const { bt, error } = useBasisTheory(btToken, { elements: true })

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
    if (actionData) {
      setLoading(false)
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
            action: `${route('/collect/card')}?token=${paymentLinkToken}`,
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
              Card number {fieldErrors.number}
            </span>
            <div
              className={clsx(
                'mt-1 flex h-12 w-full items-center justify-between overflow-hidden rounded-xl border-2 border-base pr-4 focus-within:border-focus focus-within:ring-0',
                cardNumberFocus && 'border-focus ring-0'
              )}
            >
              <div className='block w-full'>
                {fynbosEnv == 'dev' ? (
                  <TextElement
                    ref={cardNumberRef}
                    onReady={() => setLoading(false)}
                    onFocus={() => setCardNumberFocus(true)}
                    onBlur={() => setCardNumberFocus(false)}
                    style={btStyle}
                    placeholder=''
                    id='card-number'
                  />
                ) : (
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
                )}
              </div>
            </div>
            <AnimatePresence>
              {(fieldErrors.number || actionData?.errors?.number) && (
                <motion.div
                  animate={{ opacity: 1, y: 0 }}
                  initial={{ opacity: 0, y: -8 }}
                  exit={{
                    opacity: 0,
                    y: -8,
                    transition: {
                      duration: 0.2
                    }
                  }}
                  transition={{
                    type: 'spring',
                    stiffness: 400,
                    damping: 20,
                    duration: 0.3
                  }}
                  className='h-7 pl-2 pt-2'
                >
                  <p className='text-sm text-error'>
                    {fieldErrors.number || actionData?.errors?.number}
                  </p>
                </motion.div>
              )}
            </AnimatePresence>
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
              <AnimatePresence>
                {fieldErrors.date && (
                  <motion.div
                    animate={{ opacity: 1, y: 0 }}
                    initial={{ opacity: 0, y: -8 }}
                    exit={{
                      opacity: 0,
                      y: -8,
                      transition: {
                        duration: 0.2
                      }
                    }}
                    transition={{
                      type: 'spring',
                      stiffness: 400,
                      damping: 20,
                      duration: 0.3
                    }}
                    className='h-7 pl-2 pt-2'
                  >
                    <p className='text-sm text-error'>{fieldErrors.date}</p>
                  </motion.div>
                )}
              </AnimatePresence>
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
              <AnimatePresence>
                {fieldErrors.cvc && (
                  <motion.div
                    animate={{ opacity: 1, y: 0 }}
                    initial={{ opacity: 0, y: -8 }}
                    exit={{
                      opacity: 0,
                      y: -8,
                      transition: {
                        duration: 0.2
                      }
                    }}
                    transition={{
                      type: 'spring',
                      stiffness: 400,
                      damping: 20,
                      duration: 0.3
                    }}
                    className='h-7 pl-2 pt-2'
                  >
                    <p className='text-sm text-error'>{fieldErrors.cvc}</p>
                  </motion.div>
                )}
              </AnimatePresence>
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

export async function action({ request }: ActionFunctionArgs) {
  const form = await request.formData()
  const cardToken = form.get('tokenId') as string

  await validateCSRFToken(request, form)

  const errors = {
    form: '',
    number: ''
  }
  const mapping = {
    number: 'CardNumber'
  }

  const url = new URL(request.url)
  const token = url.searchParams.get('token') || ''
  let response = await grpc.createPaymentLinkCard(
    request,
    {
      basisTheoryTokenId: cardToken,
      token
    },
    {
      timeoutMs: 60 * 1000
    }
  )

  if (isConnectError(response)) {
    if (response.code == Code.InvalidArgument) {
      return response.error({ errors }, mapping)
    } else if (response.code == Code.FailedPrecondition) {
      errors.form = response.violations[0].description
      return response.error({ errors }, mapping)
    } else if (response.code == Code.AlreadyExists) {
      errors.form = 'This card is already connected to Fynbos.'
      return response.error({ errors }, mapping)
    } else {
      if (response.code == Code.Unavailable) {
        errors.form = 'We did not receive a response from our card processor.'
      }
      if (errors.form == '') {
        errors.form = 'There was an error connecting your card.'
      }
      return response.error({ errors }, mapping, { action: 'Contact support' })
    }
  }

  return redirectWithSnackbar(
    request,
    `${route('/collect/success')}?token=${token}`,
    {
      message: 'New card successfully saved.',
      icon: 'close'
    }
  )
}
