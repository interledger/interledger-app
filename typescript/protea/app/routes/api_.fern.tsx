import { envValue } from '~/env.server'
import logger, { addRequestId, withErrorLog } from '~/lib/logger.server'
import { extractOrGenerateRequestId } from '~/lib/requestContext.server'
import type { Route } from './+types/api_.fern'

export async function action({ request }: Route.ActionArgs) {
  try {
    const expectedDsn = envValue('SENTRY_DSN') || ''
    const envelope = await request.text()
    const header = envelope.split('\n')[0]
    const headerObject = JSON.parse(header)
    if (typeof headerObject.dsn == 'undefined' || headerObject.dsn == '') {
      return new Response(null, { status: 404 })
    }

    // only allow requests for our dsn
    if (headerObject.dsn !== expectedDsn) {
      return new Response(null, { status: 404 })
    }

    const url = new URL(expectedDsn)
    const projectID = url.pathname.replace('/', ``)
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
