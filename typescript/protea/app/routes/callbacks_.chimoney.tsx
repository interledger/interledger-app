import type { LoaderFunctionArgs, MetaFunction } from '@remix-run/node'
import { json } from '@remix-run/node'
import { useLoaderData } from '@remix-run/react'
import { useEffect } from 'react'
import type { ApplicationProps } from '~/components'
import { Layouts } from '~/components'
import { mergeMeta } from '~/lib/meta'

export async function loader(args: LoaderFunctionArgs) {
  const url = new URL(args.request.url)

  return json({
    issueID: url.searchParams.get('issueID'),
    status: url.searchParams.get('status'),
    kyc: url.searchParams.has('kyc')
  })
}

export const handle: ApplicationProps = {
  layout: Layouts.Focus
}

export const meta: MetaFunction = mergeMeta(() => [
  {
    title: 'Deposit'
  }
])

export default function Page() {
  const { issueID, status, kyc } = useLoaderData<typeof loader>()

  useEffect(() => {
    if (parent) {
      parent.postMessage({ issueID, status, kyc })
    }
  }, [issueID, status, kyc])
}
