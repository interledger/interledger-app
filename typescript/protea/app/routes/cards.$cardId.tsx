import { type SerializeFrom } from '@remix-run/node'
import { useNavigate, useParams, useRouteLoaderData } from '@remix-run/react'
import { useMemo } from 'react'
import { type RouteParams } from 'routes-gen'
import { CardView, Layouts, type ApplicationProps } from '~/components'
import { useGateHubStore } from '~/lib/gatehub/hooks/useGateHubStore'
import type { loader as rootLoader } from '~/root'

export const handle: ApplicationProps = {
  layout: Layouts.Wallet,
  scaffold: {
    header: {
      title: 'Cards'
    }
  }
}

// This is our current solution to avoid refetching the card details again. We
// can move away from this in the future and use a loader for this route as well.
export default function PageCardID() {
  const navigate = useNavigate()
  const { features } = useRouteLoaderData('root') as SerializeFrom<
    typeof rootLoader
  >
  const { cards } = useGateHubStore((state) => state.card)
  const { cardId } = useParams<RouteParams['/cards/:cardId']>()

  // goto root in case cards are not enabled
  if (!features.manageWalletCardsEnabled) {
    navigate('/')
  }

  if (!cardId) {
    return <>Card not found</>
  }

  const card = useMemo(
    () => cards?.find((c) => c.id === cardId),
    [cards, cardId]
  )

  if (!card) {
    return <>Card not found</>
  }

  return <CardView card={card} />
}
