import type {
  LoaderFunctionArgs,
  MetaFunction
} from '@remix-run/node'
import { json } from '@remix-run/node'
import { Link, useLoaderData } from '@remix-run/react'
import { useEffect } from 'react'
import type { ApplicationProps } from '~/components'
import { Button, GridColumn, Layouts, WalletGrid } from '~/components'
import { DialPad, DialPadIds } from '~/components/QuickPay/Dialpad'
import { useDialPadContext } from '~/lib/context/dialpad'
import { mergeMeta } from '~/lib/meta'
import { getSession } from '~/session.server'
import { getUserSession } from '~/lib/kratos.server'
import { isWalletLayout} from '~/lib/utils'

export async function loader({ request }: LoaderFunctionArgs) {
  const isWalletView = await isWalletLayout(request)

  const session = await getSession(request.headers.get('Cookie'))
  const walletAddressInfo = session.get('quickPay')
  const assetCode = walletAddressInfo?.validWalletAddress?.assetCode

  if (walletAddressInfo === undefined || assetCode === undefined) {
    throw json(
      {
        code: "QUICKPAY_SESSION_ERROR",
        message: "Payment session expired."
      },
      { status: 400 }
    )
  }
  return json({
    assetCode,
    isWalletView
  } as const)
}

export const handle: ApplicationProps = {
  layout: (match) =>
    match.data?.isWalletView ? Layouts.Wallet : Layouts.Marketing,
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
  })

  const processAmount = (e: React.MouseEvent<HTMLElement>) => {
    if (
      amountValue.indexOf(DialPadIds.Dot) === -1 ||
      amountValue.endsWith(DialPadIds.Dot)
    ) {
      setAmountValue(Number(amountValue).toFixed(2).toString())
    }

    if (Number(amountValue) === 0) {
      e.preventDefault()
    }
  }

  return (
    <WalletGrid>
      <GridColumn
        className='col-span-full mt-20 mx-auto'
      >
        <DialPad />
        <div className="flex justify-center gap-2 mt-12 w-64">
          <Link
            to={`/quick-pay/request`}
            onClick={processAmount}
            className='min-w-28'
          >
            <Button
              aria-label="request"
              disabled={Number(amountValue) === 0}
            >
              Request
            </Button>
          </Link>
          <Link
            to={`/quick-pay/pay`}
            onClick={processAmount}
            className='min-w-28'
          >
            <Button
              aria-label="pay"
              disabled={Number(amountValue) === 0}
            >
              Pay
            </Button>
          </Link>
        </div>
      </GridColumn>
    </WalletGrid>
  )
}

