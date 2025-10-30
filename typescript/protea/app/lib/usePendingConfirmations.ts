import { create } from 'zustand'
import type { PendingThreeDSConfirmation } from '~/generated/connect/backend/v1/backend_pb'

export interface StorablePendingConfirmation {
  transactionId: string
  merchantName: string
  purchaseAmount: string
  purchaseCurrency: string
  formattedDate: string
  formattedTime: string
  purchaseDate: string
  timeout: string
}

interface PendingConfirmationsStore {
  pendingConfirmations: StorablePendingConfirmation[]
  hasFetched: boolean
  timeoutIds: Map<string, NodeJS.Timeout>
  // Actions
  setHasFetched: (hasFetched: boolean) => void
  setPendingConfirmations: (
    pendingConfirmations: PendingThreeDSConfirmation[]
  ) => void
  addPendingConfirmation: (
    pendingConfirmation: PendingThreeDSConfirmation
  ) => void
  removeConfirmation: (transactionId: string) => void
  clearTimeouts: () => void
}

function dedupe(
  arr: StorablePendingConfirmation[]
): StorablePendingConfirmation[] {
  const seen = new Set()

  return arr.filter((item) => {
    if (seen.has(item.transactionId)) return false
    seen.add(item.transactionId)
    return true
  })
}

const createPendingConfirmationTimeout = (
  pendingConfirmation: PendingThreeDSConfirmation
): number => {
  const purchaseTime = new Date(pendingConfirmation.purchaseDate).valueOf()
  const timeoutMs = Number(pendingConfirmation.timeout) * 1000
  const expiryTime = purchaseTime + timeoutMs
  const remainingMs = expiryTime - Date.now()

  return remainingMs
}

const formatPendingConfirmations = (
  confirmation: PendingThreeDSConfirmation
): StorablePendingConfirmation => ({
  transactionId: confirmation.transactionId,
  merchantName: confirmation.merchantName,
  purchaseAmount: confirmation.purchaseAmount,
  purchaseCurrency: confirmation.purchaseCurrency,
  purchaseDate: confirmation.purchaseDate,
  timeout: confirmation.timeout,
  formattedDate: new Date(confirmation.purchaseDate).toLocaleDateString(
    'en-US',
    {
      month: 'long',
      day: 'numeric',
      year: 'numeric'
    }
  ),
  formattedTime: new Date(confirmation.purchaseDate).toLocaleTimeString(
    'en-US',
    {
      hour: 'numeric',
      minute: '2-digit',
      hour12: true
    }
  )
})

const usePendingConfirmationsStore = create<PendingConfirmationsStore>(
  (set, get) => ({
    pendingConfirmations: [],
    hasFetched: false,
    setHasFetched: (hasFetched: boolean) => set({ hasFetched }),
    timeoutIds: new Map(),

    setPendingConfirmations: (
      pendingConfirmations: PendingThreeDSConfirmation[]
    ) => {
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

      set({
        pendingConfirmations: formattedPendingConfirmations,
        timeoutIds: newTimeoutIds,
        hasFetched: true
      })
    },

    addPendingConfirmation: (
      pendingConfirmation: PendingThreeDSConfirmation
    ) => {
      const { timeoutIds, removeConfirmation } = get()

      const remainingMs = createPendingConfirmationTimeout(pendingConfirmation)

      if (remainingMs > 0) {
        const timeoutId = setTimeout(() => {
          removeConfirmation(pendingConfirmation.transactionId)
        }, remainingMs)
        timeoutIds.set(pendingConfirmation.transactionId, timeoutId)
        set({ timeoutIds: new Map(timeoutIds) })
      }

      set((state) => {
        const pendingConfirmations = [
          ...state.pendingConfirmations,
          formatPendingConfirmations(pendingConfirmation)
        ]

        return {
          pendingConfirmations: dedupe(pendingConfirmations)
        }
      })
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
  const clearTimeouts = usePendingConfirmationsStore(
    (state) => state.clearTimeouts
  )

  return {
    hasFetched,
    pendingConfirmations,
    initializePendingConfirmations,
    addPendingConfirmation,
    removeConfirmation,
    clearTimeouts
  }
}
