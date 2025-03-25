// import { Code } from '@bufbuild/connect'
import { Code } from '@bufbuild/connect'
import type { LoaderFunctionArgs, MetaFunction } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { Outlet, useLoaderData, useLocation } from '@remix-run/react'
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
import { isConnectError } from '~/lib/error.server'
import { grpc } from '~/lib/grpc.server'
import { mergeMeta } from '~/lib/meta'

// TODO: How to differentiate between physical and virtual cards based on the
// response from GateHub?
export async function loader({ request }: LoaderFunctionArgs) {
  if (process.env.FYNBOS_ENV !== 'local') {
    throw redirect('/')
  }

  const response = await grpc.listCards(request, {})

  if (isConnectError(response)) {
    // Temporary for non-EU wallets
    // TODO: Maybe show a message to the user that cards are not available
    // in their jurisdiction?
    if (response.code === Code.FailedPrecondition) {
      throw redirect('/')
    }
    throw response.errorResponse
  }

  return json({
    cards: response.cards
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

// TODO: How to show card type when we can differentiate between phyiscal and
// virtual?
//
// Options:
//  * Chip component positioned on the right inside the CardLink
// component with Virtual or Physical
//  * Separate sections for Physical and Virtual cards
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
                  {/* TODO: Maybe show card itself instead of the icon */}
                  <Icon>credit_card</Icon>
                  <div className='flex w-full flex-col space-y-1'>
                    <span className='truncate text-medium'>
                      {card.maskedPan}
                    </span>
                    <span className='text-xs text-weak'>
                      Expires at: {card.expiryDate}
                    </span>
                  </div>
                </div>
                <div className='flex min-w-max flex-initial items-center space-x-2'>
                  <Icon>navigate_next</Icon>
                </div>
              </CardLink>
            </Card>
          ))}
        <ButtonRouter to={route('/cards/add')}>Add card</ButtonRouter>
      </GridColumn>
      <GridColumn sticky className='col-span-full lg:col-span-6 lg:col-start-7'>
        <Outlet context={cards} />
      </GridColumn>
    </WalletGrid>
  )
}
