import { LoaderFunction, redirect, useLoaderData } from 'remix'
import { route } from 'routes-gen'
import { requireUserSession } from '~/lib/kratos'

export const loader: LoaderFunction = async ({ request }) => {
  const url = new URL(request.url)
  const flowId = url.searchParams.get('flow')
  if (flowId)
    return redirect(`${route('/settings/password')}?flow=${flowId}`, {
      headers: request.headers
    })

  const session = await requireUserSession(request)
  return session
}

export default function SettingsPage() {
  const session = useLoaderData()
  return (
    // <WalletLayout
    //   route={Routes.settings}
    //   backRoute={Routes.walletHome}
    //   header='Settings'
    //   hideNav
    // >
    <div>{session?.identity.traits.email}</div>
    //   &#9679;&#9679;&#9679;&#9679;&#9679;&#9679;&#9679;&#9679;
    // </WalletLayout>
  )
}
