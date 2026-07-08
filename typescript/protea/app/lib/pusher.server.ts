import { config } from '~/config.server'
import { getWalletId } from '~/data/wallet.server'
import type { PusherArgs } from '~/lib/usePusher'

const PUSHER_APP_KEY = config.pusher.app_key
const PUSHER_CLUSTER = config.pusher.app_cluster

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
