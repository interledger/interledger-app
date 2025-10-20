import { create } from 'zustand'
import {
  GATEHUB_APP_ID,
  GATEHUB_CARD_APP_ID,
  GATEHUB_MANAGED_USER_UUID
} from '../do-not-commit'
import {
  GateHubState,
  GateHubStore,
  StorableCard
} from '../types'

// Initial state
const initialState: GateHubState = {
  auth: {
    cardAppId: GATEHUB_CARD_APP_ID,
    appId: GATEHUB_APP_ID
  },
  user: {
    managedUserUuid: GATEHUB_MANAGED_USER_UUID
  },
  card: {}
}

export const useGateHubStore = create<GateHubStore>((set, get) => ({
  ...initialState,

  actions: {
    card: {
      setActiveCardId: (cardId: string) => {
        console.log('🔑 Setting active card id', cardId)
        set((state) => ({
          card: { ...state.card, activeCardId: cardId }
        }))
      },
      setCards: (cards: StorableCard[]) =>
        set((state) => ({
          card: {
            ...state.card,
            cards: cards.reduce((acc, c) => ({ ...acc, [c.id]: c }), {})
          }
        })),
      updateActiveCard: (card: StorableCard) =>
        set((state) => ({
          card: {
            ...state.card,
            cards: { ...state.card.cards, [card.id]: card }
          }
        })),
    }
  }
}))
