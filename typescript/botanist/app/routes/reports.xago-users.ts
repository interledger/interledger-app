import type { LoaderFunctionArgs } from 'react-router'
import { grpcClient } from '~/lib/proto.server'
import { csvReportResponse } from '~/lib/reports.server'

// Resource route: downloads the Xago users report as CSV.
export async function loader({ request }: LoaderFunctionArgs) {
  return csvReportResponse(
    request,
    grpcClient.getXagoUsersReport.bind(grpcClient),
    'Xago users report'
  )
}
