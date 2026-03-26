import { useEffect } from 'react'
import { useScript } from './useScript'

export function usePTISdk(sessionId: string, clientId: string) {
  const scriptStatus = useScript('https://sdk.pearsurge.io/0.0.18/index.js')

  useEffect(() => {
    if (scriptStatus == 'ready' && window.PTI !== undefined) {
      window.PTI.init({
        clientId: clientId,
        sessionId: sessionId
      })
    }
  }, [scriptStatus, clientId, sessionId])
}
