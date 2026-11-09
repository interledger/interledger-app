import type { LoaderFunctionArgs } from 'react-router'
import { grpcClient } from '~/lib/proto.server'

// Resource route: streams the Xago users report as a CSV download.
export async function loader({ request }: LoaderFunctionArgs) {
  try {
    const report = await grpcClient.getXagoUsersReport(
      {},
      {
        meta: {
          cookies: String(request.headers.get('cookie')) || ''
        }
      }
    )

    return new Response(report.response.content, {
      headers: {
        'Content-Type': 'text/csv; charset=utf-8',
        'Cache-Control': 'no-store',
        'Content-Disposition': `attachment; filename="${report.response.filename}"`
      }
    })
  } catch (e) {
    console.log('Failed to build Xago users report: ', e)
    return new Response('Failed to build the Xago users report.', {
      status: 500,
      headers: { 'Content-Type': 'text/plain; charset=utf-8' }
    })
  }
}
