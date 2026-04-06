import { useEffect } from 'react'
import { useScript } from './useScript'
import { usePtiConfig } from './pti-context'

export function usePTISdk(sessionId: string, clientId: string) {
  const ptiConfig = usePtiConfig()
  const scriptStatus = useScript(ptiConfig?.sdkUrl ?? '')

  useEffect(() => {
    if (
      scriptStatus == 'ready' &&
      window.PTI !== undefined &&
      clientId !== '' &&
      sessionId !== ''
    ) {
      window.PTI.init({
        clientId: clientId,
        sessionId: sessionId
      })
    }
  }, [scriptStatus, clientId, sessionId])
}
