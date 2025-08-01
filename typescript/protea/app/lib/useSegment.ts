import { useLocation } from '@remix-run/react'
import { AnalyticsBrowser } from '@segment/analytics-next'
import { useEffect } from 'react'

declare global {
  var __segmentClient: AnalyticsBrowser | undefined
}

// TODO: Can it be removed?
export function useSegment(apiKey: string) {
  const location = useLocation()
  if (!global.__segmentClient && apiKey && apiKey != '') {
    global.__segmentClient = AnalyticsBrowser.load(
      {
        writeKey: apiKey,
        cdnURL: 'https://s.fynbos.app'
      },
      {
        initialPageview: false,
        integrations: {
          'Segment.io': {
            apiHost: 's.fynbos.app/v1',
            protocol: 'https' // optional
          }
        }
      }
    )
  }

  useEffect(() => {
    if (global.__segmentClient) {
      global.__segmentClient.page()
    }
  }, [location.pathname])
}
