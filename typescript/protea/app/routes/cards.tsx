import type { LoaderFunctionArgs, MetaFunction } from '@remix-run/node'
import { json } from '@remix-run/node'
import { Outlet, useLoaderData, useLocation } from '@remix-run/react'
import crypto from 'node:crypto'
import { route } from 'routes-gen'
import type { ApplicationProps } from '~/components'
import {
  ButtonRouter,
  Card,
  CardContent,
  CardLink,
  GridColumn,
  Icon,
  Layouts,
  WalletGrid
} from '~/components'
import { Card as GRPCCard } from '~/generated/connect/backend/v1/backend_pb'
// import { isConnectError } from '~/lib/error.server'
// import { grpc } from '~/lib/grpc.server'
import { mergeMeta } from '~/lib/meta'

export async function loader(_args: LoaderFunctionArgs) {
  // const response = await grpc.listCards(request, {})
  //
  //
  // if (isConnectError(response)) {
  //   console.log(response.code)
  //   if (response.code === Code.Internal) {
  //     throw redirect('/')
  //   }
  //   throw response.errorResponse
  // }

  // return json({
  //   cards: response.cards
  // })

  // Dummy data for now
  return json({
    cards: Array.from(
      { length: 10 },
      () =>
        new GRPCCard({
          id: crypto.randomUUID(),
          nameOnCard: 'Test',
          expiryDate: '03/30'
        })
    ) as GRPCCard[]
  })
}

export const handle: ApplicationProps = {
  layout: Layouts.Wallet,
  scaffold: {
    header: {
      title: 'Cards'
    }
  }
}

export const meta: MetaFunction = mergeMeta(() => [
  {
    title: 'Cards'
  }
])

export default function Page() {
  const { cards } = useLoaderData<typeof loader>()
  const location = useLocation()

  const isMobile = typeof document !== 'undefined' && window.innerWidth < 1024
  const pathSegments = location.pathname.split('/').filter(Boolean)

  return (
    <WalletGrid>
      <GridColumn
        hideOnMobile={pathSegments[pathSegments.length - 1] !== 'cards'}
        className='col-span-full lg:col-span-6'
      >
        {cards.length === 0 && (
          <Card>
            <CardContent>
              It looks like you don't have any cards associated with your
              account.
            </CardContent>
          </Card>
        )}
        {cards.length > 0 &&
          cards.map((card) => (
            <Card key={card.id}>
              <CardLink
                preventScrollReset={!isMobile}
                prefetch='none'
                to={route('/cards/:cardId', {
                  cardId: card.id
                })}
                className='justify-between space-x-4'
              >
                <div className='flex w-7/12 items-center space-x-4'>
                  <Icon>credit_card</Icon>
                  <div className='flex w-full flex-col space-y-1'>
                    <span className='truncate text-medium'>{card.id}</span>
                    <span className='text-xs text-weak'>{card.expiryDate}</span>
                  </div>
                </div>
                <div className='flex min-w-max flex-initial items-center space-x-2'>
                  <Icon>navigate_next</Icon>
                </div>
              </CardLink>
            </Card>
          ))}
        <ButtonRouter to={route('/cards/order')}>Order card</ButtonRouter>
      </GridColumn>
      <GridColumn sticky className='col-span-full lg:col-span-6 lg:col-start-7'>
        {/*
         * TODO: Pass outlet context to avoid refetching the card details.
         * Based on the API reference there might be no difference in information
         * when fetching all cards for the current customer vs. fetching an
         * indivifual card details.
         *
         * With this approach we avoid having another loader in the `card/:cardId` route.
         */}
        <Outlet />
      </GridColumn>
    </WalletGrid>
  )
}
