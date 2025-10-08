import { GATEHUB_API_ENDPOINTS } from '~/lib/gatehub/config/endpoints'
import type { GateHubApiResponse, HttpMethod } from '~/lib/gatehub/types'
import { buildHeaders } from '~/lib/gatehub/utils/headers.builder'

/**
 * Flags to determine which headers to be used for the direct request to the card processor
 */
export interface CardProcessorApiRequirements {
  appId?: boolean
  cardAppId?: boolean
  managedUser?: boolean
  sessionToken?: string
  timestamp?: boolean
  signature?: boolean
}

export interface CardProcessorApiRequest {
  endpoint: string
  method: HttpMethod
  body?: any
  queryParams?: Record<string, string | number | boolean>
  requires?: CardProcessorApiRequirements
  customHeaders?: Record<string, string>
  baseUrl?: string
}

/**
 * Internal state for the Card Processor API client
 */
export interface CardProcessorApiClientConfig {
  baseUrl?: string
  defaultHeaders?: Record<string, string>

  appId?: string
  cardAppId?: string
  managedUserUuid?: string
}

const GATEHUB_HOST = 'https://api.sandbox.gatehub.net'

/**
 * API Client used for operation with Card Processor DIRECTLY.
 *
 * All requests are made using the config given when the client is created.
 */
class CardProcessorApiClient {
  private config: CardProcessorApiClientConfig

  constructor(config: CardProcessorApiClientConfig = {}) {
    this.config = {
      ...config,
      baseUrl: config.baseUrl ?? GATEHUB_HOST
    }
  }

  updateConfig(config: Partial<CardProcessorApiClientConfig>) {
    this.config = { ...this.config, ...config }
  }

  async makeRequest<T = any>(
    request: CardProcessorApiRequest
  ): Promise<GateHubApiResponse<T>> {
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
      throw new Error(
        JSON.stringify({
          error: data ? JSON.stringify(data) : '',
          status: response.status,
          headers: response.headers
        })
      )
    }

    return {
      data,
      status: response.status,
      headers: response.headers
    }
  }

  async getCardSensitiveData({
    jwtToken,
    cardProcessorUrl,
    httpMethod
  }: {
    jwtToken: string
    cardProcessorUrl: string
    httpMethod: HttpMethod
  }) {
    if (!jwtToken) throw new Error('JWT token required')

    return this.makeRequest({
      baseUrl: cardProcessorUrl,
      method: httpMethod,
      requires: { sessionToken: jwtToken },
      endpoint: ''
    })
  }

  async lockCard(cardId: string) {
    return this.makeRequest({
      endpoint: GATEHUB_API_ENDPOINTS.cards.lockCard(cardId),
      method: 'PUT',
      queryParams: { reasonCode: 'ClientRequestedLock' },
      requires: {
        cardAppId: true,
        managedUser: true,
        timestamp: true,
        signature: true,
        appId: true
      }
    })
  }

  async unlockCard(cardId: string) {
    return this.makeRequest({
      endpoint: GATEHUB_API_ENDPOINTS.cards.unlockCard(cardId),
      method: 'PUT',
      requires: {
        cardAppId: true,
        managedUser: true,
        timestamp: true,
        signature: true,
        appId: true
      }
    })
  }

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
}

export type { CardProcessorApiClient }
export const createCardProcessorApiClient = (
  config: CardProcessorApiClientConfig = {}
) => {
  return new CardProcessorApiClient(config)
}
