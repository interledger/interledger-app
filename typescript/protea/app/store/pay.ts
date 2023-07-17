import { v4 } from 'uuid'
import type { StateCreator } from 'zustand'
import type { SearchResult } from '~/generated/protobuf-ts/backend/v1/backend'
import type { FormattedLinkedAccount } from '~/lib/wallet.server'

export enum PayStep {
  SEARCH,
  AMOUNT,
  CONFIRM
}

export interface PaySlice {
  step: PayStep
  searchTerm: string
  results: SearchResult[]
  address: SearchResult | null
  toLinkedAccountId: string
  linkedAccounts: FormattedLinkedAccount[]
  amount: number
  note: string
  idempotencyKey: string
  setStep: (step: PayStep) => void
  setSearchTerm: (term: string) => void
  setResults: (results: SearchResult[]) => void
  setAddress: (address: SearchResult) => void
  setAmount: (amount: number) => void
  reset: () => void
}

const payInitialState = {
  step: PayStep.SEARCH,
  searchTerm: '',
  address: null,
  results: [],
  toLinkedAccountId: '',
  linkedAccounts: [],
  amount: 0,
  note: '',
  idempotencyKey: v4()
}

export const createPaySlice: StateCreator<PaySlice, [], [], PaySlice> = (
  set
) => ({
  ...payInitialState,
  setStep: (step) => set((state) => ({ step: step })),
  setSearchTerm: (term) => set((state) => ({ searchTerm: term })),
  setResults: (results) => set((state) => ({ results: results })),
  setAddress: (address) => set((state) => ({ address: address })),
  setAmount: (amount) => set((state) => ({ amount })),
  reset: () => set((state) => ({ ...payInitialState }))
})
