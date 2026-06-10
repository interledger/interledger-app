import { z } from 'zod'
import { type WalletAddress } from '@interledger/open-payments'
import { Errors, FormattedAmount, FormatAmountArgs } from './types'
import { redirect } from 'react-router'
import { getCurrencySymbol } from '~/lib/helpers'
import { envBool } from '~/env.server'

export class WalletAddressFormatError extends Error { }

export const walletSchema = z.object({
  walletAddress: z
    .string()
    .min(1, { message: 'Wallet address is required' })
    .transform((url) => toWalletAddressUrl(url))
    .superRefine(async (updatedUrl, ctx) => {
      if (updatedUrl.length === 0) return

      try {
        checkHrefFormat(updatedUrl)
        await isValidWalletAddress(updatedUrl)
      } catch (e) {
        console.log({ e })
        ctx.addIssue({
          code: z.ZodIssueCode.custom,
          message:
            e instanceof WalletAddressFormatError
              ? e.message
              : 'Invalid wallet address format'
        })
      }
    })
})

export const paymentSchema = z.object({
  senderAddress: z.string(),
  receiverAddress: z
    .string()
    .transform((val) => val.replace('$', 'https://'))
    .pipe(z.string().url({ message: 'The input is not a wallet address.' })),
  amount: z.coerce.number(),
  note: z
    .string()
    .max(255, { message: 'Payment note must be 255 characters or less.' })
    .optional()
})

export const requestSchema = z.object({
  senderAddress: z.string()
})

export const requestPaymentSchema = paymentSchema.extend({
  senderAddress: paymentSchema.shape.senderAddress.optional(),
});

function checkHrefFormat(href: string): void {
  let url: URL
  try {
    url = new URL(href)
    if (url.protocol !== 'https:') {
      throw new WalletAddressFormatError(
        'Wallet address must use HTTPS protocol'
      )
    }
  } catch (e) {
    if (e instanceof WalletAddressFormatError) {
      throw e
    }
    throw new WalletAddressFormatError(
      `Invalid wallet address URL: ${JSON.stringify(href)}`
    )
  }

  const { hash, search, port, username, password } = url

  if (hash || search || port || username || password) {
    throw new WalletAddressFormatError(
      `Wallet address URL must not contain query/fragment/port/username/password elements.`
    )
  }
}

async function isValidWalletAddress(
  walletAddressUrl: string
): Promise<boolean> {
  const response = await fetch(walletAddressUrl, {
    headers: {
      Accept: 'application/json'
    }
  })

  if (!response.ok) {
    if (response.status === 404) {
      throw new WalletAddressFormatError('This wallet address does not exist.')
    }
    throw new WalletAddressFormatError('Failed to fetch wallet address.')
  }

  const msgInvalidWalletAddress = 'Provided URL is not a valid wallet address.'
  const json = await response.json().catch((error) => {
    throw new WalletAddressFormatError(msgInvalidWalletAddress, {
      cause: error
    })
  })

  if (!isWalletAddress(json)) {
    throw new WalletAddressFormatError(msgInvalidWalletAddress)
  }

  return true
}

export function createError(key: string, message: string): Errors {
  return {
    [key]: {
      errors: [message],
    },
  }
}

export const isWalletAddress = (
  o: WalletAddress
): o is WalletAddress => {
  return !!(
    o.id &&
    typeof o.id === 'string' &&
    o.assetScale &&
    typeof o.assetScale === 'number' &&
    o.assetCode &&
    typeof o.assetCode === 'string' &&
    o.authServer &&
    typeof o.authServer === 'string' &&
    o.resourceServer &&
    typeof o.resourceServer === 'string'
  )
}

export function toWalletAddressUrl(s: string): string {
  return s.startsWith('$') ? s.replace('$', 'https://') : s
}

type FormatDateArgs = {
  date: string
  time?: boolean
  month?: Intl.DateTimeFormatOptions['month']
}
export const formatDate = ({
  date,
  time = true,
  month = 'short'
}: FormatDateArgs): string => {
  return new Date(date).toLocaleDateString('default', {
    day: '2-digit',
    month,
    year: 'numeric',
    ...(time && { hour: '2-digit', minute: '2-digit' })
  })
}

export const formatAmount = (args: FormatAmountArgs): FormattedAmount => {
  const { value, assetCode, assetScale } = args
  const formatterWithCurrency = new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: assetCode,
    maximumFractionDigits: assetScale,
    minimumFractionDigits: assetScale
  })
  const formatter = new Intl.NumberFormat('en-US', {
    maximumFractionDigits: assetScale,
    minimumFractionDigits: assetScale
  })

  const amount = Number(formatter.format(Number(`${value}e-${assetScale}`)))
  const amountWithCurrency = formatterWithCurrency.format(
    Number(`${value}e-${assetScale}`)
  )
  const symbol = getCurrencySymbol(assetCode)

  return {
    amount,
    amountWithCurrency,
    symbol
  }
}

export const routeAllowed = (featureName: string) => {
  if (!envBool(featureName)) {
    throw redirect('/')
  }
}

type FormDataObject = Record<string, FormDataEntryValue>
