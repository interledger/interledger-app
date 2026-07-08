import { readFileSync } from 'node:fs'
import { parse as parseYAML } from 'yaml'
import type { ZodType } from 'zod'
import { deepMerge } from './merge.js'
import {
  collectSecretNames,
  hasTemplateMarker,
  substituteSecrets
} from './template.js'
import { ErrNoSecretClient } from './errors.js'
import type { SecretClient } from './types.js'

export interface ParseOptions {
  /** Required when the config contains {{ secret "name" "key" }} templates. */
  secretClient?: SecretClient
  /** Kubernetes namespace used for secret lookups. Defaults to the secret client's own namespace. */
  namespace?: string
}

/**
 * Holds a parsed YAML configuration, merged across overlay files and ready
 * for secret template resolution. Port of go/configa's Config[T].
 */
export class ParsedConfig {
  private readonly merged: Record<string, unknown>
  private readonly hasTemplates: boolean
  private readonly opts: ParseOptions

  /** @internal use parseConfig() */
  constructor(merged: Record<string, unknown>, opts: ParseOptions) {
    this.merged = merged
    this.hasTemplates = hasTemplateMarker(merged)
    this.opts = opts
  }

  /**
   * Substitutes all {{ secret "name" "key" }} expressions (fetching each
   * referenced secret at most once), then validates the result against
   * schema. Skips all Kubernetes API calls entirely when no templates were
   * found during parseConfig, so it is safe to call outside a Kubernetes
   * environment.
   */
  async resolve<T>(schema: ZodType<T>): Promise<T> {
    let resolved: unknown = this.merged

    if (this.hasTemplates) {
      if (!this.opts.secretClient) {
        throw new ErrNoSecretClient()
      }
      resolved = await resolveSecretTemplates(
        this.merged,
        this.opts.secretClient,
        this.opts.namespace
      )
    }

    const result = schema.safeParse(resolved)
    if (!result.success) {
      throw new Error(`configa: ${result.error.message}`, {
        cause: result.error
      })
    }
    return result.data
  }
}

async function resolveSecretTemplates(
  merged: Record<string, unknown>,
  client: SecretClient,
  namespace: string | undefined
): Promise<unknown> {
  const names = collectSecretNames(merged)
  const cache = new Map<string, Record<string, string>>()

  await Promise.all(
    Array.from(names).map(async (name) => {
      cache.set(name, await client.getSecret(namespace ?? '', name))
    })
  )

  return substituteSecrets(merged, cache)
}

function readYAMLFile(filename: string): Record<string, unknown> {
  let raw: string
  try {
    raw = readFileSync(filename, 'utf8')
  } catch (err) {
    throw new Error(
      `configa: read config file "${filename}": ${(err as Error).message}`,
      { cause: err }
    )
  }

  let parsed: unknown
  try {
    parsed = parseYAML(raw) ?? {}
  } catch (err) {
    throw new Error(
      `configa: invalid yaml in "${filename}": ${(err as Error).message}`,
      { cause: err }
    )
  }

  if (typeof parsed !== 'object' || Array.isArray(parsed)) {
    throw new Error(
      `configa: invalid yaml in "${filename}": expected a YAML mapping at the top level`
    )
  }

  return parsed as Record<string, unknown>
}

/**
 * Reads one or more YAML configuration files and merges them in order.
 * Later files act as overlays: their values take precedence over earlier
 * ones. For nested maps the merge is deep; for scalars and arrays the
 * overlay wins entirely. Port of go/configa's Parse.
 */
export function parseConfig(
  filenames: string[],
  opts: ParseOptions = {}
): ParsedConfig {
  if (filenames.length === 0) {
    throw new Error('configa: at least one filename is required')
  }

  let merged: Record<string, unknown> | undefined
  for (const filename of filenames) {
    const parsed = readYAMLFile(filename)
    merged = merged === undefined ? parsed : deepMerge(merged, parsed)
  }

  return new ParsedConfig(merged as Record<string, unknown>, opts)
}
