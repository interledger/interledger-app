import type { StateCreator } from 'zustand'
import { create } from 'zustand'
import type { SearchResult } from '~/generated/protobuf-ts/backend/v1/backend'

export enum PayStep {
  SEARCH,
  AMOUNT,
  CONFIRM
}

interface PaySlice {
  step: PayStep
  searchTerm: string
  results: SearchResult[]
  address: SearchResult | null
  amount: number
  note: string
  setStep: (step: PayStep) => void
  setSearchTerm: (term: string) => void
  setResults: (results: SearchResult[]) => void
  setAddress: (address: SearchResult) => void
  setAmount: (amount: number) => void
}

export const createPaySlice: StateCreator<PaySlice, [], [], PaySlice> = (
  set
) => ({
  step: PayStep.SEARCH,
  searchTerm: '',
  address: null,
  results: [],
  amount: 0,
  note: '',
  setStep: (step) => set((state) => ({ step: step })),
  setSearchTerm: (term) => set((state) => ({ searchTerm: term })),
  setResults: (results) => set((state) => ({ results: results })),
  setAddress: (address) => set((state) => ({ address: address })),
  setAmount: (amount) => set((state) => ({ amount }))
})

export const useStore = create<PaySlice>()((...a) => ({
  ...createPaySlice(...a)
}))
