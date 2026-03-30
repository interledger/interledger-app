import { create } from 'zustand'
import type { CardsStore, SerializedCard, StorableCard } from './types'

/**
 * Convert Card object to StorableCard format for the store
 */
export function toStorableCard(card: SerializedCard): StorableCard {
  return {
    ...card,
    unmaskedPan: null,
    cvc2: null
  }
}

export const formatPan = (pan: string) => {
  return pan.replace(/(.{4})/g, '$1 ').trim()
}

export const panLastFour = (pan: string) => {
  return pan.slice(-4)
}

export const useCardsStore = create<CardsStore>((set) => ({
  cards: {},
  areCardsFetched: false,

  setCards: (cards: StorableCard[]) => {
    set(() => ({
      cards: cards.reduce((acc, c) => ({ ...acc, [c.id]: c }), {}),
      areCardsFetched: true
    }))
  }
}))
