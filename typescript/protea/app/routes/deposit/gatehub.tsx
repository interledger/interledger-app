import { useEffect } from 'react'
import { useLoaderData, useNavigate } from 'react-router'

import type { IframeMessage } from '~/lib/types'
import { useScaffoldStore } from '~/lib/useScaffoldStore'
import { gatehubDepositLoader } from './loader.server'

export function GatehubDepositPage() {
  const { gatehubWidgetUrl } = useLoaderData<typeof gatehubDepositLoader>()
  const [pushSnackbar] = useScaffoldStore((state) => [state.pushSnackbar])
  const navigate = useNavigate()
  useEffect(() => {
    const url = new URL(gatehubWidgetUrl)
    const handler = (event: MessageEvent<IframeMessage>) => {
      if (
        event.origin === url.origin &&
        event.data.type === 'StripeDepositCompleted'
      ) {
        pushSnackbar({
          id: 'deposit-success',
          message:
            'Deposit submitted successfully, it may take a few minutes to reflect in your account',
          icon: 'close',
          canShow: true
        })
        navigate('/')
      }
    }
    if (window) {
      window.addEventListener('message', handler)
    }

    return () => {
      if (window) {
        window.removeEventListener('message', handler)
      }
    }
  }, [gatehubWidgetUrl, navigate, pushSnackbar])

  return (
    <iframe
      title='Deposit'
      src={gatehubWidgetUrl}
      sandbox='allow-top-navigation allow-forms allow-same-origin allow-popups allow-scripts'
      scrolling='no'
      frameBorder='0'
      className='h-[750px]'
    />
  )
}
