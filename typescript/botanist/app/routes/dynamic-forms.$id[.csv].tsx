import type { LoaderArgs } from '@remix-run/node'
import { ExportDynamicFormResult } from '~/lib/wallet.server'

export async function loader({ request, params }: LoaderArgs) {
  const csv = await ExportDynamicFormResult(request, params.id as string)

  return new Response(csv, {
    status: 200,
    headers: {
      'Content-Type': 'text/csv'
    }
  })
}
