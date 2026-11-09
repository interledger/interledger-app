import type { RpcOptions, UnaryCall } from '@protobuf-ts/runtime-rpc'
import type { Empty } from '~/generated/protobuf-ts/google/protobuf/empty'
import type { CsvReport } from '~/generated/protobuf-ts/backend/admin/v1/backend'

type ReportCall = (
  input: Empty,
  options?: RpcOptions
) => UnaryCall<Empty, CsvReport>

/**
 * Builds a CSV report on the backend and returns it as a file download.
 * @param request The incoming request; its cookies authenticate the admin.
 * @param call The grpcClient method that builds the report.
 * @param label Human-readable report name, used in the error message.
 */
export async function csvReportResponse(
  request: Request,
  call: ReportCall,
  label: string
): Promise<Response> {
  try {
    const report = await call(
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
    console.log(`Failed to build ${label}: `, e)
    return new Response(`Failed to build the ${label}.`, {
      status: 500,
      headers: { 'Content-Type': 'text/plain; charset=utf-8' }
    })
  }
}
