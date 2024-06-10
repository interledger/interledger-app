export function sanitizeCMSLinks(to?: string) {
  return {
    internal: to?.startsWith('https://wallet.fynbos.app/'),
    toUrl: to?.replace('https://wallet.fynbos.app', '') ?? ''
  }
}
