declare module 'routes-gen' {
  export type RouteParams = {
    '/dynamic-form/:id.csv': { id: string }
    '/dynamic-form/:id': { id: string }
    '/dynamic-form/:id/submissions': { id: string }
    '/dynamic-form/:id/submissions/:submissionId': {
      id: string
      submissionId: string
    }
    '/pendingpayouts': Record<string, never>
    '/dynamic-forms': Record<string, never>
    '/review/:id': { id: string }
    '/review/:id/details': { id: string }
    '/wallet/:id': { id: string }
    '/wallet/:id/linked-accounts': { id: string }
    '/wallet/:id/transactions': { id: string }
    '/wallet/:id/transactions/:transactionId': {
      id: string
      transactionId: string
    }
    '/wallet/:id/balances': { id: string }
    '/wallet/:id/gatehub': { id: string }
    '/wallet/:id/profile': { id: string }
    '/wallet/:id/audit': { id: string }
    '/waitlist': Record<string, never>
    '/healthz': Record<string, never>
    '/live': Record<string, never>
    '/reviews': Record<string, never>
    '/wallets': Record<string, never>
    '/': Record<string, never>
  }

  export function route<
    T extends
      | ['/dynamic-form/:id.csv', RouteParams['/dynamic-form/:id.csv']]
      | ['/dynamic-form/:id', RouteParams['/dynamic-form/:id']]
      | [
          '/dynamic-form/:id/submissions',
          RouteParams['/dynamic-form/:id/submissions']
        ]
      | [
          '/dynamic-form/:id/submissions/:submissionId',
          RouteParams['/dynamic-form/:id/submissions/:submissionId']
        ]
      | ['/pendingpayouts']
      | ['/dynamic-forms']
      | ['/review/:id', RouteParams['/review/:id']]
      | ['/review/:id/details', RouteParams['/review/:id/details']]
      | ['/wallet/:id', RouteParams['/wallet/:id']]
      | [
          '/wallet/:id/linked-accounts',
          RouteParams['/wallet/:id/linked-accounts']
        ]
      | ['/wallet/:id/transactions', RouteParams['/wallet/:id/transactions']]
      | [
          '/wallet/:id/transactions/:transactionId',
          RouteParams['/wallet/:id/transactions/:transactionId']
        ]
      | ['/wallet/:id/balances', RouteParams['/wallet/:id/balances']]
      | ['/wallet/:id/gatehub', RouteParams['/wallet/:id/gatehub']]
      | ['/wallet/:id/profile', RouteParams['/wallet/:id/profile']]
      | ['/wallet/:id/audit', RouteParams['/wallet/:id/audit']]
      | ['/waitlist']
      | ['/healthz']
      | ['/live']
      | ['/reviews']
      | ['/wallets']
      | ['/']
  >(...args: T): (typeof args)[0]
}
