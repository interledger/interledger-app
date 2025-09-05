export function sanitizeCMSLinks(to?: string) {
  return {
    internal: to?.startsWith('https://interledger.app/'),
    toUrl: to?.replace('https://interledger.app', '') ?? ''
  }
}
