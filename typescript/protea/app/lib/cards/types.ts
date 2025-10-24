/* Store models */
export type StorableCard = {
  id: string
  nameOnCard: string
  maskedPan: string
  unmaskedPan: string | null
  expiryDate: string
  status: number
  lockLevel: string
  cvc2: string | null
}

/* Unmasked card data response from the card processor */
export type CardProcessorSensitiveDataResponse = {
  /** Unmasked PAN, ex: 1234111122223333 */
  Pan: string
  /** Expiry date, ex: 12/2025 */
  ExpiryDate: string
  /** CVC2, ex: 123 */
  Cvc2: string
}

export type CardsState = {
  /** Currently active card identifier */
  activeCardId?: string
  /** List of user's cards for quick access */
  cards?: Record<string, StorableCard>
}

type CardsActions = {
  setActiveCardId: (cardId: string) => void
  setCards: (cards: StorableCard[]) => void
  updateActiveCard: (card: StorableCard) => void
}

export type CardsStore = CardsState & CardsActions

export type HttpMethod = 'GET' | 'POST' | 'PUT' | 'DELETE' | 'PATCH'
