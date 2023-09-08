import { useLocation } from '@remix-run/react'
import { AnalyticsBrowser } from '@segment/analytics-next'
import { useEffect } from 'react'

let segmentClient: AnalyticsBrowser

declare global {
  var __segmentClient: AnalyticsBrowser | undefined
}

export function useSegment(apiKey: string) {
  const location = useLocation()
  if (!global.__segmentClient && apiKey && apiKey != '') {
    global.__segmentClient = AnalyticsBrowser.load(
      {
        writeKey: apiKey,
        cdnURL: 'https://segment-proxy.matdabomb.workers.dev'
      },
      {
        initialPageview: false,
        integrations: {
          'Segment.io': {
            apiHost: 'https://segment-proxy.matdabomb.workers.dev',
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
