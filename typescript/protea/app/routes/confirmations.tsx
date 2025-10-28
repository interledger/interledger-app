import type { LoaderFunctionArgs, MetaFunction } from '@remix-run/node'
import { json } from '@remix-run/node'
import type { ShouldRevalidateFunction } from '@remix-run/react'
import { Outlet, useLoaderData, useLocation } from '@remix-run/react'
import { route } from 'routes-gen'
import type { ApplicationProps } from '~/components'
import {
  Card,
  CardLink,
  Chip,
  ChipColor,
  Fab,
  GridColumn,
  Icon,
  Layouts,
  TimeoutDisplay,
  WalletGrid
} from '~/components'
import { mergeMeta } from '~/lib/meta'
import { mockPendingConfirmations } from '~/lib/mocks/confirmations'

export const shouldRevalidate: ShouldRevalidateFunction = ({
  currentUrl,
  defaultShouldRevalidate,
  nextUrl
}) => {
  if (currentUrl.search !== nextUrl.search) return false
  return defaultShouldRevalidate
}

export async function loader({ request }: LoaderFunctionArgs) {
  // For now, using mock data
  // TODO: Replace with actual API call: await grpc.getPending3DSConfirmations(request, {})

  const confirmations = mockPendingConfirmations

  // Format dates and group confirmations by date
  const formattedConfirmations = confirmations.map((confirmation) => ({
    transactionId: confirmation.transactionId,
    merchantName: confirmation.merchantName,
    purchaseDate: confirmation.purchaseDate,
    timeout: confirmation.timeout,
    formattedDate: new Date(
      Number(confirmation.purchaseDate)
    ).toLocaleDateString('en-US', {
      month: 'long',
      day: 'numeric',
      year: 'numeric'
    }),
    formattedTime: new Date(
      Number(confirmation.purchaseDate)
    ).toLocaleTimeString('en-US', {
      hour: 'numeric',
      minute: '2-digit',
      hour12: true
    }),
    formattedAmount: `${parseFloat(confirmation.purchaseAmount).toFixed(2)} ${
      confirmation.purchaseCurrency
    }`
  }))

  return json({
    confirmations: formattedConfirmations
  })
}

export const handle: ApplicationProps = {
  layout: Layouts.Wallet,
  scaffold: {
    header: {
      title: 'Pending Confirmations'
    },
    fab: Fab.Pay
  }
}

export const meta: MetaFunction = mergeMeta(() => [
  {
    title: 'Pending Confirmations'
  }
])

export default function Page() {
  const { confirmations } = useLoaderData<typeof loader>()
  const location = useLocation()
  const pathSegments = location.pathname.split('/').filter(Boolean)

  const isMobile = typeof document !== 'undefined' && window.innerWidth < 1024

  return (
    <WalletGrid>
      <GridColumn
        hideOnMobile={pathSegments[pathSegments.length - 1] !== 'confirmations'}
        className='col-span-full lg:col-span-6'
      >
        {confirmations && confirmations.length === 0 && (
          <Card>
            <div className='flex flex-col space-y-4 p-4'>
              <span className='text-sm text-medium'>
                You have no pending confirmations. Card purchase confirmations
                will appear here.
              </span>
            </div>
          </Card>
        )}

        {confirmations && (
          <Card key={`confirmations`}>
            {confirmations.map((confirmation) => (
              <CardLink
                preventScrollReset={!isMobile}
                prefetch='none'
                key={confirmation.transactionId}
                to={route('/confirmations/:confirmationId', {
                  confirmationId: confirmation.transactionId
                })}
                className='justify-between space-x-4'
              >
                <div className='flex w-7/12 items-center space-x-2'>
                  <div className='flex flex-col items-center'>
                    <Icon className='text-orange-500'>schedule</Icon>
                    <TimeoutDisplay
                      purchaseDate={confirmation.purchaseDate}
                      timeout={confirmation.timeout}
                    />
                  </div>
                  <div className='flex w-full flex-col space-y-1'>
                    <span className='truncate text-medium'>
                      {confirmation.merchantName}
                    </span>
                    <span className='text-xs text-weak'>
                      {confirmation.formattedTime}
                    </span>
                  </div>
                </div>
                <div className='flex min-w-max flex-initial items-center space-x-2'>
                  <div className='flex flex-col items-end space-y-1'>
                    <span className='min-w-max font-medium text-medium'>
                      {confirmation.formattedAmount}
                    </span>
                    <Chip color={ChipColor.orange}>Pending</Chip>
                  </div>
                  <Icon>navigate_next</Icon>
                </div>
              </CardLink>
            ))}
          </Card>
        )}
      </GridColumn>
      <GridColumn sticky className='col-span-full lg:col-span-6 lg:col-start-7'>
        <Outlet />
      </GridColumn>
    </WalletGrid>
  )
}
