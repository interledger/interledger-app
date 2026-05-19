import {
  CardOperation,
  CardTransactionDetails
} from '~/generated/connect/backend/v1/backend_pb'

const WITHDRAW_TYPES = [
  'Purchase',
  'ATMWithdrawal',
  'CashAdvance',
  'Preauthorization',
  'PreauthorizationIncremental',
  'PreauthorizationCompletion',
  'TransferFromAccount',
]

export function computeCardSubtotalStyles(
  details: CardTransactionDetails | undefined
): string {
  if (details?.operation === CardOperation.WITHDRAW) return 'text-error'
  if (details?.operation === CardOperation.DEPOSIT) return 'text-medium'
  if (details?.operation === CardOperation.NONE) {
    return WITHDRAW_TYPES.includes(details.type) ? 'text-error' : 'text-base'
  }
  return ''
}
