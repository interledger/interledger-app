import type { Channel, PresenceChannel } from 'pusher-js'
import Pusher from 'pusher-js'
import { useRevalidator } from '@remix-run/react'
import { useEffect, useState } from 'react'

// Hard-coding for now as this is required to evaluate client side.
const PUSHER_APP_KEY = '91988d6075551d29760a'

let pusherClient: Pusher

declare global {
  var __pusherClient: Pusher | undefined
}

// this is needed because in development we don't want to restart
// the server with every change, but we want to make sure we don't
// create a new connection to the Client with every change either.
if (process.env.NODE_ENV === 'production') {
  pusherClient = new Pusher(PUSHER_APP_KEY, {
    cluster: 'eu'
  })
} else {
  if (!global.__pusherClient) {
    global.__pusherClient = new Pusher(PUSHER_APP_KEY, {
      cluster: 'eu'
    })
  }
  pusherClient = global.__pusherClient
}

type Events = 'linkedAccount' | 'transaction' | 'kyc'

export function usePusher(walletId: string, events: Events[]) {
  const { revalidate, state } = useRevalidator()

  const channel = useChannel(`wallet-${walletId}`)

  useEvent(channel, 'transaction', () => {
    if (state == 'idle' && events.find((e) => e == 'transaction')) revalidate()
  })
  useEvent(channel, 'kyc', () => {
    if (state == 'idle' && events.find((e) => e == 'kyc')) revalidate()
  })
  // TODO: Maybe return connection state?
}

function useEvent<D>(
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

function useChannel(channelName: string) {
  const [channel, setChannel] = useState<Channel>()
  useEffect(() => {
    const _channel = pusherClient.subscribe(channelName)
    setChannel(_channel)
    return () => pusherClient.unsubscribe(channelName)
  }, [channelName])

  return channel
}
