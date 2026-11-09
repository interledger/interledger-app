import type { LoaderFunctionArgs } from 'react-router'
import { grpcClient } from '~/lib/proto.server'
import { reportResponse } from '~/lib/reports.server'

// Resource route: downloads the Xago users report as CSV, or PDF with ?format=pdf.
export async function loader({ request }: LoaderFunctionArgs) {
  return reportResponse(
    request,
    grpcClient.getXagoUsersReport.bind(grpcClient),
    'Xago users report'
  )
}
