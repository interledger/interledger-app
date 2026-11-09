import type { LoaderFunctionArgs } from 'react-router'
import { grpcClient } from '~/lib/proto.server'
import { reportResponse } from '~/lib/reports.server'

// Resource route: downloads the GateHub users report as CSV, or PDF with ?format=pdf.
export async function loader({ request }: LoaderFunctionArgs) {
  return reportResponse(
    request,
    grpcClient.getGatehubUsersReport.bind(grpcClient),
    'GateHub users report'
  )
}
