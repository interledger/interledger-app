import { create } from 'zustand'

interface FormState {
  step: number
  id: string
  formData: string
}

interface FormActions {
  stepForward: () => void
  stepBack: () => void
  setFormData: () => void
  reset: () => void
}

const formInitialState = {
  step: 0,
  id: '',
  formData: ''
}

export const useFormStore = create<FormState & FormActions>((set) => ({
  ...formInitialState,
  stepForward: () => set((state) => ({ step: state.step + 1 })),
  stepBack: () =>
    set((state) => {
      if (state.step == 0) return { ...formInitialState }
      return { step: state.step - 1 }
    }),
  setFormData: () => set((state) => ({ formData: state.formData })),
  reset: () => set((state) => ({ ...formInitialState }))
}))
