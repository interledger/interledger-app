import { create } from 'zustand'
import { CardProcessorSensitiveDataResponse, GateHubState, GateHubStore, StorableCard } from '../types'
import { GATEHUB_APP_ID, GATEHUB_CARD_APP_ID, GATEHUB_MANAGED_USER_UUID } from '../do-not-commit'

// Initial state
const initialState: GateHubState = {
  auth: {
    cardAppId: GATEHUB_CARD_APP_ID,
    appId: GATEHUB_APP_ID,
  },
  user: {
    managedUserUuid: GATEHUB_MANAGED_USER_UUID
  },
  card: {},
}

export const useGateHubStore = create<GateHubStore>((set, get) => ({
  ...initialState,

  actions: {
    card: {
      setActiveCardId: (cardId: string) =>
        set((state) => ({
          card: { ...state.card, activeCardId: cardId }
        })),
      // addCard: (card: StorableCard) =>
      //   set((state) => ({
      //     card: {
      //       ...state.card,
      //       cards: [...(state.card.cards ?? []), card]
      //     }
      //   })),
      // removeCard: (cardId: string) =>
      //   set((state) => ({
      //     card: {
      //       ...state.card,
      //       cards: (state.card.cards ?? []).filter(
      //         (c: StorableCard) => c.id !== cardId
      //       )
      //     }
      //   })),
      setCards: (cards: StorableCard[]) =>
        set((state) => ({
          card: {
            ...state.card,
            cards: cards.reduce((acc, c) => ({ ...acc, [c.id]: c }), {})
          }
        })),
        setCardSensitiveData: (sensitiveData: CardProcessorSensitiveDataResponse) => {
          const { card: { activeCardId, cards } } = get()
          if (!activeCardId) return;

          const activeCard = cards?.[activeCardId] as StorableCard
          const revealedCard: StorableCard = {
            ...activeCard,
            unmaskedPan: sensitiveData.Pan,
            cvc2: sensitiveData.Cvc2,
            expiryDate: sensitiveData.ExpiryDate
          }
          const newCards: Record<string, StorableCard> = { ...cards, [activeCardId]: revealedCard }

          set((state) => ({
            card: {
              ...state.card,
              cards: newCards
            }
          }))
        }
    },
  }
}))
