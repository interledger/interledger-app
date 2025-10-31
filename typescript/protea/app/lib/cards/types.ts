import type { SerializeFrom } from '@remix-run/node'
import type { Card } from '~/generated/connect/backend/v1/backend_pb'

export type SerializedCard = SerializeFrom<Card>

/* Unmasked card data response from the card processor */
export type CardProcessorSensitiveDataResponse = {
  /** Unmasked PAN, ex: 1234111122223333 */
  Pan: string
  /** Expiry date, ex: 12/2025 */
  ExpiryDate: string
  /** CVC2, ex: 123 */
  Cvc2: string
}

export type HttpMethod = 'GET' | 'POST' | 'PUT' | 'DELETE' | 'PATCH'
