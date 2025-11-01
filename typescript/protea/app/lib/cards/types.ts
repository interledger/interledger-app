/* Store models */

import type { SerializeFrom } from '@remix-run/node'
import type {
  Card,
  CardLockLevel,
  CardStatus,
  CardStatusReasonCode,
  CardType
} from '~/generated/connect/backend/v1/backend_pb'

export type SerializedCard = SerializeFrom<Card>

export type StorableCard = SerializedCard & {
  id: string
  nameOnCard: string
  maskedPan: string
  expiryDate: string
  status: CardStatus
  statusReasonCode: CardStatusReasonCode
  lockLevel: CardLockLevel
  type: CardType
  productCode: string
  unmaskedPan: string | null
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
  cards?: Record<string, StorableCard>
  areCardsFetched: boolean
}

type CardsActions = {
  setCards: (cards: StorableCard[]) => void
}

export type CardsStore = CardsState & CardsActions

export type HttpMethod = 'GET' | 'POST' | 'PUT' | 'DELETE' | 'PATCH'
