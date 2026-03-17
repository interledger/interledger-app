import {
  json,
  redirect,
  type LoaderFunctionArgs,
  type MetaFunction
} from '@remix-run/node'
import type { ShouldRevalidateFunction } from '@remix-run/react'
import { Outlet, useLocation } from '@remix-run/react'
import { route } from 'routes-gen'
import type { ApplicationProps } from '~/components'
import {
  Card,
  CardLink,
  Fab,
  GridColumn,
  Icon,
  Layouts,
  TimeoutDisplay,
  WalletGrid
} from '~/components'
import { usePendingConfirmations } from '~/lib/cards/usePendingConfirmations'
import { hasUserSession } from '~/lib/kratos/session.server'

export async function loader({ request }: LoaderFunctionArgs) {
  const session = hasUserSession(request)
  const returnTo = new URL(request.url).pathname

  if (!session) {
    throw redirect(
      `${route('/login')}?returnTo=${encodeURIComponent(returnTo)}`
    )
  }

  return json({})
}

export const shouldRevalidate: ShouldRevalidateFunction = ({
  currentUrl,
  defaultShouldRevalidate,
  nextUrl
}) => {
  if (currentUrl.search !== nextUrl.search) return false
  return defaultShouldRevalidate
}

export const handle: ApplicationProps = {
  layout: Layouts.Wallet,
  scaffold: {
    header: {
      title: 'Pending 3DS Confirmations'
    },
    fab: Fab.Pay
  }
}

export const meta: MetaFunction = () => [
  {
    title: 'Pending 3DS Confirmations'
  }
]

export default function Page() {
  const { pendingConfirmations: confirmations } = usePendingConfirmations()
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

        {confirmations.length > 0 && (
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
                <div className='flex w-7/12 items-center space-x-6'>
                  <div className='flex w-8 flex-col items-center space-y-1'>
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
                      {confirmation.purchaseAmount}{' '}
                      {confirmation.purchaseCurrency}
                    </span>
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
