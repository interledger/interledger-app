import { config } from '~/config.server'

// Export to ensure this is always evaluated server side.
export const PAYMENT_POINTER_BASE = config.payment_pointer_base
