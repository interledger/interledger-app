import { HydratedRouter } from 'react-router/dom';
import * as Sentry from '@sentry/react-router'
import { StrictMode, startTransition } from 'react'
import { hydrateRoot } from 'react-dom/client'

if (
  typeof window.ENV !== 'undefined' &&
  window.ENV.sentryDsn
) {
  Sentry.init({
    tunnel: '/api/fern',
    dsn: window.ENV.sentryDsn,
    release: window.ENV.sentryRelease,
    integrations: [
      Sentry.browserTracingIntegration(),
      Sentry.replayIntegration()
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
      <HydratedRouter />
    </StrictMode>
  )
})
