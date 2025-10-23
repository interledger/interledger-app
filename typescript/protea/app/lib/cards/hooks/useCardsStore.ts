import { create } from 'zustand'
import type { CardsStore, StorableCard } from '../types'

export const useCardsStore = create<CardsStore>((set, get) => ({
  activeCardId: undefined,
  cards: {},

  setActiveCardId: (cardId: string) => {
    set((state) => ({ activeCardId: cardId }))
  },
  setCards: (cards: StorableCard[]) =>
    set(() => ({
      cards: cards.reduce((acc, c) => ({ ...acc, [c.id]: c }), {})
    })),
  updateActiveCard: (card: StorableCard) =>
    set((state) => ({ cards: { ...state.cards, [card.id]: card } }))
}))
