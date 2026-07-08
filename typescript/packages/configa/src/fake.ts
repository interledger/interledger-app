import { ErrSecretNotFound } from './errors.js'
import type { SecretClient } from './types.js'

/**
 * In-memory SecretClient for tests. Mirrors go/configa's test mockClient —
 * tracks per-secret call counts so tests can assert on fetch deduplication.
 */
export class FakeSecretClient implements SecretClient {
  readonly callCount = new Map<string, number>()

  constructor(
    private readonly secrets: Record<string, Record<string, string>> = {}
  ) {}

  async getSecret(
    namespace: string,
    name: string
  ): Promise<Record<string, string>> {
    this.callCount.set(name, (this.callCount.get(name) ?? 0) + 1)
    const data = this.secrets[name]
    if (!data) {
      throw new ErrSecretNotFound(namespace, name)
    }
    return data
  }
}
