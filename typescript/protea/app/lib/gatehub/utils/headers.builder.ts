import { GateHubApiClientConfig, GateHubApiRequest } from '../GateHubApiClient'
import type { GateHubRequestOptions } from '../types'

const GATEHUB_HEADERS = {
  APP_ID: 'x-gatehub-app-id',
  TIMESTAMP: 'x-gatehub-timestamp',
  SIGNATURE: 'x-gatehub-signature',
  MANAGED_USER_UUID: 'x-gatehub-managed-user-uuid',
  CARD_APP_ID: 'x-gatehub-card-app-id'
} as const

export interface GateHubAuthData {
  appId?: string
  cardAppId?: string
  managedUserUuid?: string
  sessionToken?: string
}

export function buildGateHubHeaders(
  authData: GateHubAuthData,
  options: GateHubRequestOptions = {},
  body?: any
): Record<string, string> {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json'
  }

  if (options.includeAppId) {
    if (!authData.appId) {
      throw new Error('App ID is required.')
    }
    headers['x-gatehub-app-id'] = authData.appId
  }

  if (options.includeCardAppId) {
    if (!authData.cardAppId) {
      throw new Error('Card app ID is required.')
    }
    headers['x-gatehub-card-app-id'] = authData.cardAppId
  }

  if (options.includeManagedUserUuid) {
    if (!authData.managedUserUuid) {
      throw new Error('Managed user UUID is required.')
    }
    headers['x-gatehub-managed-user-uuid'] = authData.managedUserUuid
  }

  if (options.includeSessionToken && authData.sessionToken) {
    if (!authData.sessionToken) {
      throw new Error('Session token is required.')
    }
    headers['Authorization'] = `Bearer ${authData.sessionToken}`
  }
  
  if (options.includeTimestamp) {
    headers['x-gatehub-timestamp'] = new Date().toISOString()
  }

  if (options.includeSignature) {
    headers['x-gatehub-signature'] = authData.signature
  }

  if (options.customHeaders) {
    Object.assign(headers, options.customHeaders)
  }

  return headers
}

function getSignature(timestamp: string, method: string, url: string, body: any) {
  const args = [timestamp, method, url];
  if (body) {
    args.push(body);
  }

  const toSign = args.join("|");
  return crypto.createHmac("sha256", SECRET).update(toSign).digest("hex");
}
