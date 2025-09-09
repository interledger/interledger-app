export interface GateHubAuthContext {
  /** Application identifier for GateHub card integration */
  cardAppId: string
}

export interface GateHubUserContext {
  /** Unique identifier for the customer in Dinit system */
  customerId?: string
  /** User identifier for accessing user-specific data */
  userUuid?: string
  /** Managed user UUID for GateHub operations */
  managedUserUuid?: string
}

export type StorableCard = {
  id: string
  nameOnCard: string
  maskedPan: string
  expiryDate: string
  status: number
  lockLevel: number
}
export type GateHubCardContext = {
  /** Currently active card identifier */
  activeCardId?: string
  /** List of user's cards for quick access */
  cards?: StorableCard[]
}


export type GateHubContextState = {
  auth: GateHubAuthContext
  user: GateHubUserContext
  card: GateHubCardContext
  token: string | null
}

export type GateHubActions = {
  // Card actions
  setActiveCardId: (cardId: string) => void
  addCard: (card: StorableCard) => void
  removeCard: (cardId: string) => void
  setCards: (cards: StorableCard[]) => void

  // Token actions
  setToken: (token: string) => void

  // Reset actions
  reset: () => void
}

export type GateHubStore = GateHubContextState & GateHubActions

export type GateHubRequestOptions = {
  /** Include x-gatehub-card-app-id header */
  includeCardAppId?: boolean
  /** Include x-gatehub-managed-user-uuid header */
  includeManagedUserUuid?: boolean
  /** Include Authorization header with session token */
  includeSessionToken?: boolean
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