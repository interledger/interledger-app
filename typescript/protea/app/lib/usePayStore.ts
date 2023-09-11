import { create } from 'zustand'
import type { FormattedLinkedAccount } from '~/data/wallet.server'
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
  address: SearchResult | null
  // Amount page
  publicWalletInfo: PublicWalletInfo | null
  requiresOTP: boolean
  account: FormattedLinkedAccount | null
  amount: string
  displayAmount: string
  note: string
  paymentId: string
}

interface PayActions {
  setStep: (step: PayStep) => void
  stepBack: () => void
  setAddress: (address: SearchResult) => void
  setAmount: (amount: string) => void
  setNote: (note: string) => void
  setPublicWalletInfo: (walletInfo: PublicWalletInfo) => void
  setAccount: (account: FormattedLinkedAccount) => void
  reset: () => void

  // payment engine fields
  setPayment: (id: string, requiresOTP: boolean) => void
}

const payInitialState = {
  step: PayStep.SEARCH,
  address: null,
  publicWalletInfo: null,
  requiresOTP: false,
  account: null,
  amount: '',
  displayAmount: '$ 0.00',
  note: '',
  paymentId: ''
}

export const usePayStore = create<PayState & PayActions>()((set) => ({
  ...payInitialState,
  setStep: (step) => set((state) => ({ step: step })),
  stepBack: () =>
    set((state) => {
      switch (state.step) {
        case PayStep.CONFIRM:
          return { step: PayStep.AMOUNT }
        default:
          return { ...payInitialState }
      }
    }),
  setAddress: (address) => set((state) => ({ address: address })),
  setAmount: (amount) =>
    set((state) => ({
      amount: amount,
      displayAmount: formatMoney(parseFloat(amount || '0'))
    })),
  setPublicWalletInfo: (walletInfo) =>
    set((state) => ({ publicWalletInfo: walletInfo })),
  setNote: (note) => set((state) => ({ note })),
  setAccount: (account) => set((state) => ({ account })),
  reset: () => set((state) => ({ ...payInitialState })),

  // payment engine fields
  setPayment: (id, requiresOTP) =>
    set((state) => ({ paymentId: id, requiresOTP }))
}))

// sendPaymentPointer: string; - don't need this in the store
// receivePaymentPointer: string; - this is determined from the search result
// amount?: Amount;
// expiresAt?: Timestamp;
// description: string;
// sendLinkedAccount?: string;

const formatMoney = (value: number): string => {
  return `$ ${value.toFixed(2)}`
}
