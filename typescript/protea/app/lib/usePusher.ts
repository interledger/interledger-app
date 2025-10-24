import { useRevalidator } from '@remix-run/react'
import type { Channel, PresenceChannel } from 'pusher-js'
import Pusher from 'pusher-js'
import { useEffect, useState } from 'react'
import { useScaffoldStore } from './useScaffoldStore'

let pusherClient: Pusher

declare global {
  var __pusherClient: Pusher | undefined
}

type Events = 'linkedAccount' | 'transaction' | 'kyc' | 'identity' | 'cardReady'

export type PusherArgs = {
  appKey: string
  cluster: string
  walletId?: string
}

export function usePusher(args: PusherArgs, events: Events[]) {
  const { revalidate, state } = useRevalidator()
  const [pushSnackbar] = useScaffoldStore((state) => [state.pushSnackbar])

  if (!global.__pusherClient) {
    global.__pusherClient = new Pusher(args.appKey, {
      cluster: args.cluster
    })
  }
  pusherClient = global.__pusherClient

  let channelName: string | undefined
  if (args.walletId) {
    channelName = `wallet-${args.walletId}`
  }
  const channel = useChannel(channelName)

  usePusherEvent(channel, 'transaction', () => {
    if (state == 'idle' && events.find((e) => e == 'transaction')) revalidate()
  })
  usePusherEvent(channel, 'kyc', () => {
    if (state == 'idle' && events.find((e) => e == 'kyc')) revalidate()
  })
  usePusherEvent(channel, 'identity', () => {
    if (state == 'idle' && events.find((e) => e == 'identity')) revalidate()
  })
  usePusherEvent(channel, 'linkedAccount', () => {
    if (state == 'idle' && events.find((e) => e == 'linkedAccount'))
      revalidate()
  })
  usePusherEvent(channel, 'cardReady', () => {
    if (state == 'idle' && events.find((e) => e == 'cardReady'))
      pushSnackbar({
        message: 'Your card is now ready!',
        id: crypto.randomUUID(),
        action: 'View cards'
      })
  })
  // TODO: Maybe return connection state?
}

function usePusherEvent<D>(
  channel: Channel | PresenceChannel | undefined,
  eventName: string,
  callback: (data?: D, metadata?: { user_id: string }) => void
) {
  useEffect(() => {
    if (channel === undefined) {
      return
    } else {
      channel.bind(eventName, callback)
    }
    return () => {
      channel.unbind(eventName, callback)
    }
  }, [channel, eventName, callback])
}

function useChannel(channelName?: string) {
  const [channel, setChannel] = useState<Channel>()
  useEffect(() => {
    if (!channelName) return
    const _channel = pusherClient.subscribe(channelName)
    setChannel(_channel)
    return () => pusherClient.unsubscribe(channelName)
  }, [channelName])

  return channel
}
