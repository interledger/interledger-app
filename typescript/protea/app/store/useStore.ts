import { create } from 'zustand'
import type { FormattedLinkedAccount } from '~/lib/wallet.server'

export interface State {
  accounts: FormattedLinkedAccount[]
  setAccounts: (accounts: FormattedLinkedAccount[]) => void
  reset: () => void
}

const initialState = {
  accounts: []
}

export const useStore = create<State>((set) => ({
  ...initialState,
  setAccounts: (accounts) => set((state) => ({ accounts })),
  reset: () => set((state) => ({ ...initialState }))
}))
