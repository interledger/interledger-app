const DUMMY_ORIGIN = 'https://dummy.origin'

export function safeReturnTo(
  value: string | undefined | null,
  fallback: string = '/'
): string {
  if (!value) return fallback

  // In order to disallow protocol relative URLs //example.com
  if (value.startsWith('//')) return fallback

  try {
    // We provde the dummy origin as a base to check if the returnTo value includes
    // the origin. If the value includes the origin, it means that the returnTo value
    // might include a mailicious origin, therefore we return the fallback.
    // We should not accept absolute URLs, only relative URLs.
    const url = new URL(value, DUMMY_ORIGIN)
    if (url.origin !== DUMMY_ORIGIN) return fallback

    return url.pathname + url.search
  } catch (err) {
    return fallback
  }
}
