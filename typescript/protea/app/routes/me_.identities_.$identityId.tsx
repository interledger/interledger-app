import type { LoaderArgs, MetaFunction } from '@remix-run/node'
import { json } from '@remix-run/node'
import { AnchorRouter, Card, FynbosIcon, Layouts, Router } from '~/components'
import { getPublicIdentity, getPublicWalletDetails } from '~/lib/wallet.server'
import { useLoaderData } from '@remix-run/react'
import { DateTime } from 'luxon'
import { route } from 'routes-gen'
import { hasUserSession } from '~/lib/kratos.server'

export async function loader({ request, params }: LoaderArgs) {
  const identity = await getPublicIdentity(request, params.identityId as string)
  const wallet = await getPublicWalletDetails(request, identity.walletId)

  if (identity.state !== 'verified')
    throw json({}, { status: 404, statusText: 'Not found' })

  const isUser = hasUserSession(request)
  return json({
    wallet: {
      publicName: wallet.publicName
    },
    identity: {
      ...identity,
      verifiedAt: DateTime.fromSeconds(
        parseInt(identity.verifiedAt?.seconds ?? '')
      ).toFormat('dd MMM yyyy')
    },
    isUser
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
  const { identity, wallet, isUser } = useLoaderData<typeof loader>()

  return (
    <>
      {identity.state == 'verified' && (
        <Card>
          <p>{wallet.publicName} has linked their Fynbos wallet to Twitter.</p>
          <img
            className='max-w-[310px] mt-4'
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
      )}
      {!isUser && (
        <Card className='space-y-4'>
          <h1 className='font-display text-lg font-medium'>Sign up</h1>
          <div className='flex items-start space-x-4'>
            <div className='flex items-center justify-between rounded-full bg-nav p-5 text-medium'>
              <FynbosIcon />
            </div>
            <div className='flex flex-col space-y-4'>
              <p className='text-sm text-medium'>
                Sign up with Fynbos to reserve your wallet address and start
                transacting.
              </p>
              <Router
                className='text-sm font-medium text-primary'
                to={route('/signup')}
              >
                Sign up now
              </Router>
            </div>
          </div>
        </Card>
      )}
    </>
  )
}
