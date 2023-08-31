export function sanitizeCMSLinks(to?: string) {
  return {
    internal: to?.startsWith('https://fynbos.app/'),
    toUrl: to?.replace('https://fynbos.app', '') ?? ''
  }
}
