export function trimHeaders(header: Headers, keys: string[]): Headers {
  const newHeader = new Headers(header)

  const lowerCaseKeys = keys.map((key) => key.toLowerCase())

  newHeader.forEach((value, key) => {
    if (lowerCaseKeys.includes(key.toLowerCase())) {
      newHeader.delete(key)
    }
  })

  return newHeader
}
