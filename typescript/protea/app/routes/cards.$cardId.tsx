import { type SerializeFrom } from '@remix-run/node'
import { useNavigate, useParams, useRouteLoaderData } from '@remix-run/react'
import { useMemo } from 'react'
import { type RouteParams } from 'routes-gen'
import { CardView, Layouts, type ApplicationProps } from '~/components'
import { useCardsStore } from '~/lib/cards/hooks/useCardsStore'
import type { loader as rootLoader } from '~/root'

export const handle: ApplicationProps = {
  layout: Layouts.Wallet,
  scaffold: {
    header: {
      title: 'Cards'
    }
  }
}

export default function PageCardID() {
  const navigate = useNavigate()
  const { features } = useRouteLoaderData('root') as SerializeFrom<
    typeof rootLoader
  >
  const { cards } = useCardsStore()
  const { cardId } = useParams<RouteParams['/cards/:cardId']>()
  const card = useMemo(
    () => (cardId ? cards?.[cardId] : undefined),
    [cards, cardId]
  )

  if (!features.manageWalletCardsEnabled) {
    navigate('/')
  }

  if (!card || !cardId) {
    return <>Card not found</>
  }

  return <CardView card={card} />
}
