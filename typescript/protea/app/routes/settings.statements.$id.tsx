import { LoaderArgs } from '@remix-run/node'
import { requireUserSession } from '~/lib/kratos.server'
import { grpcClient, StatusError, isGrpcError } from '~/lib/proto.server'

export async function loader({ request, params }: LoaderArgs) {
  await requireUserSession(request)
  const cookie = request.headers.get('cookie')
  let res = await grpcClient
    .getStatementPDF(
      {
        statementId: String(params.id)
      },
      {
        meta: {
          cookies: cookie || ''
        }
      }
    )
    .then((v) => v)
    .catch(StatusError)
  if (isGrpcError(res)) {
    throw res
  }

  const statement = res.response.statementPdf
  const blob = new Blob([statement], { type: 'application/pdf' })

  return new Response(blob, {
    headers: {
      'Content-Type': 'application/pdf',
      'Content-Disposition': `attachment; filename="statement.pdf"`
    }
  })
}
