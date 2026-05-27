import { create } from 'zustand'

interface DialPadState {
  amountValue: string
  assetCode: string
}

interface DialPadActions {
  setAmountValue: (amount: string) => void
  setAssetCode: (assetCode: string) => void
  reset: () => void
}

const dialPadInitialState: DialPadState = {
  amountValue: '0',
  assetCode: 'USD',
}

export const useDialPadStore = create<DialPadState & DialPadActions>()((set) => ({
  ...dialPadInitialState,
  setAmountValue: (amountValue) => set({ amountValue }),
  setAssetCode: (assetCode) => set({ assetCode }),
  reset: () => set({ ...dialPadInitialState }),
}))