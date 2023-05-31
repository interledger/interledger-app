import type { LoaderArgs, MetaFunction } from '@remix-run/node'
import { json } from '@remix-run/node'
import { AnchorRouter, Button, Card, Layouts, Router } from '~/components'
import { getPublicIdentity, getPublicWalletDetails } from '~/lib/wallet.server'
import { Form, useLoaderData } from '@remix-run/react'
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
      identifierWithPrefix: '@' + identity.identifier,
      walletUrlWithoutProtocol: removeProtocol(wallet.publicName),
      verifiedAt: DateTime.fromSeconds(
        parseInt(identity.verifiedAt?.seconds ?? '')
      ).toFormat('dd MMM yyyy')
    },
    isUser
  })
}

export const meta: MetaFunction<typeof loader> = ({ data }) => {
  const metaContent = {
    title: `${data.identity.identifierWithPrefix} has verified they are a real person`,
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
          <p className='mt-4'>
            This identity card shows that
            <AnchorRouter
              to={`https://twitter.com/${identity.identifier}`}
              className='text-primary'
            >
              {' '}
              {identity.identifierWithPrefix}{' '}
            </AnchorRouter>
            is the same person as
            <AnchorRouter to={identity.wallet} className='text-primary'>
              {' '}
              {identity.walletUrlWithoutProtocol}{' '}
            </AnchorRouter>
            who Fynbos have verified is a real person.
          </p>
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
      <Form id='me' action={`/me/${identity.walletUrlWithoutProtocol}`} method='post' className='hidden' />
      <input
        form='me'
        value={'paymentPointer'}
        name='paymentPointer'
        type='hidden'
      />
      {isUser && (
        <Button form='me' type='submit'>
          Pay {identity.identifierWithPrefix}
        </Button>
      )}
      {!isUser && (
        <Card className='space-y-4'>
          <h1 className='font-display text-lg font-medium'>What is Fynbos?</h1>
          <p className='text-sm text-medium'>
            Fynbos is a digital wallet for verifying identities, paying
            contacts, and building trust.
          </p>
          <Router
            className='text-sm font-medium text-primary'
            to={route('/signup')}
          >
            Get your own identity card
          </Router>
        </Card>
      )}
    </>
  )
}

function removeProtocol(url: string): string {
  return url.replace(/(http(s)?:\/\/)/i, '')
}
