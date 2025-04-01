import type { DataFunctionArgs, EntryContext } from '@remix-run/node'
import { createReadableStreamFromReadable } from '@remix-run/node'
import { RemixServer, isRouteErrorResponse } from '@remix-run/react'
import * as Sentry from '@sentry/remix'
import isbot from 'isbot'
import { renderToPipeableStream } from 'react-dom/server'
import { PassThrough } from 'stream'

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
  if (error instanceof Error) {
    Sentry.captureRemixServerException(error, 'remix.server', request).catch(
      (e) => {
        console.error('Error capturing error', e)
      }
    )
  } else {
    // Opt out for 404 errors
    if (isRouteErrorResponse(error) && error.status === 404) {
      return
    }
    Sentry.captureException(error)
  }
  console.error(error)
}

export default function handleRequest(
  request: Request,
  responseStatusCode: number,
  responseHeaders: Headers,
  remixContext: EntryContext
) {
  // 1. Generate a nonce for each request (recommended for security)
  //    You might skip this if using a simpler policy without nonces initially.
  // const nonce = crypto.randomBytes(16).toString("hex")

  // 2. Define your CSP policy string
  //    Start restrictive and loosen as needed.
  //    Using the nonce for scripts is highly recommended over 'unsafe-inline'.
  const csp = [
    // `default-src 'self'`,
    // Allow scripts from 'self' and inline scripts using the generated nonce
    // `script-src 'self' 'nonce-${nonce}' 'strict-dynamic'`,
    // Allow styles from 'self' and potentially 'unsafe-inline' if needed,
    // though external CSS files are better. Consider nonces for inline styles too if applicable.
    // `style-src 'self' 'unsafe-inline'`,
    // `img-src 'self' data:`, // Allow images from self and data URIs
    // `font-src 'self'`,
    // `connect-src 'self'`, // Control fetch/XHR/WebSockets
    // `frame-src 'self'`, // Control frames
    `object-src 'none'`, // Disallow plugins (Flash, etc.)
    `base-uri 'self'`,
    `form-action 'self'`,
    `frame-ancestors 'none'` // Prevent clickjacking
    // Add other directives as needed (e.g., connect-src for APIs, img-src for CDNs)
  ].join('; ')

  // 3. Add the CSP header to the response headers
  responseHeaders.set('Content-Security-Policy', csp)

  // 4. Some other security ones
  responseHeaders.set(
    'Strict-Transport-Security',
    'max-age=31536000; includeSubDomains'
  )
  responseHeaders.set('X-Content-Type-Options', 'nosniff')
  // responseHeaders.set('X-Frame-Options', 'SAMEORIGIN')

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
          console.error(error)
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
          console.error(error)
          responseStatusCode = 500
        }
      }
    )

    setTimeout(abort, ABORT_DELAY)
  })
}
