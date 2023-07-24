import type { LoaderArgs } from '@remix-run/node'
import { json } from '@remix-run/node'
import { Form, useLoaderData, useParams } from '@remix-run/react'
import { route } from 'routes-gen'
import type { ApplicationProps } from '~/components'
import {
  Card,
  CardContent,
  CardHeader,
  CardLink,
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

  return json({
    card: cardDetails.response
  })
}

export const handle: ApplicationProps = {
  layout: Layouts.Wallet,
  scaffold: {
    header: {
      back: route('/accounts'),
      title: (match) => match.data.name
    },
    isNested: true
  }
}

export default function Page() {
  const { card } = useLoaderData<typeof loader>()
  const params = useParams()

  return (
    <>
      <Form
        id='edit-linked-account-name'
        action={`/accounts/${params.accountId}`}
        method='post'
        className='hidden'
      />
      <Card>
        <CardHeader>Card details</CardHeader>
        <CardContent>
          <div className='flex w-full flex-col justify-between space-y-1'>
            <span className='text-weak'>Card number</span>
            <span className='text-medium'>
              {card?.bin}{' '}
              {Array(4 - (card?.bin.length % 4))
                .fill('*')
                .join('')}{' '}
              **** {card?.last4}
            </span>
          </div>
          <div className='mt-4 flex w-full flex-col space-y-1'>
            <span className='text-weak'>Card type</span>
            <span className='text-medium'>{card.type}</span>
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
