import { readFile } from 'node:fs/promises'
import { Agent, fetch as undiciFetch } from 'undici'
import {
  ErrSecretFetchFailed,
  ErrSecretForbidden,
  ErrSecretNotFound
} from './errors.js'
import type { SecretClient } from './types.js'

const DEFAULT_TOKEN_PATH =
  '/var/run/secrets/kubernetes.io/serviceaccount/token'
const DEFAULT_CA_PATH = '/var/run/secrets/kubernetes.io/serviceaccount/ca.crt'
const DEFAULT_NAMESPACE_PATH =
  '/var/run/secrets/kubernetes.io/serviceaccount/namespace'

const REQUEST_TIMEOUT_MS = 10_000

export interface InClusterSecretClientPaths {
  tokenPath?: string
  caPath?: string
  namespacePath?: string
}

interface Credentials {
  token: string
  defaultNamespace: string
  apiBase: string
  dispatcher: Agent
}

interface K8sSecretResponse {
  data?: Record<string, string>
}

/**
 * Fetches Kubernetes secrets using the pod's in-cluster service account
 * credentials. Credentials are loaded lazily on the first getSecret() call,
 * so the client is safe to construct outside Kubernetes. Port of
 * go/configa/k8s.go's InClusterSecretClient.
 */
export class InClusterSecretClient implements SecretClient {
  private readonly tokenPath: string
  private readonly caPath: string
  private readonly namespacePath: string
  private credentialsPromise: Promise<Credentials> | undefined

  constructor(paths: InClusterSecretClientPaths = {}) {
    this.tokenPath = paths.tokenPath ?? DEFAULT_TOKEN_PATH
    this.caPath = paths.caPath ?? DEFAULT_CA_PATH
    this.namespacePath = paths.namespacePath ?? DEFAULT_NAMESPACE_PATH
  }

  private loadCredentials(): Promise<Credentials> {
    // Memoized promise plays the role of Go's sync.Once: credential files
    // are read at most once, even under concurrent getSecret() calls.
    if (!this.credentialsPromise) {
      this.credentialsPromise = this.doLoadCredentials()
    }
    return this.credentialsPromise
  }

  private async doLoadCredentials(): Promise<Credentials> {
    let token: string
    try {
      token = (await readFile(this.tokenPath, 'utf8')).trim()
    } catch (err) {
      throw new Error(
        `configa: read service account token: ${(err as Error).message}`,
        { cause: err }
      )
    }

    let ca: string
    try {
      ca = await readFile(this.caPath, 'utf8')
    } catch (err) {
      throw new Error(
        `configa: read service account CA cert: ${(err as Error).message}`,
        { cause: err }
      )
    }

    let defaultNamespace: string
    try {
      defaultNamespace = (await readFile(this.namespacePath, 'utf8')).trim()
    } catch (err) {
      throw new Error(
        `configa: read service account namespace: ${(err as Error).message}`,
        { cause: err }
      )
    }

    const host = process.env.KUBERNETES_SERVICE_HOST
    const port = process.env.KUBERNETES_SERVICE_PORT
    if (!host || !port) {
      throw new Error(
        'configa: KUBERNETES_SERVICE_HOST or KUBERNETES_SERVICE_PORT not set'
      )
    }

    return {
      token,
      defaultNamespace,
      apiBase: `https://${host}:${port}`,
      dispatcher: new Agent({ connect: { ca } })
    }
  }

  async getSecret(
    namespace: string,
    name: string
  ): Promise<Record<string, string>> {
    const creds = await this.loadCredentials()
    const ns = namespace || creds.defaultNamespace

    const url = `${creds.apiBase}/api/v1/namespaces/${ns}/secrets/${name}`
    let response: Awaited<ReturnType<typeof undiciFetch>>
    try {
      response = await undiciFetch(url, {
        headers: { authorization: `Bearer ${creds.token}` },
        dispatcher: creds.dispatcher,
        signal: AbortSignal.timeout(REQUEST_TIMEOUT_MS)
      })
    } catch (err) {
      throw new Error(
        `configa: k8s request failed: ${(err as Error).message}`,
        { cause: err }
      )
    }

    if (response.status === 404) {
      throw new ErrSecretNotFound(ns, name)
    }
    if (response.status === 403 || response.status === 401) {
      throw new ErrSecretForbidden(ns, name)
    }
    if (response.status !== 200) {
      throw new ErrSecretFetchFailed(ns, name, `HTTP ${response.status}`)
    }

    const body = (await response.json()) as K8sSecretResponse
    const result: Record<string, string> = {}
    for (const [key, value] of Object.entries(body.data ?? {})) {
      result[key] = Buffer.from(value, 'base64').toString('utf8')
    }
    return result
  }
}

/** Returns a SecretClient backed by the Kubernetes in-cluster service account. */
export function newInClusterSecretClient(): SecretClient {
  return new InClusterSecretClient()
}
