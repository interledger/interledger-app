import type { Access, Amount } from './rafikiauth.server'

type ConsentCard = {
  label: string
  value?: string
  description?: string
}

type GrantAction = Access['actions'][number]

export function buildConsentCards(access?: Access): ConsentCard[] {
  if (!access) return []

  const actions = new Set<GrantAction>(access.actions)
  const amount: Amount | undefined =
    access.limits?.debitAmount ?? access.limits?.receiveAmount

  const canSpend = actions.has('create')
  const viewAll = actions.has('read-all') || actions.has('list-all')
  const viewOwn = actions.has('read') || actions.has('list')
  const showOwnView = viewOwn && !viewAll

  const cards: ConsentCard[] = []

  if (canSpend) {
    const card: ConsentCard = amount
      ? { label: 'Total amount to debit', value: formatAmount(amount) }
      : { label: 'Make payments on your behalf' }

    if (showOwnView) {
      card.description = 'Can also view your payments'
    }
    cards.push(card)
  }

  if (viewAll) {
    cards.push({ label: 'View all payments on your account' })
  }

  if (showOwnView && !canSpend) {
    cards.push({ label: ownViewLabel(actions) })
  }

  return cards
}

function ownViewLabel(actions: Set<GrantAction>): string {
  const canList = actions.has('list')
  const canRead = actions.has('read')
  if (canList && canRead) return 'View your payments'
  if (canList) return 'View a list of your payments'
  return 'View your payment detail'
}

function formatAmount(amount: Amount): string {
  let currency = '$'
  if (amount.assetCode != 'USD') {
    currency = amount.assetCode
  }

  const amt = parseInt(amount.value) * Math.pow(10, -amount.assetScale)
  return `${currency} ${amt.toFixed(2)}`
}
