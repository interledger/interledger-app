import type { DataFunctionArgs, EntryContext } from '@remix-run/node'
import { createReadableStreamFromReadable } from '@remix-run/node'
import { RemixServer, isRouteErrorResponse } from '@remix-run/react'
import * as Sentry from '@sentry/remix'
import isbot from 'isbot'
import { renderToPipeableStream } from 'react-dom/server'
import { PassThrough } from 'stream'
import logger, { addRequestId } from './lib/logger.server'
import { extractOrGenerateRequestId } from './lib/requestContext.server'

export const streamTimeout = 5_000

// Track request timing for logging
function getResponseTime(startTime: number): number {
  return Date.now() - startTime
}

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
  const startTime = Date.now()
  const requestId = extractOrGenerateRequestId(request)
  const url = new URL(request.url)
  
  // Log incoming request
  logger.debug(
    {
      ...addRequestId(requestId),
      method: request.method,
      url: url.pathname + url.search,
      userAgent: request.headers.get('user-agent'),
    },
    `${request.method} ${url.pathname}${url.search}`
  )

  const handler = isbot(request.headers.get('user-agent'))
    ? handleBotRequest(
        request,
        responseStatusCode,
        responseHeaders,
        remixContext,
        requestId,
        startTime
      )
    : handleBrowserRequest(
        request,
        responseStatusCode,
        responseHeaders,
        remixContext,
        requestId,
        startTime
      )

  return handler.then((response) => {
    // Log response
    logger.info(
      {
        ...addRequestId(requestId),
        method: request.method,
        url: url.pathname + url.search,
        statusCode: response.status,
        responseTime: getResponseTime(startTime),
      },
      `${request.method} ${url.pathname}${url.search} ${response.status} - ${getResponseTime(startTime)}ms`
    )
    return response
  }).catch((error) => {
    // Log error
    logger.error(
      {
        ...addRequestId(requestId),
        method: request.method,
        url: url.pathname + url.search,
        error: error instanceof Error ? error.message : String(error),
        responseTime: getResponseTime(startTime),
      },
      `${request.method} ${url.pathname}${url.search} failed`
    )
    throw error
  })
}

function handleBotRequest(
  request: Request,
  responseStatusCode: number,
  responseHeaders: Headers,
  remixContext: EntryContext,
  requestId: string,
  startTime: number
) {
  return new Promise<Response>((resolve, reject) => {
    const { pipe, abort } = renderToPipeableStream(
      <RemixServer
        context={remixContext}
        url={request.url}
        abortDelay={streamTimeout + 1000}
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
            { 
              ...addRequestId(requestId),
              error: error instanceof Error ? error.message : String(error),
              responseTime: getResponseTime(startTime),
            },
            'Error rendering to bot'
          )
        }
      }
    )

    setTimeout(abort, streamTimeout + 1000)
  })
}

function handleBrowserRequest(
  request: Request,
  responseStatusCode: number,
  responseHeaders: Headers,
  remixContext: EntryContext,
  requestId: string,
  startTime: number
) {
  return new Promise<Response>((resolve, reject) => {
    const { pipe, abort } = renderToPipeableStream(
      <RemixServer
        context={remixContext}
        url={request.url}
        abortDelay={streamTimeout + 1000}
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
            { 
              ...addRequestId(requestId),
              error: error instanceof Error ? error.message : String(error),
              responseTime: getResponseTime(startTime),
            },
            'Error rendering to browser'
          )
          responseStatusCode = 500
        }
      }
    )

    setTimeout(abort, streamTimeout + 1000)
  })
}
