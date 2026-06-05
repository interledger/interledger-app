import type { CardTransactionDetails } from '~/generated/connect/backend/v1/backend_pb'
import { CardOperation } from '~/generated/connect/backend/v1/backend_pb'

const WITHDRAW_TYPES = [
  'Purchase',
  'ATMWithdrawal',
  'CashAdvance',
  'Preauthorization',
  'PreauthorizationIncremental',
  'PreauthorizationCompletion',
  'TransferFromAccount'
]
const DEPOSIT_TYPES = ['TransferToAccount']

export function computeCardPaymentLabel(
  details: CardTransactionDetails | undefined
): string {
  if (details?.operation === CardOperation.WITHDRAW) return 'Payment from'
  if (details?.operation === CardOperation.DEPOSIT) return 'Payment to'
  if (details?.operation === CardOperation.NONE) {
    if (WITHDRAW_TYPES.includes(details.type)) return 'Payment from'
    if (DEPOSIT_TYPES.includes(details.type)) return 'Payment to'
  }
  return 'Operation on'
}

export function computeCardSubtotalStyles(
  details: CardTransactionDetails | undefined
): string {
  if (details?.operation === CardOperation.WITHDRAW) return 'text-error'
  if (details?.operation === CardOperation.DEPOSIT) return 'text-medium'
  if (details?.operation === CardOperation.NONE) {
    return WITHDRAW_TYPES.includes(details.type) ? 'text-error' : 'text-medium'
  }
  return ''
}
