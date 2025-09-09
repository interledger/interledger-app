import type { GateHubRequestOptions } from '../types'

export interface GateHubAuthData {
  cardAppId?: string
  managedUserUuid?: string
  sessionToken?: string
}

export function buildGateHubHeaders(
  authData: GateHubAuthData,
  options: GateHubRequestOptions = {}
): Record<string, string> {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json'
  }

  if (options.includeCardAppId !== false && authData.cardAppId) {
    headers['x-gatehub-card-app-id'] = authData.cardAppId
  }

  if (options.includeManagedUserUuid && authData.managedUserUuid) {
    headers['x-gatehub-managed-user-uuid'] = authData.managedUserUuid
  }

  if (authData.sessionToken) {
    headers['Authorization'] = `Bearer ${authData.sessionToken}`
  }

  if (options.customHeaders) {
    Object.assign(headers, options.customHeaders)
  }

  return headers
}
