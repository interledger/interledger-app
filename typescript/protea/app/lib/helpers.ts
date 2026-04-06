import { Errors } from './types'

export function formatError(error: Errors | null | undefined) {
  const filteredErrors = error?.errors?.filter(
    (e): e is string => Boolean(e)
  )

  if (!filteredErrors?.length) return undefined

  return filteredErrors.join(', ')
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