import type {
  LoaderFunctionArgs,
  MetaFunction
} from '@remix-run/node'
import { json } from '@remix-run/node'
import { useLoaderData } from '@remix-run/react'
import { useEffect } from 'react'
import type { ApplicationProps } from '~/components'
import { Button, GridColumn, Layouts, WalletGrid } from '~/components'
import { BackButton } from '~/components/QuickPay'
import { DialPad, DialPadIds } from '~/components/QuickPay/Dialpad'
import { useDialPadContext } from '~/lib/context/dialpad'
import { mergeMeta } from '~/lib/meta'
import { getSession } from '~/session.server'
import { getUserSession } from '~/lib/kratos.server'

export async function loader({ request }: LoaderFunctionArgs) {
  let isLoggedIn

  try {
    await getUserSession(request)
    isLoggedIn = true

  } catch (err) {
    isLoggedIn = false
  }

  const session = await getSession(request.headers.get('Cookie'))
  const walletAddressInfo = session.get('quickPay')
  const assetCode = walletAddressInfo?.validWalletAddress?.assetCode

  if (walletAddressInfo === undefined || assetCode === undefined) {
    throw json(
      {
        code: "QUICKPAY_SESSION_ERROR",
        title: "Payment session expired."
      },
      { status: 400 }
    )
  }
  return json({
    assetCode,
    isLoggedIn
  } as const)
}

export const handle: ApplicationProps = {
  layout: (match) =>
    match.data?.isLoggedIn ? Layouts.Wallet : Layouts.Marketing,
  scaffold: {
    header: { title: 'Interledger Pay' }
  }
}

export const meta: MetaFunction = mergeMeta(() => [
  {
    title: 'Interledger Pay'
  }
])

export default function Page() {
  const { assetCode } = useLoaderData<typeof loader>()
  const { amountValue, setAmountValue, setAssetCode } = useDialPadContext()

  useEffect(() => {
    setAssetCode(assetCode)
  }, [setAssetCode, assetCode])

  const handleNavigation = (e: React.MouseEvent<HTMLElement>, url: string) => {
    if (
      amountValue.indexOf(DialPadIds.Dot) === -1 ||
      amountValue.endsWith(DialPadIds.Dot)
    ) {
      setAmountValue(Number(amountValue).toFixed(2).toString())
    }

    if (Number(amountValue) === 0) {
      return e.preventDefault()
    }
    document.location = url
  }

  return (
    <WalletGrid>
      <GridColumn
        className='col-span-full mt-20 mx-auto'
      >
        <BackButton title="Back" to="/quick-pay" resetAmount/>
        <DialPad />
        <div className="flex justify-center gap-2 mt-12 w-64">
          <Button
            aria-label="request"
            onClick={e => handleNavigation(e, `/quick-pay/request`)}
            disabled={Number(amountValue) === 0}
          >
            Request
          </Button>

          <Button
            aria-label="pay"
            onClick={e => handleNavigation(e, `/quick-pay/pay`)}
            disabled={Number(amountValue) === 0}
          >
            Pay
          </Button>
        </div>
      </GridColumn>
    </WalletGrid>
  )
}

