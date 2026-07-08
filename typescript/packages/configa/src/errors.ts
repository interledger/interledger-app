/**
 * ErrNoSecretClient is thrown by resolve() when the config contains
 * {{ secret }} template expressions but no SecretClient was provided.
 */
export class ErrNoSecretClient extends Error {
  constructor() {
    super(
      'configa: config contains {{ secret }} templates but no SecretClient was provided'
    )
    this.name = 'ErrNoSecretClient'
  }
}

export class ErrSecretNotFound extends Error {
  constructor(namespace: string, name: string) {
    super(`configa: kubernetes secret not found: ${namespace}/${name}`)
    this.name = 'ErrSecretNotFound'
  }
}

export class ErrSecretForbidden extends Error {
  constructor(namespace: string, name: string) {
    super(
      `configa: forbidden: insufficient permissions to read kubernetes secret: ${namespace}/${name}`
    )
    this.name = 'ErrSecretForbidden'
  }
}

export class ErrSecretFetchFailed extends Error {
  constructor(namespace: string, name: string, detail: string) {
    super(
      `configa: failed to fetch kubernetes secret: ${namespace}/${name} (${detail})`
    )
    this.name = 'ErrSecretFetchFailed'
  }
}
