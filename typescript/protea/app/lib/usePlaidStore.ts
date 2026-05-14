import { create } from 'zustand'

// usePlaidStore holds the /plaid route's client-only state: the current Link
// token, the live link summary echoed back from the backend, the last raw
// response from a product call, and the most recent error string for the
// debug panel.
//
// Convention mirrors useScaffoldStore.ts: a State + Actions split with a
// frozen `initialState` constant so `reset()` is a one-liner.

/**
 * Mirror of the server-side PlaidProduct union (intentionally duplicated to
 * avoid pulling a `.server.ts` import into client-bundled code).
 */
export type PlaidProduct =
  | 'accounts'
  | 'auth'
  | 'balance'
  | 'identity'
  | 'transactions'

/** Snapshot of the last response per product, for the debug panel. */
export type LastResponsesMap = Partial<Record<PlaidProduct, unknown>>

interface PlaidStoreState {
  /** Active Plaid Link token; null when none is in flight. */
  linkToken: string | null

  /** Item id returned by the most recent successful exchange. */
  itemId: string | null

  /** Friendly institution name for the linked item, if any. */
  institutionName: string | null

  /** ISO timestamp recorded when the link was established. */
  linkedAt: string | null

  /** Last error surfaced to the user (mirrors snackbar copy). */
  lastError: string | null

  /** Last product fetch by product key. Keyed → product call. */
  lastResponses: LastResponsesMap

  /**
   * Most recently fetched product — the page renders only this card so the
   * results don't stack on top of each other when the user pokes multiple
   * endpoints in a row.
   */
  activeProduct: PlaidProduct | null

  /** True while Plaid Link is opening or a public-token exchange is in flight. */
  isLinking: boolean
}

interface PlaidStoreActions {
  setLinkToken: (token: string | null) => void
  setLinked: (input: {
    itemId: string
    institutionName?: string | null
    linkedAt?: string | null
  }) => void
  clearLinked: () => void
  setLastResponse: (product: PlaidProduct, response: unknown) => void
  setActiveProduct: (product: PlaidProduct | null) => void
  setLastError: (message: string | null) => void
  setIsLinking: (linking: boolean) => void
  reset: () => void
}

const initialState: PlaidStoreState = {
  linkToken: null,
  itemId: null,
  institutionName: null,
  linkedAt: null,
  lastError: null,
  lastResponses: {},
  activeProduct: null,
  isLinking: false
}

export const usePlaidStore = create<PlaidStoreState & PlaidStoreActions>()(
  (set) => ({
    ...initialState,

    setLinkToken: (linkToken) => set({ linkToken }),

    setLinked: ({ itemId, institutionName, linkedAt }) =>
      set({
        itemId,
        institutionName: institutionName ?? null,
        linkedAt: linkedAt ?? new Date().toISOString(),
        linkToken: null,
        isLinking: false,
        lastError: null
      }),

    clearLinked: () =>
      set({
        itemId: null,
        institutionName: null,
        linkedAt: null,
        lastResponses: {},
        activeProduct: null
      }),

    setLastResponse: (product, response) =>
      set((state) => ({
        lastResponses: { ...state.lastResponses, [product]: response },
        activeProduct: product
      })),

    setActiveProduct: (activeProduct) => set({ activeProduct }),

    setLastError: (lastError) => set({ lastError }),

    setIsLinking: (isLinking) => set({ isLinking }),

    reset: () => set(initialState)
  })
)
