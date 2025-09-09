import { create } from 'zustand'
import {
  GateHubContextState,
  GateHubRequestOptions,
  GateHubStore,
  StorableCard
} from '../types'

// Initial state
const initialState: GateHubContextState = {
  auth: {
    cardAppId: ''
  },
  user: {},
  card: {},
  token: null
}

export const useGateHubStore = create<GateHubStore>((set, get) => ({
  ...initialState,

  buildHeaders: (options: GateHubRequestOptions = {}) => {
    const state = get()
    const headers: Record<string, string> = {
      'Content-Type': 'application/json'
    }

    if (options.includeCardAppId !== false && state.auth.cardAppId) {
      headers['x-gatehub-card-app-id'] = state.auth.cardAppId
    }

    if (options.includeManagedUserUuid && state.user.managedUserUuid) {
      headers['x-gatehub-managed-user-uuid'] = state.user.managedUserUuid
    }

    if (options.includeSessionToken && state.token) {
      const token = state.token
      if (token) {
        headers['Authorization'] = `Bearer ${token}`
      }
    }

    if (options.customHeaders) {
      Object.assign(headers, options.customHeaders)
    }

    return headers
  },

  setActiveCardId: (cardId: string) =>
    set((state) => ({
      card: { ...state.card, activeCardId: cardId }
    })),
  addCard: (card: StorableCard) =>
    set((state) => ({
      card: {
        ...state.card,
        cards: [...(state.card.cards ?? []), card]
      }
    })),
  removeCard: (cardId: string) =>
    set((state) => ({
      card: {
        ...state.card,
        cards: (state.card.cards ?? []).filter(
          (c: StorableCard) => c.id !== cardId
        )
      }
    })),
  setCards: (cards: StorableCard[]) =>
    set((state) => ({
      card: {
        ...state.card,
        cards
      }
    })),
  setToken: (token: string | null) =>
    set(() => ({
      token
    })),

  reset: () => set(() => ({ ...initialState }))
}))
