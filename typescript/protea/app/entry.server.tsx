import type { DataFunctionArgs, EntryContext } from '@remix-run/node'
import { createReadableStreamFromReadable } from '@remix-run/node'
import { RemixServer, isRouteErrorResponse } from '@remix-run/react'
import * as Sentry from '@sentry/remix'
import isbot from 'isbot'
import { renderToPipeableStream } from 'react-dom/server'
import { PassThrough } from 'stream'
import logger, { addRequestId } from './lib/logger.server'
import { extractOrGenerateRequestId } from './lib/requestContext.server'

const ABORT_DELAY = 5_000

if (process.env.SENTRY_DSN) {
  Sentry.init({
    dsn: process.env.SENTRY_DSN,
    release: process.env.SENTRY_RELEASE,
    tracesSampleRate: 1,
    environment: process.env.FYNBOS_ENV,
    integrations: [
      new Sentry.Integrations.RequestData({
        include: {
          cookies: false
        }
      })
    ]
  })
}

export function handleError(
  error: unknown,
  { request }: DataFunctionArgs
): void {
  const requestId = extractOrGenerateRequestId(request)
  
  if (error instanceof Error) {
    Sentry.captureRemixServerException(error, 'remix.server', request).catch(
      (e) => {
        logger.error(
          { ...addRequestId(requestId), error: e instanceof Error ? e.message : String(e) },
          'Failed to capture error in Sentry'
        )
      }
    )
  } else {
    // Opt out for 404 errors
    if (isRouteErrorResponse(error) && error.status === 404) {
      return
    }
    Sentry.captureException(error)
  }
  
  logger.error(
    { ...addRequestId(requestId), error: error instanceof Error ? error.message : String(error) },
    'Unhandled error in server'
  )
}

export default function handleRequest(
  request: Request,
  responseStatusCode: number,
  responseHeaders: Headers,
  remixContext: EntryContext
) {
  return isbot(request.headers.get('user-agent'))
    ? handleBotRequest(
        request,
        responseStatusCode,
        responseHeaders,
        remixContext
      )
    : handleBrowserRequest(
        request,
        responseStatusCode,
        responseHeaders,
        remixContext
      )
}

function handleBotRequest(
  request: Request,
  responseStatusCode: number,
  responseHeaders: Headers,
  remixContext: EntryContext
) {
  return new Promise((resolve, reject) => {
    const { pipe, abort } = renderToPipeableStream(
      <RemixServer
        context={remixContext}
        url={request.url}
        abortDelay={ABORT_DELAY}
      />,
      {
        onAllReady() {
          const body = new PassThrough()

          responseHeaders.set('Content-Type', 'text/html')

          resolve(
            new Response(createReadableStreamFromReadable(body), {
              headers: responseHeaders,
              status: responseStatusCode
            })
          )

          pipe(body)
        },
        onShellError(error: unknown) {
          reject(error)
        },
        onError(error: unknown) {
          responseStatusCode = 500
          logger.error(
            { error: error instanceof Error ? error.message : String(error) },
            'Error rendering to bot'
          )
        }
      }
    )

    setTimeout(abort, ABORT_DELAY)
  })
}

function handleBrowserRequest(
  request: Request,
  responseStatusCode: number,
  responseHeaders: Headers,
  remixContext: EntryContext
) {
  return new Promise((resolve, reject) => {
    const { pipe, abort } = renderToPipeableStream(
      <RemixServer
        context={remixContext}
        url={request.url}
        abortDelay={ABORT_DELAY}
      />,
      {
        onShellReady() {
          const body = new PassThrough()

          responseHeaders.set('Content-Type', 'text/html')

          resolve(
            new Response(createReadableStreamFromReadable(body), {
              headers: responseHeaders,
              status: responseStatusCode
            })
          )

          pipe(body)
        },
        onShellError(error: unknown) {
          reject(error)
        },
        onError(error: unknown) {
          logger.error(
            { error: error instanceof Error ? error.message : String(error) },
            'Error rendering to browser'
          )
          responseStatusCode = 500
        }
      }
    )

    setTimeout(abort, ABORT_DELAY)
  })
}
