import {
  Card,
  CardLockLevel,
  CardStatus,
  CardStatusReasonCode
} from '~/generated/connect/backend/v1/backend_pb'

/**
 * Mock card data for testing and development
 */
export const mockCards: Card[] = [
  new Card({
    id: 'card-1',
    nameOnCard: 'John Doe',
    maskedPan: '**** **** **** 1234',
    status: CardStatus.ACTIVE,
    expiryDate: '12/25',
    statusReasonCode: CardStatusReasonCode.UNKNOWN,
    lockLevel: CardLockLevel.UNKNOWN
  }),

  new Card({
    id: 'card-2',
    nameOnCard: 'Jane Smith',
    maskedPan: '**** **** **** 5678',
    status: CardStatus.BLOCKED,
    expiryDate: '08/26',
    statusReasonCode: CardStatusReasonCode.CLIENT_REQUESTED_LOCK,
    lockLevel: CardLockLevel.CLIENT
  }),

  new Card({
    id: 'card-3',
    nameOnCard: 'Ion Macelarul',
    maskedPan: '**** **** **** 9012',
    status: CardStatus.IN_CREATION,
    expiryDate: '03/27',
    statusReasonCode: CardStatusReasonCode.UNKNOWN,
    lockLevel: CardLockLevel.UNKNOWN
  })
]
