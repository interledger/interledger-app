import type { Route } from './+types/quick-pay_.amount'
import { data } from 'react-router'
import { useLoaderData, useNavigate } from 'react-router'
import type { MetaFunction } from 'react-router'
import { useEffect } from 'react'
import { flushSync } from 'react-dom'
import type { ApplicationProps } from '~/components'
import { Button, GridColumn, Layouts, WalletGrid } from '~/components'
import { DialPad, DialPadIds } from '~/components/QuickPay/Dialpad'
import { useDialPadContext } from '~/lib/context/dialpad'
import { mergeMeta } from '~/lib/meta'
import { getSession } from '~/session.server'
import { routeAllowed } from '~/lib/utils.server'

export async function loader({ request }: Route.LoaderArgs) {
  routeAllowed('OP_INTPAY_ENABLED')
  const session = await getSession(request.headers.get('Cookie'))
  const walletAddressInfo = session.get('quickPay')
  const assetCode = walletAddressInfo?.validWalletAddress?.assetCode

  if (walletAddressInfo === undefined || assetCode === undefined) {
    throw data(
      {
        code: "QUICKPAY_SESSION_ERROR",
        title: "Payment session expired."
      },
      { status: 400 }
    )
  }
  return data({
    assetCode,
  } as const)
}

export const handle: ApplicationProps = {
  layout: (_match, context) => context?.isUser ? Layouts.Wallet : Layouts.Marketing,
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
  const navigate = useNavigate()
  const { assetCode } = useLoaderData<typeof loader>()
  const { amountValue, setAmountValue, setAssetCode } = useDialPadContext()

  useEffect(() => {
    setAssetCode(assetCode)
  }, [setAssetCode, assetCode])

  const handleNavigation = (e: React.MouseEvent<HTMLElement>, url: string) => {
    let normalizedAmount: string = amountValue
    if (
      amountValue.indexOf(DialPadIds.Dot) === -1 ||
      amountValue.endsWith(DialPadIds.Dot)
    ) {
      normalizedAmount = Number(amountValue).toFixed(2).toString()
    }

    if (Number(amountValue) === 0) {
      return e.preventDefault()
    }

    flushSync(() => {
      setAmountValue(normalizedAmount)
    })

    navigate(url)
  }

  return (
    <WalletGrid>
      <GridColumn
        className='col-span-full mt-20 mx-auto'
      >
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
