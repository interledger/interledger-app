import { createHmac } from 'crypto'
import {
  CardProcessorApiClientConfig,
  CardProcessorApiRequest
} from '../CardProcessorApiClient'
import { GATEHUB_SECRET } from '../do-not-commit'

export interface GateHubAuthData {
  appId?: string
  cardAppId?: string
  managedUserUuid?: string
  sessionToken?: string
}

export function buildHeaders(
  request: CardProcessorApiRequest,
  config: CardProcessorApiClientConfig,
  url: string
): Record<string, string> {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json'
  }

  console.log('request', request)

  const { requires, customHeaders } = request

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
    console.log('sessionToken', requires.sessionToken)
    headers['Authorization'] = `Bearer ${requires.sessionToken}`
  }

  if (requires.timestamp) {
    const timestamp = Date.now()
    headers['x-gatehub-timestamp'] = timestamp.toString()
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
    args.push(JSON.stringify(body))
  }

  console.log('timestamp', timestamp)
  console.log('method', method)
  console.log('url', url)
  console.log('body', body)
  console.log('args', args)

  const toSign = args.join('|')
  console.log('toSign', toSign)
  console.log('secret', GATEHUB_SECRET)
  return createHmac('sha256', GATEHUB_SECRET).update(toSign).digest('hex')
}
