import { json, redirect, type LoaderFunctionArgs } from '@remix-run/node'
import { useLoaderData, type UIMatch } from '@remix-run/react'
import { route } from 'routes-gen'
import {
  CardView,
  Chip,
  ChipColor,
  Layouts,
  type ApplicationProps
} from '~/components'
import { MasterCardLogo } from '~/components/CardView/MasterCardLogo'
import { getCardDetailsWithTx, getFeatures } from '~/data/wallet.server'
import { CardType } from '~/generated/connect/backend/v1/backend_pb'

export const handle: ApplicationProps = {
  layout: Layouts.Wallet,
  scaffold: {
    header: {
      title: 'Cards',
      back: route('/cards'),
      actions: (match: UIMatch<typeof loader>) => {
        if (match.data.card?.type === CardType.VIRTUAL) {
          return {
            key: 'Virtual',
            nodes: (
              <>
                <Chip color={ChipColor.purple}>Virtual</Chip>
                <Chip color={ChipColor.blue}>
                  <MasterCardLogo size='xs' />
                  {match.data.card.maskedPan}
                </Chip>
              </>
            )
          }
        } else if (match.data.card?.type === CardType.PHYSICAL) {
          return {
            key: 'Physical',
            nodes: (
              <>
                <Chip color={ChipColor.indigo}>Physical</Chip>
                <Chip color={ChipColor.blue}>
                  <MasterCardLogo size='xs' />
                  {match.data.card.maskedPan}
                </Chip>
              </>
            )
          }
        }

        return null
      }
    },
    isNested: true
  }
}

export async function loader({ request, params }: LoaderFunctionArgs) {
  const features = await getFeatures(request)
  if (!features.manageWalletCardsEnabled) {
    throw redirect('/')
  }

  const cardId = params.cardId

  if (!cardId) {
    throw redirect('/')
  }

  const response = await getCardDetailsWithTx(request, { id: cardId })

  return json(response)
}

export default function PageCardID() {
  const { card, transactions } = useLoaderData<typeof loader>()

  return <CardView card={card!} transactions={transactions} />
}
