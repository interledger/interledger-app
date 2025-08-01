// Placeholder page

import { type SerializeFrom } from '@remix-run/node'
import { useNavigate, useOutletContext, useParams, useRouteLoaderData } from '@remix-run/react'
import { type RouteParams } from 'routes-gen'
import { Layouts, type ApplicationProps } from '~/components'
import type { loader } from './cards'
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
  const cards = useOutletContext<SerializeFrom<typeof loader>['cards']>()
  const { cardId } = useParams<RouteParams['/cards/:cardId']>()

  // goto root in case cards are not enabled
  if (!features.manageWalletCardsEnabled) {
    navigate("/")
  }

  if (!cardId) {
    return <>TODO: NotFound</>
  }

  const card = cards.find((c) => c.id === cardId)

  if (!card) {
    return <>TODO: NotFound</>
  }

  return <div>Placeholder</div>
}
