import { v4 } from 'uuid'
import { create } from 'zustand'
import type {
  PublicWalletInfo,
  SearchResult
} from '~/generated/protobuf-ts/backend/v1/backend'

export enum PayStep {
  SEARCH,
  AMOUNT,
  CONFIRM
}

interface PayState {
  step: PayStep
  // Search page
  searchTerm: string
  results: SearchResult[]
  address: SearchResult | null
  // Amount page
  publicWalletInfo: PublicWalletInfo | null
  quoteId: string
  accountId: string
  amount: string
  displayAmount: string
  note: string
  // ThreeDSecure page
  idempotencyKey: string
}

interface PayActions {
  setStep: (step: PayStep) => void
  setSearchTerm: (term: string) => void
  setResults: (results: SearchResult[]) => void
  setAddress: (address: SearchResult) => void
  setAmount: (amount: string) => void
  setNote: (note: string) => void
  setAccountId: (id: string) => void
  setQuoteId: (id: string) => void
  reset: () => void
}

const payInitialState = {
  step: PayStep.SEARCH,
  searchTerm: '',
  address: null,
  publicWalletInfo: null,
  quoteId: '',
  results: [],
  accountId: '',
  amount: '',
  displayAmount: '$ 0.00',
  note: '',
  idempotencyKey: v4()
}

export const usePayStore = create<PayState & PayActions>((set) => ({
  ...payInitialState,
  setStep: (step) => set((state) => ({ step: step })),
  setSearchTerm: (term) => set((state) => ({ searchTerm: term })),
  setResults: (results) => set((state) => ({ results: results })),
  setAddress: (address) => set((state) => ({ address: address })),
  setAmount: (amount) =>
    set((state) => ({
      amount: amount,
      displayAmount: formatMoney(parseFloat(amount || '0'))
    })),
  setNote: (note) => set((state) => ({ note })),
  setAccountId: (id) => set((state) => ({ accountId: id })),
  setQuoteId: (id) => set((state) => ({ quoteId: id })),
  reset: () => set((state) => ({ ...payInitialState }))
}))

// sendPaymentPointer: string; - don't need this in the store
// receivePaymentPointer: string; - this is determined from the search result
// amount?: Amount;
// expiresAt?: Timestamp;
// description: string;
// sendLinkedAccount?: string;

const formatMoney = (value: number): string => {
  console.log('formatMoney', value)
  return `$ ${value.toFixed(2)}`
}
