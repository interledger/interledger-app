/* Store models */
export type StorableCard = {
  id: string
  nameOnCard: string
  maskedPan: string
  expiryDate: string
  status: number
  lockLevel: number
}

/* Store state types */
interface GateHubAuthState {
  /** Application identifier for GateHub card integration */
  cardAppId: string
  /** Application identifier for GateHub card integration */
  appId: string
}

interface GateHubUserState {
  /** Unique identifier for the customer in Dinit system */
  customerId?: string
  /** User identifier for accessing user-specific data */
  userUuid?: string
  /** Managed user UUID for GateHub operations */
  managedUserUuid?: string
}

type GateHubCardState = {
  /** Currently active card identifier */
  activeCardId?: string
  /** List of user's cards for quick access */
  cards?: StorableCard[]
}

export type GateHubState = {
  auth: GateHubAuthState
  user: GateHubUserState
  card: GateHubCardState
  token: string | null
}

type GateHubActions = {
  // Card actions
  card: {
    setActiveCardId: (cardId: string) => void
    addCard: (card: StorableCard) => void
    removeCard: (cardId: string) => void
    setCards: (cards: StorableCard[]) => void
  }

  // Token actions
  token: {
    setToken: (token: string) => void
  }

  // Reset actions
  reset: () => void
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
