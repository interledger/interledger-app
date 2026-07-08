/**
 * SecretClient fetches Kubernetes secrets by name, returning all data keys.
 */
export interface SecretClient {
  getSecret(namespace: string, name: string): Promise<Record<string, string>>
}
