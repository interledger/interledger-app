import { z } from 'zod'
import { createClient, getIncomingPaymentGrant, getWalletAddress } from './open-payments.server'
import { Errors, FormattedAmount, FormatAmountArgs, WalletAddressType } from './types'
import { getUserSession } from '~/lib/kratos.server'

export class WalletAddressFormatError extends Error { }

export async function getValidWalletAddress(walletAddress: string) {
  const opClient = await createClient()
  const response = await getWalletAddress(walletAddress, opClient)
  return response
}

export async function createRequestPayment(args: {
  receiverAddress: string
  amount: number
  note?: string
}) {
  const opClient = await createClient()
  const walletAddress = await getWalletAddress(args.receiverAddress, opClient)

  const amountObj = {
    value: BigInt(
      (args.amount * 10 ** walletAddress.assetScale).toFixed()
    ).toString(),
    assetCode: walletAddress.assetCode,
    assetScale: walletAddress.assetScale
  }

  const incomingPaymentGrant = await getIncomingPaymentGrant(
    walletAddress.authServer,
    opClient
  )

  // create incoming payment with amount
  return await opClient.incomingPayment
    .create(
      {
        url: walletAddress.resourceServer,
        accessToken: incomingPaymentGrant.access_token?.value || ''
      },
      {
        expiresAt: new Date(Date.now() + 6000 * 60 * 5).toISOString(),
        walletAddress: walletAddress.id,
        incomingAmount: amountObj,
        metadata: {
          description: args.note
        }
      }
    )
    .catch(() => {
      throw new Error('Unable to create incoming payment for request.')
    })
}

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
  senderAddress: z.string().optional(),
  receiverAddress: z
    .string()
    .transform((val) => val.replace('$', 'https://'))
    .pipe(z.string().url({ message: 'The input is not a wallet address.' })),
  amount: z.coerce.number(),
  note: z.string().optional()
})

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

export function formatError(error: Errors | null | undefined) {
  const filteredErrors = error?.errors?.filter(
    (e): e is string => Boolean(e)
  )

  if (!filteredErrors?.length) return undefined

  return filteredErrors.join(', ')
}

export function createError(key: string, message: string): Errors {
  return {
    [key]: {
      errors: [message],
    },
  }
}

export const getCurrencySymbol = (assetCode: string): string => {
  return new Intl.NumberFormat('en-US', {
    currency: assetCode,
    style: 'currency',
    maximumFractionDigits: 0,
    minimumFractionDigits: 0
  })
    .format(0)
    .replace(/0/g, '')
    .trim()
}

export const isWalletAddress = (
  o: WalletAddressType
): o is WalletAddressType => {
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
export async function isWalletLayout(request: Request) {
  try {
      await getUserSession(request)
      return true
  
    } catch (err) {
      return false
    }
}


