import type {
  ActionFunctionArgs,
  LoaderFunctionArgs,
  MetaFunction
} from '@remix-run/node'
import { useLoaderData, useRouteLoaderData } from '@remix-run/react'
import type { ApplicationProps } from '~/components'
import { Layouts } from '~/components'
import { mergeMeta } from '~/lib/meta'
import { getSession, commitSession } from '~/session.server'
import { Form, useActionData } from '@remix-run/react'
import { WalletGrid, GridColumn, TextField, Button } from '~/components'
import { getValidWalletAddress } from '~/lib/utils'
import { json, redirect } from '@remix-run/node'
import { getUserSession } from '~/lib/kratos.server'
import { type SerializeFrom } from '@remix-run/node'
import type { loader as rootLoader } from '~/root'
import { type ActionData } from "~/lib/types"

export async function loader({ request }: LoaderFunctionArgs) {
  let isLoggedIn

  try {
    await getUserSession(request)
    isLoggedIn = true

  } catch (err) {
    isLoggedIn = false
  }
  return json({
    isLoggedIn
  })
}

export const handle: ApplicationProps = {
  layout: (match) =>
    match.data?.isLoggedIn ? Layouts.Wallet : Layouts.Marketing,
  scaffold: {
    header: { title: 'Interledger Pay' }
  }
}

export const meta: MetaFunction = mergeMeta(() => [
  { title: 'Interledger Pay' }
])

export default function Page() {
  const actionData = useActionData<ActionData>()
  const { walletAddress } = useRouteLoaderData("root") as SerializeFrom<
    typeof rootLoader
  >

  return (
    <WalletGrid>
      <GridColumn className="col-span-full mt-20 mx-auto">
        <div className="text-3xl">Pay anyone, anywhere in the world.</div>

        <Form method="POST" id="ilpay-form" className="mt-16 max-w-96">
          <TextField
            type="text"
            label="Enter your wallet address"
            placeholder="Wallet address"
            name="walletAddress"
            autoFocus
            value={walletAddress || ""}
            errorMessage={String(actionData?.errors?.walletAddress || '')}
          />

          <Button
            form="ilpay-form"
            type="submit"
            name="intent"
            className="max-w-xs mt-12 mx-auto"
          >
            Pay now
          </Button>
        </Form>
      </GridColumn>
    </WalletGrid>
  )
}

export async function action({ request }: ActionFunctionArgs) {
  const session = await getSession(request.headers.get('Cookie'))
  const formData = Object.fromEntries(await request.formData())
  const walletAddress = String(formData?.walletAddress)

  session.set('quickPay', {
    walletAddress: walletAddress
  })

  try {
    const validWalletAddress = await getValidWalletAddress(walletAddress)
    session.set('quickPay', {
      validWalletAddress: validWalletAddress
    })

  } catch (err) {
    console.log({ err })
    return json({ errors: { walletAddress: "Your wallet address is not valid." } })
  }

  return redirect('/quick-pay/amount', {
    headers: { 'Set-Cookie': await commitSession(session) }
  })
}
