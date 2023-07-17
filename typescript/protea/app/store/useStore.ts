import { create } from 'zustand'
import type { PaySlice } from '~/store/pay'
import { createPaySlice } from '~/store/pay'

export const useStore = create<PaySlice>()((...a) => ({
  ...createPaySlice(...a)
}))
