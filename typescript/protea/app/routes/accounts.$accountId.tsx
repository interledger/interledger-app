import type { LoaderArgs } from '@remix-run/node'
import { json } from '@remix-run/node'
import { useLoaderData, useParams } from '@remix-run/react'
import { useState } from 'react'
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
  Dialog,
  Icon,
  Layouts,
  Router,
  TextButton
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

  if (isGrpcError(cardDetails)) {
    throw json({}, httpMapping(cardDetails.code))
  }

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
  const [limitationsDialog, setLimitationsDialog] = useState<boolean>(false)

  return (
    <>
      {card.state == 'OwnershipReviewRequired' && (
        <Card>
          <CardContent>
            <p>
              This card is currently in review. You will be unable transact
              until verified. We will email you once the verification is
              complete.
            </p>
          </CardContent>
        </Card>
      )}
      {card.state == 'Rejected' && (
        <Card>
          <CardContent className='flex flex-col gap-y-4'>
            <p>
              This card cannot process payments because of an address mismatch.
            </p>
            <Router className='text-primary' to={route('/support')}>
              Contact support
            </Router>
          </CardContent>
        </Card>
      )}
      {card.state == 'Verified' && card.canSend && !card.canReceive && (
        <Card>
          <CardContent>
            <p>
              This card is enabled to send payments, but unable to receive
              payments.
            </p>
            <TextButton
              className='mt-4 text-primary'
              onClick={() => setLimitationsDialog(true)}
            >
              More information
            </TextButton>
          </CardContent>
        </Card>
      )}
      {card.state == 'Verified' && !card.canSend && card.canReceive && (
        <Card>
          <CardContent>
            <p>
              This card is enabled to receive payments, but unable to make
              payments.
            </p>
            <TextButton
              className='mt-4 text-primary'
              onClick={() => setLimitationsDialog(true)}
            >
              More information
            </TextButton>
          </CardContent>
        </Card>
      )}
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
      <Dialog open={limitationsDialog} setOpen={setLimitationsDialog}>
        <CardHeader>
          <h1 className='text-xl font-medium'>Card limitations</h1>
        </CardHeader>
        <CardContent>
          <p className='text-medium'>
            Limitations are placed on some cards enabling sending or receiving
            money only. Please contact your card issuer to remove limitations,
            or connect another card.
          </p>

          <div className='flex w-full justify-end space-x-6 pt-4'>
            <TextButton
              type='button'
              onClick={() => setLimitationsDialog(false)}
            >
              Close
            </TextButton>
          </div>
        </CardContent>
      </Dialog>
    </>
  )
}
