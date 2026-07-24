import type { Access, Amount } from './rafikiauth.server'

type ConsentCard = {
  label: string
  value?: string
}

type GrantAction = Access['actions'][number]

export function buildConsentCards(access?: Access): ConsentCard[] {
  if (!access) return []

  const actions = new Set<GrantAction>(access.actions)
  const amount: Amount | undefined =
    access.limits?.debitAmount ?? access.limits?.receiveAmount

  const canSpend = actions.has('create')
  const canList = actions.has('list')
  const canReadAll = actions.has('read-all')
  const cards: ConsentCard[] = []

  if (canSpend) {
    cards.push(
      amount
        ? { label: 'Total amount to debit', value: formatAmount(amount) }
        : { label: 'Make payments on your behalf' }
    )
  }

  if (canReadAll) {
    cards.push({ label: 'View all payments on your account' })
  }

  if (!canSpend && !canReadAll && canList) {
    cards.push({
      label: actions.has('read')
        ? 'View your payments'
        : 'View a list of your payments'
    })
  }

  return cards
}

function formatAmount(amount: Amount): string {
  let currency = '$'
  if (amount.assetCode != 'USD') {
    currency = amount.assetCode
  }

  const amt = parseInt(amount.value) * Math.pow(10, -amount.assetScale)
  return `${currency} ${amt.toFixed(2)}`
}
