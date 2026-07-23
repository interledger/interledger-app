import { envValue } from '~/env.server'

const DUMMY_ORIGIN = 'https://dummy.origin'

// Comma-separated list of extra hosts that returnTo is allowed to redirect
// to, e.g. "wallet.example.com,pay.example.org". Supports a leading "."
// to allow any subdomain, e.g. ".example.com" matches "foo.example.com".
// Configure via the ALLOWED_RETURN_TO_HOSTS env var.

function isAllowedHost(hostname: string, allowedHosts: string[]): boolean {
  const normalizedHostname = hostname.toLowerCase()
  return allowedHosts.some((allowed) => {
    if (allowed.startsWith('.')) {
      // ".example.com" allows "example.com" and any subdomain of it
      return (
        normalizedHostname === allowed.slice(1) ||
        normalizedHostname.endsWith(allowed)
      )
    }
    return normalizedHostname === allowed
  })
}

export function safeReturnTo(
  value: string | undefined | null,
  fallback: string = '/'
): string {
  const allowedHostNames = [envValue('PAYMENT_POINTER_BASE')]

  if (!value) return fallback

  // In order to disallow protocol relative URLs //example.com
  if (value.startsWith('//')) return fallback

  try {
    // We provde the dummy origin as a base to check if the returnTo value includes
    // the origin. If the value includes the origin, it means that the returnTo value
    // might include a mailicious origin, therefore we return the fallback.
    // We should not accept absolute URLs, only relative URLs — unless the
    // host is on our explicit allowlist.
    const url = new URL(value, DUMMY_ORIGIN)

    if (url.origin === DUMMY_ORIGIN) {
      return url.pathname + url.search
    }

    // Cross-origin: only allow https, and only to an allowlisted host.
    if (
      url.protocol === 'https:' &&
      isAllowedHost(url.hostname, allowedHostNames)
    ) {
      return url.toString()
    }

    return fallback
  } catch (err) {
    return fallback
  }
}
