import type { Route } from './+types/quick-pay'
import { data, redirect } from 'react-router'
import { Form, useActionData, useRouteLoaderData } from 'react-router'
import type { MetaFunction } from 'react-router'
import { z } from 'zod'
import { type ApplicationProps, Layouts, WalletGrid, GridColumn, TextField, Button } from '~/components'
import { mergeMeta } from '~/lib/meta'
import { BackButton } from '~/components/QuickPay'
import { getSession, commitSession } from '~/session.server'
import { createError, routeAllowed, walletSchema } from '~/lib/utils.server'
import type { RootLoaderData } from '~/root'
import { type ActionData, QuickPaySession } from '~/lib/types'
import { getValidWalletAddress } from '~/lib/open-payments.server'
import { formatError } from '~/lib/helpers'
import logger from '~/lib/logger.server'

export async function loader({ request }: Route.LoaderArgs) {
  routeAllowed('OP_INTPAY_ENABLED')
}

export const handle: ApplicationProps = {
  layout: (_match, context) => context?.isUser ? Layouts.Wallet : Layouts.Marketing,
  scaffold: {
    header: { title: 'Interledger Pay' }
  }
}

export const meta: MetaFunction = mergeMeta(() => [
  { title: 'Interledger Pay' }
])

export default function Page() {
  const actionData = useActionData<ActionData>()
  const { walletAddress } = useRouteLoaderData('root') as RootLoaderData
  const errors = actionData?.errors

  return (
    <WalletGrid>
      <GridColumn className="col-span-full mt-20 mx-auto">
        <BackButton title="Home" to="/" />
        <div className="text-3xl">Pay anyone, anywhere in the world.</div>
        <Form method="POST" id="ilpay-form" className="mt-16 max-w-96">
          <TextField
            type="text"
            label="Enter your wallet address"
            placeholder="Wallet address"
            name="walletAddress"
            autoFocus
            defaultValue={walletAddress || ""}
            errorMessage={formatError(errors?.walletAddress)}
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

export async function action({ request }: Route.ActionArgs) {
  const session = await getSession(request.headers.get('Cookie'))
  const sessionData: QuickPaySession = {}
  const formData = Object.fromEntries(await request.formData())
  const result = await walletSchema.safeParseAsync(formData)

  if (!result.success) {
    const errors = z.treeifyError(result.error).properties
    return data({
      errors
    })
  }
  const walletAddress = String(formData?.walletAddress)

  try {
    const senderAddress = await getValidWalletAddress(walletAddress)
    sessionData.senderAddress = senderAddress
    session.set('quickPay', sessionData)

  } catch (err) {
    logger.error({ err }, 'Invalid wallet address')
    return data({ errors: createError("walletAddress", "Your wallet address is not valid.") })
  }

  return redirect('/quick-pay/amount', {
    headers: { 'Set-Cookie': await commitSession(session) }
  })
}
