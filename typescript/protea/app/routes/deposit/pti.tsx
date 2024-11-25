import { useLoaderData, useSubmit } from '@remix-run/react'
import { useEffect } from 'react'
import { PtiWidget } from '~/generated/connect/backend/v1/backend_pb'
import { FiantSdkMessage } from '~/lib/fiant'
import { useScaffoldStore } from '~/lib/useScaffoldStore'
import { useScript } from '~/lib/useScript'

export default function PtiDepositPage() {
  const { ptiWidget: widget, csrfToken } = useLoaderData<{ provider: string, ptiWidget?: PtiWidget | undefined, csrfToken: string }>();
  const submit = useSubmit()
  const [setLoading] = useScaffoldStore((state) => [state.setLoading])
  const scriptStatus = useScript(
    widget?.sdkUrl || 'https://sdk.platform.fiant.io/0.0.23/index.js'
  )
  useEffect(() => {
    // This ensures that loading is false when this route is unmounted.
    return () => {
      setLoading(false)
    }
  }, [setLoading])

  useEffect(() => {
    if (scriptStatus == 'ready' && typeof (window as any).PTI !== 'undefined') {
      const styling = {
        mode: window.matchMedia('(prefers-color-scheme: dark)').matches
          ? 'dark'
          : 'light',
        primaryColor: '#3b82f6',
        backgroundColor: '#f8fafc',
        fontFamily: 'Poppins'
      }

        ; (window as any).PTI.init({
          clientId: widget?.clientId,
          generateTokenPath: widget?.generateTokenPath,
          ptiFormsUrl: widget?.formsUrl || 'https://forms.platform.fiant.io'
        })
        ; (window as any).PTI.form({
          type: 'FIAT_FUNDING',
          requestId: widget?.requestId,
          userId: widget?.userId,
          scenarioId: widget?.scenarioId,
          parentElement: document.getElementById('payment_form'),
          lang: 'en',
          styleConfig: styling
        })
      setLoading(false)
    }

    const handleMessage = (message: MessageEvent<FiantSdkMessage>) => {
      console.log('message:', message.data)
      if (message.data.name === 'UserTransactionCompleted') {
        setLoading(true)
        // let formData = new FormData()
        // formData.append('tokenId', message.data.createdId)
        // formData.append('csrfToken', csrfToken)
        // submit(null, {
        //   action: `/connect/card`,
        //   method: 'post'
        // })
      }
    }
    window.addEventListener('message', handleMessage)

    return () => {
      window.removeEventListener('message', handleMessage)
    }
  }, [scriptStatus, widget, setLoading, submit])

  return (
    <>
      <div id='payment_form' className='h-[750px]' />
    </>
  )

}
