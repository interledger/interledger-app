function isPlainObject(v: unknown): v is Record<string, unknown> {
  return typeof v === 'object' && v !== null && !Array.isArray(v)
}

/**
 * deepMerge returns a new object that is overlay applied on top of base.
 * When both values are plain objects, they are merged recursively.
 * In all other cases (scalars, arrays) the overlay value replaces the base value.
 * Port of go/configa/merge.go's deepMerge.
 */
export function deepMerge(
  base: Record<string, unknown>,
  overlay: Record<string, unknown>
): Record<string, unknown> {
  const result: Record<string, unknown> = { ...base }
  for (const [key, overlayValue] of Object.entries(overlay)) {
    const baseValue = result[key]
    if (isPlainObject(baseValue) && isPlainObject(overlayValue)) {
      result[key] = deepMerge(baseValue, overlayValue)
      continue
    }
    result[key] = overlayValue
  }
  return result
}
