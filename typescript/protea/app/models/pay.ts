import type { StateCreator } from 'zustand/esm'

interface PaySlice {
  amount: number
  setAmount: (amount: number) => void
}

export const createPaySlice: StateCreator<PaySlice, [], [], PaySlice> = (
  set
) => ({
  amount: 0,
  setAmount: (amount) => set((state) => ({ amount }))
})
