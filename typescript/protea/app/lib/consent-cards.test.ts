import { describe, expect, it } from 'vitest'
import { buildConsentCards } from './consent-cards'
import type { Access, Amount } from './rafikiauth.server'

const USD_10: Amount = { value: '1000', assetScale: 2, assetCode: 'USD' }

function op(actions: Access['actions'], limits?: Access['limits']): Access {
  return { type: 'outgoing-payment', actions, limits }
}

describe('buildConsentCards', () => {
  it('shows the debit amount for create + limits', () => {
    expect(buildConsentCards(op(['create'], { debitAmount: USD_10 }))).toEqual([
      { label: 'Total amount to debit', value: '$ 10.00' }
    ])
  })

  it('falls back to receiveAmount when there is no debit limit', () => {
    expect(
      buildConsentCards(op(['create'], { receiveAmount: USD_10 }))
    ).toEqual([{ label: 'Total amount to debit', value: '$ 10.00' }])
  })

  it('shows a generic spend message for create without limits', () => {
    expect(buildConsentCards(op(['create']))).toEqual([
      { label: 'Make payments on your behalf' }
    ])
  })

  it('does not surface own-payment view when create is present', () => {
    expect(buildConsentCards(op(['create', 'list']))).toEqual([
      { label: 'Make payments on your behalf' }
    ])
    expect(buildConsentCards(op(['create', 'read', 'list']))).toEqual([
      { label: 'Make payments on your behalf' }
    ])
    expect(
      buildConsentCards(op(['create', 'read'], { debitAmount: USD_10 }))
    ).toEqual([{ label: 'Total amount to debit', value: '$ 10.00' }])
  })

  it('gives read-all its own card', () => {
    expect(buildConsentCards(op(['read-all']))).toEqual([
      { label: 'View all payments on your account' }
    ])
  })

  it('does not surface list-all (rejected w/ 403 upstream)', () => {
    expect(buildConsentCards(op(['list-all']))).toEqual([])
  })

  it('lets a cross-account view absorb the own-payment view (list + read-all)', () => {
    expect(buildConsentCards(op(['list', 'read-all']))).toEqual([
      { label: 'View all payments on your account' }
    ])
  })

  it('shows spend and cross-account view together, without own-view (create + list + read-all)', () => {
    expect(buildConsentCards(op(['create', 'list', 'read-all']))).toEqual([
      { label: 'Make payments on your behalf' },
      { label: 'View all payments on your account' }
    ])
  })

  it('pairs the amount card with a cross-account view (create + limit + read-all)', () => {
    expect(
      buildConsentCards(op(['create', 'read-all'], { debitAmount: USD_10 }))
    ).toEqual([
      { label: 'Total amount to debit', value: '$ 10.00' },
      { label: 'View all payments on your account' }
    ])
  })

  it('shows a solitary list card', () => {
    expect(buildConsentCards(op(['list']))).toEqual([
      { label: 'View a list of your payments' }
    ])
  })

  it('merges a solitary read + list into one card', () => {
    expect(buildConsentCards(op(['read', 'list']))).toEqual([
      { label: 'View your payments' }
    ])
  })

  it('shows no card for a lone read (handled as an identity request)', () => {
    expect(buildConsentCards(op(['read']))).toEqual([])
  })

  it('ignores complete and unsupported actions', () => {
    expect(buildConsentCards(op(['create', 'complete']))).toEqual([
      { label: 'Make payments on your behalf' }
    ])
  })

  it('returns no cards for a grant with no supported actions', () => {
    expect(buildConsentCards(op([]))).toEqual([])
  })
})
