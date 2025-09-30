import { create } from 'zustand'
import { GateHubState, GateHubStore, StorableCard } from '../types'
import { GATEHUB_APP_ID, GATEHUB_CARD_APP_ID, GATEHUB_MANAGED_USER_UUID } from '../do-not-commit'

// Initial state
const initialState: GateHubState = {
  // todo: find a way to load those dinamically
  // ???? security issue?
  auth: {
    cardAppId: GATEHUB_CARD_APP_ID,
    appId: GATEHUB_APP_ID,
  },
  user: {
    managedUserUuid: GATEHUB_MANAGED_USER_UUID
  },
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
