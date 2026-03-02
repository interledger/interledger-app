import type { ActionFunctionArgs } from 'react-router';
import logger, { addRequestId, withErrorLog } from '~/lib/logger.server'
import { extractOrGenerateRequestId } from '~/lib/requestContext.server'

export async function action({ request }: ActionFunctionArgs) {
  try {
    let expectedDsn = process.env.SENTRY_DSN || ''
    let envelope = await request.text()
    let header = envelope.split('\n')[0]
    let headerObject = JSON.parse(header)
    if (typeof headerObject.dsn == 'undefined' || headerObject.dsn == '') {
      return new Response(null, { status: 404 })
    }

    // only allow requests for our dsn
    if (headerObject.dsn !== expectedDsn) {
      return new Response(null, { status: 404 })
    }

    let url = new URL(expectedDsn)
    let projectID = url.pathname.replace('/', ``)
    await fetch(`https://${url.hostname}/api/${projectID}/envelope`, {
      method: 'POST',
      body: envelope,
      headers: { 'Content-Type': 'application/x-sentry-envelope' }
    })

    return new Response(null, { status: 200 })
  } catch (error) {
    const requestId = extractOrGenerateRequestId(request)
    logger.error(
      { ...addRequestId(requestId), ...withErrorLog(error) },
      'Sentry envelope processing error'
    )
    return new Response(null, { status: 404 })
  }
}
