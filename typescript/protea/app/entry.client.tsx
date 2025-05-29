import { RemixBrowser, useLocation, useMatches } from '@remix-run/react'
import * as Sentry from '@sentry/remix'
import { StrictMode, startTransition, useEffect } from 'react'
import { hydrateRoot } from 'react-dom/client'

if (
  typeof (window as any).ENV !== 'undefined' &&
  (window as any).ENV.sentryDsn
) {
  Sentry.init({
    tunnel: '/api/fern',
    dsn: (window as any).ENV.sentryDsn,
    release: (window as any).ENV.sentryRelease,
    integrations: [
      new Sentry.BrowserTracing({
        routingInstrumentation: Sentry.remixRouterInstrumentation(
          useEffect,
          useLocation,
          useMatches
        )
      }),
      new Sentry.Replay()
    ],
    tracesSampleRate: 1.0,
    tracePropagationTargets: ['https://interledger.app'],
    replaysSessionSampleRate: 0,
    replaysOnErrorSampleRate: 1.0
  })
}

startTransition(() => {
  hydrateRoot(
    document,
    <StrictMode>
      <RemixBrowser />
    </StrictMode>
  )
})
