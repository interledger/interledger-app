declare module 'routes-gen' {
  export type RouteParams = {
    '/blog/connecting-the-internet-economy': {}
    '/recovery/useRecoveryFlow': {}
    '/signup/useSignupFlow': {}
    '/verify/useVerifyFlow': {}
    '/login/useLoginFlow': {}
    '/': {}
    '/activity/transaction/:id': { id: string }
    '/settings/password': {}
    '/transact/preview': {}
    '/transact/receive': {}
    '/activity/filter': {}
    '/activity': {}
    '/settings': {}
    '/transact': {}
    '/withdraw': {}
    '/connect': {}
    '/deposit': {}
    '/home': {}
    '/recovery': {}
    '/logout': {}
    '/signup': {}
    '/verify': {}
    '/login': {}
    '/blog': {}
  }

  export function route<
    T extends
      | ['/blog/connecting-the-internet-economy']
      | ['/recovery/useRecoveryFlow']
      | ['/signup/useSignupFlow']
      | ['/verify/useVerifyFlow']
      | ['/login/useLoginFlow']
      | ['/']
      | ['/activity/transaction/:id', RouteParams['/activity/transaction/:id']]
      | ['/settings/password']
      | ['/transact/preview']
      | ['/transact/receive']
      | ['/activity/filter']
      | ['/activity']
      | ['/settings']
      | ['/transact']
      | ['/withdraw']
      | ['/connect']
      | ['/deposit']
      | ['/home']
      | ['/recovery']
      | ['/logout']
      | ['/signup']
      | ['/verify']
      | ['/login']
      | ['/blog']
  >(...args: T): typeof args[0]
}
