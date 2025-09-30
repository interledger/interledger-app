import { createHmac } from 'crypto'
import { GateHubApiClientConfig, GateHubApiRequest } from '../GateHubApiClient'

export interface GateHubAuthData {
  appId?: string
  cardAppId?: string
  managedUserUuid?: string
  sessionToken?: string
}

export function buildHeaders(
  request: GateHubApiRequest,
  config: GateHubApiClientConfig,
  url: string
): Record<string, string> {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json'
  }

  const { requires, customHeaders } = request
  console.log('requires', requires)

  if (customHeaders) {
    Object.assign(headers, customHeaders)
  }

  if (!requires) {
    return headers
  }

  if (requires.appId) {
    if (!config.appId) {
      throw new Error('App ID is required.')
    }
    headers['x-gatehub-app-id'] = config.appId
  }

  if (requires.cardAppId) {
    if (!config.cardAppId) {
      throw new Error('Card app ID is required.')
    }
    headers['x-gatehub-card-app-id'] = config.cardAppId
  }

  if (requires.managedUser) {
    if (!config.managedUserUuid) {
      throw new Error('Managed user UUID is required.')
    }
    headers['x-gatehub-managed-user-uuid'] = config.managedUserUuid
  }

  if (requires.sessionToken) {
    if (!config.sessionToken) {
      throw new Error('Session token is required.')
    }
    headers['Authorization'] = `Bearer ${config.sessionToken}`
  }

  if (requires.timestamp) {
    headers['x-gatehub-timestamp'] = new Date().toISOString()
  }

  if (requires.signature) {
    headers['x-gatehub-signature'] = getSignature(
      headers['x-gatehub-timestamp'],
      request.method,
      url,
      request.body
    )
  }

  return headers
}

// TODO THIS SHOULD NOT BE ON THE CLIENT
export function getSignature(
  timestamp: string,
  method: string,
  url: string,
  body?: string
) {
  const args = [timestamp, method, url]
  if (body) {
    args.push(body)
  }

  const toSign = args.join('|')
  return createHmac('sha256', 'SECRET_KEY')
    .update(toSign)
    .digest('hex')
}
