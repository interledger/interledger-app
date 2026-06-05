/**
 * Extracts only the requested header keys from an existing Headers object.
 *
 * Problem:
 * Originally, this function cloned the Headers object and used `newHeader.delete(key)`
 * inside `newHeader.forEach()`. With Node's native `undici` Headers implementation
 * (used heavily after migrating to Vite + newer Node versions), mutating the
 * Headers object during iteration causes keys to shift. As a result, the loop
 * skipped multiple headers (like `Content-Length`, `Content-Encoding`), which
 * inadvertently leaked them to the client and caused React Router / fetch to fail
 * with a mismatched content length "Failed to fetch" network error.
 *
 * Solution:
 * Iterate over the read-only source `Headers` object and append only the allowed
 * keys into a brand new `Headers` object.
 */
export function trimHeaders(header: Headers, keys: string[]): Headers {
  const newHeader = new Headers()
  const lowerCaseKeys = keys.map((key) => key.toLowerCase())

  header.forEach((value, key) => {
    if (lowerCaseKeys.includes(key.toLowerCase())) {
      newHeader.append(key, value)
    }
  })

  return newHeader
}
