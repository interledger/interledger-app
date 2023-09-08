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
      { writeKey: apiKey },
      { initialPageview: false }
    )
  }

  useEffect(() => {
    if (global.__segmentClient) {
      global.__segmentClient.page()
    }
  }, [location.pathname])
}
