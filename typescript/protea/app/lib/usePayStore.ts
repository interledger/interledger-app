import { create } from 'zustand'

export enum PayStep {
  AMOUNT,
  CONFIRM
}

interface PayState {
  step: PayStep
}

interface PayActions {
  setStep: (step: PayStep) => void
  stepBack: () => void
  reset: () => void
}

const payInitialState = {
  step: PayStep.AMOUNT
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
  reset: () => set((state) => ({ ...payInitialState }))
}))
