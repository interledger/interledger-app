import type { LoaderArgs, MetaFunction } from '@remix-run/node'
import { json } from '@remix-run/node'
import { AnchorRouter, Card, Layouts } from '~/components'
import { getPublicIdentity } from '~/lib/wallet.server'
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

export const meta: MetaFunction<typeof loader> = ({ data }) => {
  const metaContent = {
    title: `@${data.identity.identifier} has verified they are a real person`,
    description:
      'Fynbos has verified that this person is real and this is the public proof of their twitter identity.'
  }

  return {
    title: metaContent.title,
    description: metaContent.description,
    'og:title': metaContent.title,
    'og:url': 'https://fynbos.app/me/identities/' + data.identity.signatureHash,
    'og:description': metaContent.description,
    'og:image': `https://cdn.fynbos.app/identities/${data.identity.signatureHash}/twitter-og.png`,
    'twitter:url':
      'https://fynbos.app/me/identities/' + data.identity.signatureHash,
    'twitter:image': `https://cdn.fynbos.app/identities/${data.identity.signatureHash}/twitter-og.png`,
    'twitter:title': metaContent.title,
    'twitter:description': metaContent.description
  }
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
          src={`https://cdn.fynbos.app/identities/${identity.signatureHash}/twitter.png`}
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
