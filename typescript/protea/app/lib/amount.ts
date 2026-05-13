/**
 * Parse a decimal amount string (e.g. "1.50") into a scaled bigint
 * (e.g. 150n for scale=2). Used to convert form input into the
 * scaled integer values expected by the backend's Amount proto.
 *
 * Behavior matches the legacy per-route copies: extra fractional
 * digits beyond `scale` are truncated, missing digits are right-padded
 * with zeros, and an empty string returns 0n.
 */
export function stringToBigInt(amount: string, scale = 2): bigint {
  if (amount === '') return BigInt(0)
  const dotIndex = amount.lastIndexOf('.')
  if (dotIndex > -1) {
    const parts = amount.split('.')
    return BigInt(parts[0] + parts[1].slice(0, scale).padEnd(scale, '0'))
  }
  return BigInt(parseFloat(amount) * 10 ** scale)
}
