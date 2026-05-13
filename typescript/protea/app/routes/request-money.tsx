import { Code } from '@bufbuild/connect'
import { Timestamp } from '@bufbuild/protobuf'
import { useState } from 'react'
import { Form, href, useActionData, useLoaderData } from 'react-router'
import type { ApplicationProps } from '~/components'
import {
  Button,
  ButtonRouter,
  Card,
  CardContent,
  Layouts,
  TextField
} from '~/components'
import { getWalletInfo } from '~/data/wallet.server'
import { stringToBigInt } from '~/lib/amount'
import { jsonWithCSRF, validateCSRFToken } from '~/lib/csrf.server'
import { isConnectError } from '~/lib/error.server'
import { grpc } from '~/lib/grpc.server'
import { mergeMeta } from '~/lib/meta'
import { generateQR, qrSvg } from '~/lib/qr.server'
import type { Route } from './+types/request-money'

// Gatehub-only on the backend → EUR scale=2 for v1. When multi-asset
// support lands, source assetCode/assetScale from the wallet address
// (WalletInfo doesn't carry them today).
const ASSET_CODE = 'EUR'
const ASSET_SCALE = 2

type RequestErrors = {
  form?: string
  amount?: string
  expiresAt?: string
}

type RequestSuccess = {
  url: string
  formattedAmount: string
  description: string
  expiresAt: string
  qrSvg: string
}

export const handle: ApplicationProps = {
  layout: Layouts.Focus,
  scaffold: {
    header: {
      title: 'Request money',
      back: '/'
    }
  }
}

export const meta = mergeMeta(() => [{ title: 'Request money' }])

export async function loader({ request }: Route.LoaderArgs) {
  const wallet = await getWalletInfo(request)
  return jsonWithCSRF(request, {
    assetCode: ASSET_CODE,
    walletUrl: wallet.url
  })
}

export async function action({ request }: Route.ActionArgs) {
  const form = await request.formData()
  await validateCSRFToken(request, form)

  const amount = String(form.get('amount') || '').trim()
  const expiresAt = String(form.get('expiresAt') || '').trim()
  const description = String(form.get('description') || '').trim()

  const errors: RequestErrors = {}
  if (!amount || !(parseFloat(amount) > 0)) {
    errors.amount = 'Enter an amount greater than zero.'
  }
  const expiresAtDate = new Date(expiresAt)
  if (!expiresAt || Number.isNaN(expiresAtDate.getTime())) {
    errors.expiresAt = 'Choose a valid expiration date and time.'
  } else if (expiresAtDate.getTime() < Date.now() + 30_000) {
    errors.expiresAt = 'Expiration must be at least 30 seconds from now.'
  }
  if (Object.keys(errors).length > 0) {
    return { errors }
  }

  const response = await grpc.createIncomingPaymentRequest(request, {
    value: stringToBigInt(amount, ASSET_SCALE),
    expiresAt: Timestamp.fromDate(expiresAtDate),
    description
  })

  if (isConnectError(response)) {
    if (response.code === Code.InvalidArgument) {
      return response.error({
        errors: { form: response._err.rawMessage || 'Invalid request.' }
      })
    }
    if (response.code === Code.NotFound) {
      return response.error({
        errors: {
          form: 'Wallet not provisioned for incoming payments. Set up a wallet address first.'
        }
      })
    }
    if (response.code === Code.Unauthenticated) {
      return response.error({
        errors: { form: 'You must be signed in to request money.' }
      })
    }
    return response.error({
      errors: {
        form: `Could not create payment request (${Code[response.code]}).`
      }
    })
  }

  const qr = await generateQR(response.url)
  const success: RequestSuccess = {
    url: response.url,
    formattedAmount: `${amount} ${ASSET_CODE}`,
    description,
    expiresAt: expiresAtDate.toISOString(),
    qrSvg: qrSvg(qr)
  }
  return { success }
}

export default function RequestMoneyRoute() {
  const { csrfToken, assetCode, walletUrl } = useLoaderData<typeof loader>()
  const actionData = useActionData<typeof action>()

  if (actionData && 'success' in actionData && actionData.success) {
    return <RequestResult success={actionData.success} />
  }

  const errors: RequestErrors =
    (actionData && 'errors' in actionData && actionData.errors) || {}

  return (
    <RequestForm
      csrfToken={csrfToken}
      assetCode={assetCode}
      walletUrl={walletUrl}
      errors={errors}
    />
  )
}

function RequestForm({
  csrfToken,
  assetCode,
  walletUrl,
  errors
}: {
  csrfToken: string
  assetCode: string
  walletUrl: string
  errors: RequestErrors
}) {
  return (
    <Form
      method='post'
      action={href('/request-money')}
      className='flex flex-col gap-y-4'
    >
      <input type='hidden' name='csrfToken' value={csrfToken} />
      <div className='w-full rounded-[1.25rem] bg-nav p-2'>
        <div className='p-2'>
          <p className='ml-2 text-sm font-medium text-medium'>
            My wallet address
          </p>
          <p className='ml-2 mt-1 break-all text-base text-medium'>
            {walletUrl}
          </p>
        </div>
      </div>
      <Card>
        <CardContent>
          <TextField
            id='amount'
            name='amount'
            label='Amount'
            type='number'
            min='0'
            step='0.01'
            placeholder='0.00'
            required
            aria-invalid={Boolean(errors.amount) || undefined}
            errorMessage={errors.amount}
            appendIcon={
              <span className='rounded-full bg-nav px-3 py-1 text-sm font-medium text-strong'>
                {assetCode}
              </span>
            }
          />
          <p className='ml-2 mt-2 text-xs text-weak'>
            Conversion rates will apply
          </p>
        </CardContent>
      </Card>
      <Card>
        <CardContent>
          <TextField
            id='expiresAt'
            name='expiresAt'
            label='Expiration date'
            labelSuffix='*'
            type='datetime-local'
            defaultValue={defaultExpiresAtLocal()}
            required
            aria-invalid={Boolean(errors.expiresAt) || undefined}
            errorMessage={errors.expiresAt}
          />
        </CardContent>
      </Card>
      <Card>
        <CardContent>
          <TextField
            id='description'
            name='description'
            label='Description'
            type='text'
            placeholder='Enter description'
            maxLength={140}
          />
        </CardContent>
      </Card>
      {errors.form && (
        <p className='text-sm text-red-600' role='alert'>
          {errors.form}
        </p>
      )}
      <Button type='submit'>Generate request</Button>
    </Form>
  )
}

function RequestResult({ success }: { success: RequestSuccess }) {
  const [copied, setCopied] = useState(false)

  const onCopy = async () => {
    try {
      await navigator.clipboard.writeText(success.url)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    } catch {
      // Clipboard may be blocked; the user can still select the URL manually.
    }
  }

  return (
    <div className='flex flex-col gap-y-4'>
      <Card>
        <CardContent className='flex flex-col items-center gap-y-4'>
          <div
            className='h-64 w-64'
            dangerouslySetInnerHTML={{ __html: success.qrSvg }}
          />
          <p className='text-lg font-medium'>{success.formattedAmount}</p>
          {success.description && (
            <p className='text-center text-medium'>{success.description}</p>
          )}
          <p className='break-all text-center text-xs text-weak'>
            {success.url}
          </p>
          <p className='text-xs text-weak'>
            Expires {new Date(success.expiresAt).toLocaleString()}
          </p>
        </CardContent>
      </Card>
      <Button type='button' onClick={onCopy}>
        {copied ? 'Link copied' : 'Copy link'}
      </Button>
      <ButtonRouter to={href('/request-money')} reloadDocument>
        Generate another request
      </ButtonRouter>
    </div>
  )
}

function defaultExpiresAtLocal(): string {
  const d = new Date()
  d.setDate(d.getDate() + 7)
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(
    d.getHours()
  )}:${pad(d.getMinutes())}`
}
