import { create } from 'zustand'
import type { CardsStore, StorableCard } from '../types'

export const useCardsStore = create<CardsStore>((set, get) => ({
  cards: {},

  setCards: (cards: StorableCard[]) =>
    set(() => ({
      cards: cards.reduce((acc, c) => ({ ...acc, [c.id]: c }), {})
    }))
}))
