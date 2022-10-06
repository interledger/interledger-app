import { renderToString } from 'react-dom/server'
import type { EntryContext } from '@remix-run/node'
import { RemixServer } from '@remix-run/react'
import { sdk } from '~/lib/otel.server'

if (process.env.OTEL_EXPORTER_OTLP_HEADERS) {
  sdk.start().catch((error) => {
    console.log(error)
  })
}

export default function handleRequest(
  request: Request,
  responseStatusCode: number,
  responseHeaders: Headers,
  remixContext: EntryContext
) {
  const markup = renderToString(
    <RemixServer context={remixContext} url={request.url} />
  )

  responseHeaders.set('Content-Type', 'text/html')

  return new Response('<!DOCTYPE html>' + markup, {
    status: responseStatusCode,
    headers: responseHeaders
  })
}
