import { create } from 'zustand'
import type { PendingConfirmation } from './mocks/confirmations'

interface PendingConfirmationsStore {
  // State
  pendingConfirmations: any[]
  hasFetched: boolean
  timeoutIds: Map<string, NodeJS.Timeout>
  // Actions
  setHasFetched: (hasFetched: boolean) => void
  setPendingConfirmations: (pendingConfirmations: PendingConfirmation[]) => void
  addPendingConfirmation: (pendingConfirmation: PendingConfirmation) => void
  removeConfirmation: (transactionId: string) => void
  clearTimeouts: () => void
}

const createPendingConfirmationTimeout = (
  pendingConfirmation: PendingConfirmation
): number => {
  const purchaseTime = Number(pendingConfirmation.purchaseDate)
  const timeoutMs = Number(pendingConfirmation.timeout) * 1000
  const expiryTime = purchaseTime + timeoutMs
  const remainingMs = expiryTime - Date.now()

  return remainingMs
}

const formatPendingConfirmations = (confirmation: PendingConfirmation) => ({
  transactionId: confirmation.transactionId,
  merchantName: confirmation.merchantName,
  purchaseDate: confirmation.purchaseDate,
  timeout: confirmation.timeout,
  formattedDate: new Date(Number(confirmation.purchaseDate)).toLocaleDateString(
    'en-US',
    {
      month: 'long',
      day: 'numeric',
      year: 'numeric'
    }
  ),
  formattedTime: new Date(Number(confirmation.purchaseDate)).toLocaleTimeString(
    'en-US',
    {
      hour: 'numeric',
      minute: '2-digit',
      hour12: true
    }
  ),
  formattedAmount: `${parseFloat(confirmation.purchaseAmount).toFixed(2)} ${
    confirmation.purchaseCurrency
  }`
})

const usePendingConfirmationsStore = create<PendingConfirmationsStore>(
  (set, get) => ({
    pendingConfirmations: [],
    hasFetched: false,
    setHasFetched: (hasFetched: boolean) => set({ hasFetched }),
    timeoutIds: new Map(),

    setPendingConfirmations: (pendingConfirmations: PendingConfirmation[]) => {
      const { clearTimeouts, removeConfirmation } = get()

      clearTimeouts()

      // Schedule timeouts for each confirmation
      const newTimeoutIds = new Map<string, NodeJS.Timeout>()
      pendingConfirmations.forEach((confirmation) => {
        const remainingMs = createPendingConfirmationTimeout(confirmation)

        if (remainingMs > 0) {
          const timeoutId = setTimeout(() => {
            removeConfirmation(confirmation.transactionId)
          }, remainingMs)
          newTimeoutIds.set(confirmation.transactionId, timeoutId)
        } else {
          // Already expired, remove immediately
          removeConfirmation(confirmation.transactionId)
        }
      })

      const formattedPendingConfirmations = pendingConfirmations.map(
        formatPendingConfirmations
      )

      console.log(
        '[usePendingConfirmationsStore] 🐳 setPendingConfirmations called'
      )
      set({
        pendingConfirmations: formattedPendingConfirmations,
        timeoutIds: newTimeoutIds,
        hasFetched: true
      })
    },

    addPendingConfirmation: (pendingConfirmation: PendingConfirmation) => {
      const { timeoutIds, removeConfirmation } = get()

      console.log(
        '[usePendingConfirmationsStore] 🐳 addPendingConfirmation called',
        pendingConfirmation
      )
      set((state) => ({
        pendingConfirmations: [
          ...state.pendingConfirmations,
          pendingConfirmation
        ]
      }))

      const remainingMs = createPendingConfirmationTimeout(pendingConfirmation)

      if (remainingMs > 0) {
        const timeoutId = setTimeout(() => {
          removeConfirmation(pendingConfirmation.transactionId)
        }, remainingMs)
        timeoutIds.set(pendingConfirmation.transactionId, timeoutId)
        set({ timeoutIds: new Map(timeoutIds) })
      }
    },

    removeConfirmation: (transactionId: string) => {
      const { timeoutIds } = get()

      // Clear timeout if exists
      const timeoutId = timeoutIds.get(transactionId)
      if (timeoutId) {
        clearTimeout(timeoutId)
        timeoutIds.delete(transactionId)
      }

      // Remove confirmation
      set((state) => ({
        pendingConfirmations: state.pendingConfirmations.filter(
          (c) => c.transactionId !== transactionId
        ),
        timeoutIds: new Map(timeoutIds)
      }))
    },

    clearTimeouts: () => {
      const { timeoutIds } = get()
      timeoutIds.forEach((id) => clearTimeout(id))
      set({ timeoutIds: new Map() })
    }
  })
)

export const usePendingConfirmations = () => {
  const pendingConfirmations = usePendingConfirmationsStore(
    (state) => state.pendingConfirmations
  )
  const initializePendingConfirmations = usePendingConfirmationsStore(
    (state) => state.setPendingConfirmations
  )
  const addPendingConfirmation = usePendingConfirmationsStore(
    (state) => state.addPendingConfirmation
  )
  const removeConfirmation = usePendingConfirmationsStore(
    (state) => state.removeConfirmation
  )
  const hasFetched = usePendingConfirmationsStore((state) => state.hasFetched)
  const clearTimeouts = usePendingConfirmationsStore((state) => state.clearTimeouts)

  return {
    hasFetched,
    pendingConfirmations,
    initializePendingConfirmations,
    addPendingConfirmation,
    removeConfirmation,
    clearTimeouts
  }
}
