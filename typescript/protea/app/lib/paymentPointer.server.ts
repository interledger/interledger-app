import { envValue } from '~/env.server'

// Export to ensure this is always evaluated server side.
export const PAYMENT_POINTER_BASE =
  envValue("PAYMENT_POINTER_BASE") || 'ilp.link'
