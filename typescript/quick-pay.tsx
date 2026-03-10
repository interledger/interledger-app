import type { MetaFunction, ActionFunctionArgs } from '@remix-run/node'
import { useCallback, useState } from 'react'
import { useActionData, Form } from '@remix-run/react'
import { json, redirect } from '@remix-run/node'
import { z } from 'zod'

import type { ApplicationProps } from '~/components'
import { Layouts } from '~/components'
import { mergeMeta } from '~/lib/meta'
import { getSession, commitSession } from '~/session.server'
import { getValidWalletAddress } from '~/lib/validators.server'
import { WalletGrid, GridColumn, TextField, Button } from '~/components'

// --- Zod schema (no conform-to)
const walletSchema = z.object({
  walletAddress: z.string().min(1, 'Please enter a wallet address.')
})

export const handle: ApplicationProps = {
  layout: Layouts.Marketing,
  scaffold: {
    header: { title: 'Interledger Pay' }
  }
}

export const meta: MetaFunction = mergeMeta(() => [
  { title: 'Interledger Pay' }
])

export default function Page() {
  const actionData = useActionData<typeof action>()

  const [height, setHeight] = useState<null | number>(null)
  const divHeight = useCallback((node: HTMLDivElement | null) => {
    if (node) setHeight(node.getBoundingClientRect().height)
  }, [])

  return (
    <>
      <WalletGrid ref={divHeight}>
        <GridColumn className="col-span-full mt-20 mx-auto">
          <div className="text-3xl">Pay anyone, anywhere in the world.</div>

          <Form method="POST" id="ilpay-form" className="mt-16 max-w-96">
            <TextField
              type="text"
              label="Enter your wallet address"
              placeholder="Wallet address"
              name="walletAddress"
              autoFocus
              errorMessage={String(actionData?.errors?.walletAddress)}
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
    </>
  )
}

export async function action({ request }: ActionFunctionArgs) {
  const session = await getSession(request.headers.get('Cookie'))
  const formData = await request.formData()

  const fields = {
    walletAddress: formData.get('walletAddress')
  }

  // Validate with Zod manually
  const result = await walletSchema.safeParseAsync(fields)

  if (!result.success) {
    return json({
      status: 'error',
      fields,
      errors: result.error.flatten().fieldErrors
    })
  }

  // Now validate the wallet address itself
  try {
    const walletAddress = await getValidWalletAddress(result.data.walletAddress)

    session.set('wallet-address', { walletAddress })

    return redirect('/quick-pay/amount', {
      headers: {
        'Set-Cookie': await commitSession(session)
      }
    })
  } catch (err) {
    return json({
      status: 'error',
      fields,
      errors: { walletAddress: 'Your wallet address is not valid.' }
    })
  }
}