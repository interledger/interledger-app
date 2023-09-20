import { create } from 'zustand'

export enum PayStep {
  UNKNOWN,
  AMOUNT,
  CONFIRM
}

interface PayState {
  step: PayStep
  amount: string
  displayAmount: string
}

interface PayActions {
  setStep: (step: PayStep) => void
  stepBack: () => void
  setAmount: (amount: string) => void
  reset: () => void
}

const payInitialState = {
  step: PayStep.UNKNOWN,
  amount: '',
  displayAmount: '$ 0.00'
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
  setAmount: (amount) =>
    set((state) => ({
      amount: amount,
      displayAmount: formatMoney(parseFloat(amount || '0'))
    })),
  reset: () => set((state) => ({ ...payInitialState }))
}))

const formatMoney = (value: number): string => {
  return `$ ${value.toFixed(2)}`
}
