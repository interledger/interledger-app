import { type SerializeFrom } from '@remix-run/node'
import { useNavigate, useParams, useRouteLoaderData } from '@remix-run/react'
import { useEffect, useMemo } from 'react'
import { route, type RouteParams } from 'routes-gen'
import { Layouts, type ApplicationProps } from '~/components'
import { CardView } from '~/components/Cards'
import { useCardsStore } from '~/lib/cards/useCardsStore'
import type { loader as rootLoader } from '~/root'

export const handle: ApplicationProps = {
  layout: Layouts.Wallet,
  scaffold: {
    header: {
      title: 'Cards',
      back: route('/cards')
    },
    isNested: true
  }
}

export default function PageCardID() {
  const navigate = useNavigate()
  const { features } = useRouteLoaderData('root') as SerializeFrom<
    typeof rootLoader
  >
  const { cards, areCardsFetched } = useCardsStore()
  const { cardId } = useParams<RouteParams['/cards/:cardId']>()
  const card = useMemo(
    () => (cardId ? cards?.[cardId] : undefined),
    [cards, cardId]
  )

  useEffect(() => {
    if (areCardsFetched && (!card || !cardId)) {
      navigate('/cards')
    }
  }, [areCardsFetched, card, cardId, navigate])

  useEffect(() => {
    if (!features.manageWalletCardsEnabled) {
      navigate('/')
    }
  }, [features.manageWalletCardsEnabled, navigate])

  if (!card) return null

  return <CardView card={card} />
}
