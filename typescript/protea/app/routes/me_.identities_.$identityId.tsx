import type { LoaderArgs } from '@remix-run/node'
import { json } from '@remix-run/node'
import {
  AnchorRouter,
  Card,
  Icon,
  Layouts,
  OutlineButton,
  Router,
  Switch
} from '~/components'
import { getPublicIdentity, getPublicWalletDetails } from '~/lib/wallet.server'
import { useLoaderData } from '@remix-run/react'
import { DateTime } from 'luxon'

export async function loader({ request, params }: LoaderArgs) {
  const identity = await getPublicIdentity(request, params.identityId as string)

  if (identity.state !== 'verified')
    throw json({}, { status: 404, statusText: 'Not found' })

  return json({
    identity: {
      ...identity,
      verifiedAt: DateTime.fromSeconds(
        parseInt(identity.verifiedAt?.seconds ?? '')
      ).toFormat('dd MMM yyyy')
    }
  })
}

export const handle = {
  layout: Layouts.FocusLayout
}

export default function Page() {
  // const { editable, wallet, paymentPointer, paymentPointerParam } =
  //   useLoaderData<typeof loader>()
  const { identity } = useLoaderData<typeof loader>()

  return (
    identity.state == 'verified' && (
      <Card>
        <img
          className='max-w-[310px]'
          loading='lazy'
          alt='Identity card'
          src='https://cdn.fynbos.app/identities/template.png'
        />
        <Card.Item variant='col' className='mt-4'>
          <span className='text-medium'>Verification date</span>
          <span className='font-medium'>{identity.verifiedAt}</span>
        </Card.Item>
        <Card.Item variant='col' className='mt-4'>
          <span className='text-medium'>Public proof</span>
          <AnchorRouter
            to={identity.proof}
            className='font-medium text-primary break-all'
          >
            {identity.proof}
          </AnchorRouter>
        </Card.Item>
      </Card>
    )
  )
}
