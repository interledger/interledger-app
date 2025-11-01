import { create } from 'zustand'

interface PinChangePopupStore {
  isOpen: boolean
  pendingCallback: ((pin: string) => void) | null
  openPopup: (callback: (pin: string) => void) => void
  closePopup: () => void
  handleSubmit: (pin: string) => void
}

export const usePinChangePopupStore = create<PinChangePopupStore>(
  (set, get) => ({
    isOpen: false,
    pendingCallback: null,

    openPopup: (callback: (pin: string) => void) => {
      set({ isOpen: true, pendingCallback: callback })
    },

    closePopup: () => {
      set({ isOpen: false, pendingCallback: null })
    },

    handleSubmit: (pin: string) => {
      const { pendingCallback, closePopup } = get()
      if (pendingCallback) {
        pendingCallback(pin)
      }
      closePopup()
    }
  })
)
