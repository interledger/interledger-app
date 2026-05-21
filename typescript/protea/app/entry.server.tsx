

import type { EntryContext } from 'react-router';
import { createReadableStreamFromReadable } from '@react-router/node';
import { ServerRouter, isRouteErrorResponse } from 'react-router';
import * as Sentry from '@sentry/react-router'
import isbot from 'isbot'
import { renderToPipeableStream } from 'react-dom/server'
import { PassThrough } from 'stream'
import logger, { addRequestId } from './lib/logger.server'
import { extractOrGenerateRequestId } from './lib/requestContext.server'
import { envValue, envVarValidation } from './env.server'

envVarValidation()

export const streamTimeout = 5_000

// Track request timing for logging
function getResponseTime(startTime: number): number {
  return Date.now() - startTime
}

if (envValue("SENTRY_DSN")) {
  Sentry.init({
    dsn: envValue("SENTRY_DSN"),
    release: envValue("SENTRY_RELEASE"),
    tracesSampleRate: 1,
    environment: envValue("SENTRY_ENV_LABEL"),
    integrations: [
      Sentry.requestDataIntegration({
        include: {
          cookies: false
        }
      })
    ]
  })
}

export function handleError(
  error: unknown,
  { request }: { request: Request }
): void {
  const requestId = extractOrGenerateRequestId(request)

  if (error instanceof Error) {
    Sentry.captureException(error)
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
  reactRouterContext: EntryContext
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
      reactRouterContext,
      requestId,
      startTime
    )
    : handleBrowserRequest(
      request,
      responseStatusCode,
      responseHeaders,
      reactRouterContext,
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
  reactRouterContext: EntryContext,
  requestId: string,
  startTime: number
) {
  return new Promise<Response>((resolve, reject) => {
    const { pipe, abort } = renderToPipeableStream(
      <ServerRouter
        context={reactRouterContext}
        url={request.url}
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
  });
}

function handleBrowserRequest(
  request: Request,
  responseStatusCode: number,
  responseHeaders: Headers,
  reactRouterContext: EntryContext,
  requestId: string,
  startTime: number
) {
  return new Promise<Response>((resolve, reject) => {
    const { pipe, abort } = renderToPipeableStream(
      <ServerRouter
        context={reactRouterContext}
        url={request.url}
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
  });
}
