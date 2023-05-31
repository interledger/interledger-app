import type { PusherArgs } from '~/lib/usePusher'
import { getWalletId } from '~/lib/wallet.server'

const PUSHER_APP_KEY = process.env.PUSHER_APP_KEY || '91988d6075551d29760a'
const PUSHER_CLUSTER = process.env.PUSHER_APP_CLUSTER || 'eu'

export async function getPusherArgs(request: Request): Promise<PusherArgs> {
  return {
    appKey: PUSHER_APP_KEY,
    cluster: PUSHER_CLUSTER,
    walletId: await getWalletId(request)
  }
}
