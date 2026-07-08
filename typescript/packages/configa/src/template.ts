// Matches a single {{ ... }} template block (no nested braces supported —
// matches the Go text/template usage in go/configa, which only ever embeds
// a single "secret" function call per expression).
const TEMPLATE_BLOCK_RE = /\{\{[^{}]*\}\}/g

// Matches exactly `{{ secret "name" "key" }}`, whitespace-tolerant.
const SECRET_CALL_RE =
  /^\{\{\s*secret\s+"((?:\\.|[^"\\])*)"\s+"((?:\\.|[^"\\])*)"\s*\}\}$/

function unescape(s: string): string {
  return s.replace(/\\(.)/g, '$1')
}

interface SecretCall {
  name: string
  key: string
}

function parseTemplateBlock(block: string): SecretCall {
  const match = SECRET_CALL_RE.exec(block)
  if (!match) {
    throw new Error(`configa: unsupported template expression: ${block}`)
  }
  return { name: unescape(match[1]), key: unescape(match[2]) }
}

/** True if any string leaf in the tree contains a `{{` marker. */
export function hasTemplateMarker(value: unknown): boolean {
  if (typeof value === 'string') {
    return value.includes('{{')
  }
  if (Array.isArray(value)) {
    return value.some(hasTemplateMarker)
  }
  if (value !== null && typeof value === 'object') {
    return Object.values(value as Record<string, unknown>).some(
      hasTemplateMarker
    )
  }
  return false
}

/**
 * Walks the whole tree and returns the set of distinct secret names
 * referenced by {{ secret "name" "key" }} expressions. Throws immediately
 * on any {{ }} expression that isn't a well-formed secret call — this
 * mirrors go/configa's template.Parse failing fast for unknown functions,
 * before any Kubernetes API calls happen.
 */
export function collectSecretNames(value: unknown): Set<string> {
  const names = new Set<string>()
  walkStrings(value, (s) => {
    for (const block of s.match(TEMPLATE_BLOCK_RE) ?? []) {
      names.add(parseTemplateBlock(block).name)
    }
  })
  return names
}

/**
 * Returns a new tree with every {{ secret "name" "key" }} expression
 * substituted using the pre-fetched secret cache (name -> data keys).
 * Throws if a referenced key is absent from its secret's data.
 */
export function substituteSecrets(
  value: unknown,
  cache: Map<string, Record<string, string>>
): unknown {
  if (typeof value === 'string') {
    if (!value.includes('{{')) {
      return value
    }
    return value.replace(TEMPLATE_BLOCK_RE, (block) => {
      const { name, key } = parseTemplateBlock(block)
      const data = cache.get(name)
      const resolved = data?.[key]
      if (resolved === undefined) {
        throw new Error(
          `configa: key "${key}" not found in secret "${name}"`
        )
      }
      return resolved
    })
  }
  if (Array.isArray(value)) {
    return value.map((item) => substituteSecrets(item, cache))
  }
  if (value !== null && typeof value === 'object') {
    const result: Record<string, unknown> = {}
    for (const [k, v] of Object.entries(value as Record<string, unknown>)) {
      result[k] = substituteSecrets(v, cache)
    }
    return result
  }
  return value
}

function walkStrings(value: unknown, visit: (s: string) => void): void {
  if (typeof value === 'string') {
    visit(value)
    return
  }
  if (Array.isArray(value)) {
    value.forEach((item) => walkStrings(item, visit))
    return
  }
  if (value !== null && typeof value === 'object') {
    Object.values(value as Record<string, unknown>).forEach((v) =>
      walkStrings(v, visit)
    )
  }
}
