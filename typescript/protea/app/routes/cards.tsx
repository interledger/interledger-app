import type { LoaderFunctionArgs, MetaFunction } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { Outlet, useLoaderData, useLocation } from '@remix-run/react'
import { useEffect } from 'react'
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
import { getFeatures } from '~/data/wallet.server'
import { Card as CardProto } from '~/generated/connect/backend/v1/backend_pb'
import { useCardsStore } from '~/lib/cards/hooks/useCardsStore'
import type { StorableCard } from '~/lib/cards/types'
import { isConnectError } from '~/lib/error.server'
import { grpc } from '~/lib/grpc.server'
import { mergeMeta } from '~/lib/meta'
import { mockCards } from '~/lib/mocks/cards'

export async function loader({ request }: LoaderFunctionArgs) {
  const features = await getFeatures(request)
  if (!features.manageWalletCardsEnabled) {
    throw redirect('/')
  }

  const response = await grpc.listCards(request, {})

  if (isConnectError(response)) {
    throw response.errorResponse
  }

  return json({
    cards: [...response.cards, ...mockCards]
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

/**
 * Convert Card object to StorableCard format for the store
 */
function convertCardToStorableCard(card: CardProto): StorableCard {
  return {
    id: card.id,
    nameOnCard: card.nameOnCard,
    maskedPan: card.maskedPan,
    unmaskedPan: null,
    expiryDate: card.expiryDate,
    status: card.status,
    lockLevel:
      // Wrong protobuf type will return lockLevels as strings
      (card.lockLevel as unknown as string) ?? 'CARD_LOCK_LEVEL_UNKNOWN',
    cvc2: null
  }
}

export const formatCardNumber = (cardNumber: string) => {
  return cardNumber.replace(/(.{4})/g, '$1 ').trim()
}

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
  const { setCards } = useCardsStore()

  useEffect(() => {
    const storableCards = cards.map((c) =>
      convertCardToStorableCard(c as CardProto)
    )
    setCards(storableCards)
  }, [cards, setCards])

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
                      {formatCardNumber(card.maskedPan)}
                    </span>
                    <span className='text-xs text-weak'>
                      Expires at: {card.expiryDate.slice(0, 2)}/
                      {card.expiryDate.slice(2)}
                    </span>
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
        <Outlet context={cards} />
      </GridColumn>
    </WalletGrid>
  )
}
