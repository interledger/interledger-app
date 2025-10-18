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

/* Card PIN response from the card processor */
export type CardProcessorPinResponse = {
  /** Card PIN, ex: 1234 */
  pin: string
}

export type GateHubState = {
  auth: {
    /** Application identifier for GateHub card integration */
    cardAppId: string
    /** Application identifier for GateHub card integration */
    appId: string
  }
  user: {
    /** Unique identifier for the customer in Dinit system */
    customerId?: string
    /** User identifier for accessing user-specific data */
    userUuid?: string
    /** Managed user UUID for GateHub operations */
    managedUserUuid?: string
  }
  card: {
    /** Currently active card identifier */
    activeCardId?: string
    /** List of user's cards for quick access */
    cards?: Record<string, StorableCard>
  }
}

type GateHubActions = {
  // Card actions
  card: {
    setActiveCardId: (cardId: string) => void
    setCards: (cards: StorableCard[]) => void
    updateActiveCard: (card: StorableCard) => void
  }
}

export type GateHubStore = GateHubState & {
  actions: GateHubActions
}

/* GateHub Client types */
export type GateHubRequestOptions = {
  /** Include x-gatehub-app-id header */
  includeAppId?: boolean
  /** Include x-gatehub-card-app-id header */
  includeCardAppId?: boolean
  /** Include x-gatehub-managed-user-uuid header */
  includeManagedUserUuid?: boolean
  /** Include Authorization header with session token */
  includeSessionToken?: boolean
  /** Include x-gatehub-timestamp header */
  includeTimestamp?: boolean
  /** Include x-gatehub-signature header */
  includeSignature?: boolean
  /** Additional custom headers */
  customHeaders?: Record<string, string>
}

export type HttpMethod = 'GET' | 'POST' | 'PUT' | 'DELETE' | 'PATCH'

export type GateHubApiResponse<T = any> = {
  data?: T
  error?: string
  status: number
  headers: Headers
}
