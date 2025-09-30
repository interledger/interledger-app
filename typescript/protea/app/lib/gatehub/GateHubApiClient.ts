import { GATEHUB_API_ENDPOINTS } from '~/lib/gatehub/config/endpoints'
import type { GateHubApiResponse, HttpMethod } from '~/lib/gatehub/types'
import { buildHeaders } from '~/lib/gatehub/utils/headers.builder'

/**
 * Flags to determine which headers to be used for the direct GateHub request
 */
export interface GateHubApiRequirements {
  appId?: boolean
  cardAppId?: boolean
  managedUser?: boolean
  sessionToken?: boolean
  timestamp?: boolean
  signature?: boolean
}

export interface GateHubApiRequest {
  endpoint: string
  method: HttpMethod
  body?: any
  queryParams?: Record<string, string | number | boolean>
  requires?: GateHubApiRequirements
  customHeaders?: Record<string, string>
  baseUrl?: string
}

/**
 * Internal state for the GateHub API client
 */
export interface GateHubApiClientConfig {
  baseUrl?: string
  defaultHeaders?: Record<string, string>

  appId?: string
  cardAppId?: string
  managedUserUuid?: string
  sessionToken?: string
}

const GATEHUB_HOST = 'https://api.sandbox.gatehub.net'

/**
 * API Client used for operation with GateHub DIRECTLY.
 *
 * All requests are made using the config given when the client is created.
 */
class GateHubApiClient {
  private config: GateHubApiClientConfig

  constructor(config: GateHubApiClientConfig = {}) {
    this.config = {
      ...config,
      baseUrl: config.baseUrl ?? GATEHUB_HOST
    }
  }

  updateConfig(config: Partial<GateHubApiClientConfig>) {
    this.config = { ...this.config, ...config }
  }

  async makeRequest<T = any>(
    request: GateHubApiRequest
  ): Promise<GateHubApiResponse<T>> {
    try {
      let url = `${request.baseUrl || this.config.baseUrl}${request.endpoint}`
      if (request.queryParams) {
        const params = new URLSearchParams()
        Object.entries(request.queryParams).forEach(([key, value]) => {
          params.append(key, String(value))
        })
        url += `?${params.toString()}`
      }

      const headers = buildHeaders(request, this.config, url)

      const response = await fetch(url, {
        method: request.method,
        headers,
        ...(request.body && { body: JSON.stringify(request.body) })
      })

      let data: T | undefined
      const contentType = response.headers.get('content-type')
      if (contentType?.includes('application/json')) {
        data = await response.json()
      }

      if (!response.ok) {
        console.error('GateHub API request failed:', data)
        return {
          error: data ? JSON.stringify(data) : `HTTP ${response.status}`,
          status: response.status,
          headers: response.headers
        }
      }

      return {
        data,
        status: response.status,
        headers: response.headers
      }
    } catch (error) {
      return {
        error: error instanceof Error ? error.message : 'Unknown error',
        status: 0,
        headers: new Headers()
      }
    }
  }

  /* Card operations */
  async listCards(customerId: string) {
    if (!customerId) throw new Error('Customer ID required for listing cards')

    return this.makeRequest({
      endpoint: GATEHUB_API_ENDPOINTS.cards.list(customerId),
      method: 'GET',
      requires: { managedUser: true }
    })
  }

  async getCardDetails(cardId: string) {
    if (!cardId) throw new Error('Card ID required')

    return this.makeRequest({
      endpoint: GATEHUB_API_ENDPOINTS.cards.details(cardId),
      method: 'GET',
      requires: { managedUser: true }
    })
  }

  async getCardTransactions(
    cardId: string,
    pageSize?: number,
    pageNumber?: number
  ) {
    if (!cardId) throw new Error('Card ID required')

    return this.makeRequest({
      endpoint: GATEHUB_API_ENDPOINTS.cards.transactions(cardId),
      method: 'GET',
      queryParams: {
        ...(pageSize && { pageSize }),
        ...(pageNumber && { pageNumber })
      },
      requires: { managedUser: true }
    })
  }

  async lockCard(cardId: string, reasonCode?: string, note?: string) {
    if (!cardId) throw new Error('Card ID required')

    return this.makeRequest({
      endpoint: GATEHUB_API_ENDPOINTS.cards.lock(cardId),
      method: 'PUT',
      body: { note },
      queryParams: reasonCode ? { reasonCode } : undefined,
      requires: { managedUser: true }
    })
  }

  async unlockCard(cardId: string, note?: string) {
    if (!cardId) throw new Error('Card ID required')

    return this.makeRequest({
      endpoint: GATEHUB_API_ENDPOINTS.cards.unlock(cardId),
      method: 'PUT',
      body: { note },
      requires: { managedUser: true }
    })
  }

  async closeCard(cardId: string, reasonCode?: string) {
    if (!cardId) throw new Error('Card ID required')

    return this.makeRequest({
      endpoint: GATEHUB_API_ENDPOINTS.cards.close(cardId),
      method: 'DELETE',
      queryParams: reasonCode ? { reasonCode } : undefined,
      requires: { managedUser: true }
    })
  }

  async getCardLimits(cardId: string) {
    if (!cardId) throw new Error('Card ID required')

    return this.makeRequest({
      endpoint: GATEHUB_API_ENDPOINTS.cards.limits.get(cardId),
      method: 'GET',
      requires: { managedUser: true }
    })
  }

  async setCardLimits(cardId: string, limits: any) {
    if (!cardId) throw new Error('Card ID required')

    return this.makeRequest({
      endpoint: GATEHUB_API_ENDPOINTS.cards.limits.create(cardId),
      method: 'POST',
      body: limits,
      requires: { managedUser: true }
    })
  }

  /* Address operations */
  async listAddresses() {
    return this.makeRequest({
      endpoint: GATEHUB_API_ENDPOINTS.addresses.list,
      method: 'GET'
    })
  }

  async createAddress(addressData: any) {
    return this.makeRequest({
      endpoint: GATEHUB_API_ENDPOINTS.addresses.create,
      method: 'POST',
      body: addressData
    })
  }

  async deleteAddress(addressId: string) {
    return this.makeRequest({
      endpoint: GATEHUB_API_ENDPOINTS.addresses.delete(addressId),
      method: 'DELETE'
    })
  }

  /* Token operations */
  async getCardDataToken(cardId: string, publicKey: string) {
    if (!cardId || !publicKey)
      throw new Error('Card ID and public key required')

    return this.makeRequest({
      endpoint: GATEHUB_API_ENDPOINTS.tokens.cardData,
      method: 'POST',
      body: { cardId, publicKey },
      requires: {
        managedUser: true,
        cardAppId: true,
        appId: true,
        timestamp: true,
        signature: true
      }
    })
  }

  async getPinToken(cardId: string, publicKey: string) {
    if (!cardId || !publicKey)
      throw new Error('Card ID and public key required')

    return this.makeRequest({
      endpoint: GATEHUB_API_ENDPOINTS.tokens.pin,
      method: 'POST',
      body: { cardId, publicKey },
      requires: { managedUser: true }
    })
  }

  async getPinChangeToken(cardId: string) {
    if (!cardId) throw new Error('Card ID required')

    return this.makeRequest({
      endpoint: GATEHUB_API_ENDPOINTS.tokens.pinChange,
      method: 'POST',
      body: { cardId },
      requires: { managedUser: true }
    })
  }

  async getAppleProvisioningToken(cardId: string, publicKey: string) {
    if (!cardId || !publicKey)
      throw new Error('Card ID and public key required')

    return this.makeRequest({
      endpoint: GATEHUB_API_ENDPOINTS.tokens.appleProvisioning,
      method: 'POST',
      body: { cardId, publicKey },
      requires: { managedUser: true }
    })
  }

  async getGoogleProvisioningToken(cardId: string, publicKey: string) {
    if (!cardId || !publicKey)
      throw new Error('Card ID and public key required')

    return this.makeRequest({
      endpoint: GATEHUB_API_ENDPOINTS.tokens.googleProvisioning,
      method: 'POST',
      body: { cardId, publicKey },
      requires: { managedUser: true }
    })
  }
}

export type { GateHubApiClient }
export const createGateHubApiClient = (config: GateHubApiClientConfig = {}) => {
  return new GateHubApiClient(config)
}
