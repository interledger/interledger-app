import type { LoaderArgs } from '@remix-run/node'
import { json } from '@remix-run/node'
import { useLoaderData, useParams } from '@remix-run/react'
import { route } from 'routes-gen'
import type { ApplicationProps } from '~/components'
import {
  Card,
  CardContent,
  CardHeader,
  CardLink,
  CardTitle,
  Chip,
  ChipColor,
  Icon,
  Layouts
} from '~/components'
import { Label } from '~/components/Label'
import {
  StatusError,
  grpcClient,
  httpMapping,
  isGrpcError
} from '~/lib/proto.server'

export async function loader({ request, params }: LoaderArgs) {
  const cardDetails = await grpcClient
    .getCardDetails(
      { id: params.accountId as string },
      { meta: { cookies: String(request.headers.get('cookie')) || '' } }
    )
    .then((v) => v)
    .catch(StatusError)
  console.log(cardDetails)

  if (isGrpcError(cardDetails)) {
    throw json({}, httpMapping(cardDetails.code))
  }

  // const card: CardDetails = {
  //   id: params.accountId as string,
  //   network: 'Visa',
  //   bin: '4243',
  //   last4: '4265',
  //   type: 'Debit card',
  //   expiration: '2024/18',
  //   nickname: 'Bobs uncle',
  //   state: 'Verified',
  //   canSend: false,
  //   canReceive: true
  // }

  return json({
    card: cardDetails.response
  })
}

export const handle: ApplicationProps = {
  layout: Layouts.Wallet,
  scaffold: {
    header: {
      back: route('/accounts'),
      actions: (match) => {
        console.log(match)
        const state = match.data.card.state
        if (state == 'Verified') {
          const canSend = match.data.card.canSend
          const canReceive = match.data.card.canReceive
          if (canSend && canReceive) {
            return [
              <Chip key='send' color={ChipColor.indigo}>
                Send
              </Chip>,
              <Chip key='receive' color={ChipColor.purple}>
                Receive
              </Chip>
            ]
          } else if (!canSend && canReceive) {
            return <Chip color={ChipColor.purple}>Receive only</Chip>
          } else if (canSend && !canReceive) {
            return <Chip color={ChipColor.indigo}>Send only</Chip>
          } else return null
        } else if (state == 'OwnershipReviewRequired') {
          return <Chip color={ChipColor.orange}>In review</Chip>
        } else if (state == 'Rejected') {
          return <Chip color={ChipColor.red}>Rejected</Chip>
        } else return null
      }
    },
    isNested: true
  }
}

export default function Page() {
  const { card } = useLoaderData<typeof loader>()
  const params = useParams()

  return (
    <>
      <Card>
        <CardHeader>
          <CardTitle>Card details</CardTitle>
        </CardHeader>
        <CardContent>
          <div className='flex w-full flex-col justify-between space-y-1'>
            <span className='text-weak'>Card number</span>
            <span className='text-medium'>
              {card?.bin.replace(' ', '').slice(0, 4)}{' '}
              {card?.bin.replace(' ', '').slice(4).padEnd(4, '*')} ****{' '}
              {card?.last4}
            </span>
          </div>
          <div className='mt-4 flex w-full flex-col space-y-1'>
            <span className='text-weak'>Card type</span>
            <span className='text-medium'>{card.type}</span>
          </div>
          <div className='mt-4 flex w-full flex-col space-y-1'>
            <span className='text-weak'>Expiration date</span>
            <span className='text-medium'>{card.expiration}</span>
          </div>
        </CardContent>
      </Card>
      <Card>
        <Label>Nickname</Label>
        <CardLink
          className='flex items-center justify-between'
          to={route('/accounts/:accountId/name', {
            accountId: params.accountId as string
          })}
        >
          {card?.nickname}
          <Icon>navigate_next</Icon>
        </CardLink>
      </Card>
    </>
  )
}
