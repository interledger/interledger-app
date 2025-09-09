import { create } from 'zustand'
import { GateHubState, GateHubStore, StorableCard } from '../types'

// Initial state
const initialState: GateHubState = {
  auth: {
    cardAppId: ''
  },
  user: {},
  card: {},
  token: null
}

export const useGateHubStore = create<GateHubStore>((set, get) => ({
  ...initialState,

  actions: {
    card: {
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
        }))
    },
    token: {
      setToken: (token: string | null) =>
        set(() => ({
          token: token
        }))
    },

    reset: () => set(() => ({ ...initialState }))
  }
}))
