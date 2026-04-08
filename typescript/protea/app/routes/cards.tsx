import { useEffect } from 'react'
import type { ShouldRevalidateFunction } from 'react-router'
import {
  data,
  href,
  Outlet,
  redirect,
  useLoaderData,
  useLocation
} from 'react-router'
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
import { CardProcessingPlaceholder } from '~/components/Cards'
import { PhysicalCardChip, VirtualCardChip } from '~/components/Cards/CardChips'
import { getFeatures } from '~/data/wallet.server'
import { CardType } from '~/generated/connect/backend/v1/backend_pb'
import {
  panLastFour,
  toStorableCard,
  useCardsStore
} from '~/lib/cards/useCardsStore'
import { isConnectError } from '~/lib/error.server'
import { grpc } from '~/lib/grpc.server'
import { mergeMeta } from '~/lib/meta'
import type { Route } from './+types/cards'

/**
 * Let actions declare if they need revalidation via shouldRevalidate field.
 */
export const shouldRevalidate: ShouldRevalidateFunction = ({
  actionResult,
  defaultShouldRevalidate
}) => {
  // Actions explicitly declare if they need revalidation
  if (actionResult && 'shouldRevalidate' in actionResult) {
    return actionResult.shouldRevalidate === true
  }
  return defaultShouldRevalidate
}

export async function loader({ request }: Route.LoaderArgs) {
  const features = await getFeatures(request)
  if (!features.manageWalletCardsEnabled) {
    throw redirect('/')
  }

  const response = await grpc.listCards(request, {})

  if (isConnectError(response)) {
    throw response.errorResponse
  }

  return data({
    isWaitingForCreation: response.isWaitingForCreation,
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

export const meta = mergeMeta(() => [
  {
    title: 'Cards'
  }
])

export default function Page() {
  const { cards, isWaitingForCreation } = useLoaderData<typeof loader>()
  const location = useLocation()
  const { setCards } = useCardsStore()

  useEffect(() => {
    const storableCards = cards.map((c) => toStorableCard(c))
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
        {isWaitingForCreation ? (
          <CardProcessingPlaceholder />
        ) : (
          <>
            {cards.length === 0 && (
              <Card>
                <CardContent>
                  It looks like you don't have any cards associated with your
                  account.
                </CardContent>
              </Card>
            )}
            {cards.length > 0 &&
              cards.map((card: import('~/lib/cards/types').SerializedCard) => (
                <Card key={card.id} className='relative'>
                  <CardLink
                    preventScrollReset={!isMobile}
                    prefetch='intent'
                    to={href('/cards/:cardId', {
                      cardId: card.id
                    })}
                    className='justify-between space-x-4'
                  >
                    <div className='flex w-7/12 items-center space-x-4'>
                      {/* TODO: Maybe show card itself instead of the icon */}
                      <Icon>credit_card</Icon>
                      <span className='truncate text-medium'>
                        &bull;&bull;&bull;&bull; {panLastFour(card.maskedPan)}
                      </span>
                    </div>
                    <div className='flex min-w-max flex-initial items-center space-x-2'>
                      {card.type === CardType.VIRTUAL ? (
                        <VirtualCardChip />
                      ) : (
                        <PhysicalCardChip />
                      )}
                      <Icon>navigate_next</Icon>
                    </div>
                  </CardLink>
                </Card>
              ))}
            <ButtonRouter to={href('/cards/order')}>Order card</ButtonRouter>
          </>
        )}
      </GridColumn>
      <GridColumn sticky className='col-span-full lg:col-span-6 lg:col-start-7'>
        <Outlet context={cards} />
      </GridColumn>
    </WalletGrid>
  )
}
