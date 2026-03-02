import type { LoaderFunctionArgs, MetaFunction } from 'react-router';
import { data } from 'react-router';
import { useLoaderData } from 'react-router';
import { useEffect } from 'react'
import type { ApplicationProps } from '~/components'
import { Layouts } from '~/components'
import { mergeMeta } from '~/lib/meta'

export async function loader(args: LoaderFunctionArgs) {
  const url = new URL(args.request.url)

  return data({
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
