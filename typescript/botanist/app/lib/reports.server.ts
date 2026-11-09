import type { RpcOptions, UnaryCall } from '@protobuf-ts/runtime-rpc'
import {
  ReportFormat,
  type ReportFile,
  type UsersReportRequest
} from '~/generated/protobuf-ts/backend/admin/v1/backend'

type ReportCall = (
  input: UsersReportRequest,
  options?: RpcOptions
) => UnaryCall<UsersReportRequest, ReportFile>

/**
 * Builds a report on the backend and returns it as a file download.
 * The format comes from the `format` query parameter: `pdf`, or CSV otherwise.
 * @param request The incoming request; its cookies authenticate the admin.
 * @param call The grpcClient method that builds the report.
 * @param label Human-readable report name, used in the error message.
 */
export async function reportResponse(
  request: Request,
  call: ReportCall,
  label: string
): Promise<Response> {
  const format =
    new URL(request.url).searchParams.get('format') === 'pdf'
      ? ReportFormat.PDF
      : ReportFormat.CSV

  try {
    const report = await call(
      { format },
      {
        meta: {
          cookies: String(request.headers.get('cookie')) || ''
        }
      }
    )

    return new Response(report.response.content, {
      headers: {
        'Content-Type': report.response.contentType,
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
