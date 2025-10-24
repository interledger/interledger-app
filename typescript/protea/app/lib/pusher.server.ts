import { getWalletId } from '~/data/wallet.server'
import type { PusherArgs } from '~/lib/usePusher'

const PUSHER_APP_KEY = process.env.PUSHER_APP_KEY || '91988d6075551d29760a'
const PUSHER_CLUSTER = process.env.PUSHER_APP_CLUSTER || 'eu'

export async function getPusherArgs(request: Request): Promise<PusherArgs> {
  const args: PusherArgs = {
    appKey: PUSHER_APP_KEY,
    cluster: PUSHER_CLUSTER
  }

  try {
    const walletId = await getWalletId(request)
    args.walletId = walletId
  } catch (_: unknown) {
    // do nothing - the user is not logged in
  }

  return args
}
