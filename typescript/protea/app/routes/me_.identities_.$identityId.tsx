import type { LoaderArgs, MetaFunction } from '@remix-run/node'
import { json } from '@remix-run/node'
import { Form, useLoaderData } from '@remix-run/react'
import { DateTime } from 'luxon'
import { route } from 'routes-gen'
import type { ApplicationProps } from '~/components'
import {
  AnchorRouter,
  Button,
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  Chip,
  ChipColor,
  Layouts,
  Router
} from '~/components'
import { hasUserSession } from '~/lib/kratos.server'
import { getPublicIdentity, getPublicWalletDetails } from '~/lib/wallet.server'

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
      walletUrlWithoutProtocol: removeProtocol(wallet.publicName),
      verifiedAt: DateTime.fromSeconds(
        parseInt(identity.verifiedAt?.seconds ?? '')
      ).toFormat('dd MMM yyyy')
    },
    isUser
  })
}

export const meta: MetaFunction<typeof loader> = ({ data }) => {
  const metaContent = (() => {
    switch (data.identity.platform) {
      case 'twitter':
        return {
          title: `@${data.identity.identifier} has verified they are a real person`,
          description:
            'Fynbos has verified that this person is real and this is the public proof of their Twitter identity.'
        }
      case 'domain':
        return {
          title: `${data.identity.identifier} is connected to a real person`,
          description:
            'Fynbos has verified that this domain is connected to a real person and this is the public proof of their domain identity.'
        }
      default:
        return {}
    }
  })()

  return {
    title: metaContent.title,
    description: metaContent.description,
    'og:title': metaContent.title,
    'og:url': 'https://fynbos.app/me/identities/' + data.identity.signatureHash,
    'og:description': metaContent.description,
    'og:image': `https://cdn.fynbos.app/identities/${data.identity.signatureHash}/${data.identity.platform}-og.png`,
    'twitter:url':
      'https://fynbos.app/me/identities/' + data.identity.signatureHash,
    'twitter:image': `https://cdn.fynbos.app/identities/${data.identity.signatureHash}/${data.identity.platform}-og.png`,
    'twitter:title': metaContent.title,
    'twitter:description': metaContent.description
  }
}

export const handle: ApplicationProps = {
  layout: Layouts.Focus,
  scaffold: {
    header: {
      back: (match) => `/me/${match.data.identity.walletUrlWithoutProtocol}`,
      title: (match) => {
        switch (match.data.identity.platform) {
          case 'twitter':
            return `@${match.data.identity.identifier}`
          case 'domain':
            return match.data.identity.identifier
        }
      },
      actions: (match) =>
        match.data.identity.state == 'verified' ? (
          <Chip color={ChipColor.green}>Verified</Chip>
        ) : null
    }
  }
}

export default function Page() {
  const { identity, isUser } = useLoaderData<typeof loader>()

  return (
    <>
      {identity.platform == 'twitter' && <Twitter />}
      {identity.platform == 'domain' && <Domain />}
      {!isUser && (
        <Card>
          <CardHeader>
            <CardTitle>What is Fynbos?</CardTitle>
          </CardHeader>
          <CardContent className='flex flex-col space-y-4'>
            <p className='text-medium'>
              Fynbos is a digital wallet for verifying identities, paying
              contacts, and building trust.
            </p>
            <Router
              className='text-sm font-medium text-primary'
              to={route('/signup')}
            >
              Get your own identity card
            </Router>
          </CardContent>
        </Card>
      )}
      <Form
        id='me'
        action={`/me/${identity.walletUrlWithoutProtocol}`}
        method='post'
        className='hidden'
      />
      <input
        form='me'
        value={'paymentPointer'}
        name='paymentPointer'
        type='hidden'
      />
    </>
  )
}

function Twitter() {
  const { identity, isUser } = useLoaderData<typeof loader>()

  return (
    <>
      {identity.state == 'verified' && (
        <Card>
          <CardHeader>
            <CardTitle>Twitter verification</CardTitle>
          </CardHeader>
          <CardContent>
            <p>
              <AnchorRouter to={identity.wallet} className='text-primary'>
                {' '}
                {identity.walletUrlWithoutProtocol}{' '}
              </AnchorRouter>
              has linked their Fynbos wallet to Twitter.
            </p>
            <p className='mt-4'>
              This identity card shows that
              <AnchorRouter
                to={`https://twitter.com/${identity.identifier}`}
                className='text-primary'
              >
                {' '}
                @{identity.identifier}{' '}
              </AnchorRouter>
              is the same person as
              <AnchorRouter to={identity.wallet} className='text-primary'>
                {' '}
                {identity.walletUrlWithoutProtocol}{' '}
              </AnchorRouter>
              who Fynbos have verified is a real person.
            </p>
            <img
              className='mt-4 max-w-[310px]'
              loading='lazy'
              alt='Identity card'
              src={`https://cdn.fynbos.app/identities/${identity.signatureHash}/twitter.png`}
            />
            <div className='mt-4 flex w-full flex-col space-y-1'>
              <span className='text-medium'>Verification date</span>
              <span className='font-medium'>{identity.verifiedAt}</span>
            </div>
            <div className='mt-4 flex w-full flex-col space-y-1'>
              <span className='text-medium'>Public proof</span>
              <AnchorRouter
                to={identity.proof}
                className='break-all font-medium text-primary'
              >
                {identity.proof}
              </AnchorRouter>
            </div>
          </CardContent>
        </Card>
      )}
      {isUser && (
        <Button form='me' type='submit'>
          Pay @{identity.identifier}
        </Button>
      )}
    </>
  )
}

function Domain() {
  const { identity, isUser } = useLoaderData<typeof loader>()

  return (
    <>
      {identity.state == 'verified' && (
        <Card>
          <CardHeader>
            <CardTitle>Domain verification</CardTitle>
          </CardHeader>
          <CardContent>
            <p>
              <AnchorRouter to={identity.wallet} className='text-primary'>
                {' '}
                {identity.walletUrlWithoutProtocol}{' '}
              </AnchorRouter>
              has linked their domain to their Fynbos wallet.
            </p>
            <p className='mt-4'>
              This identity card shows that
              <AnchorRouter to={identity.identifier} className='text-primary'>
                {' '}
                {identity.identifier}{' '}
              </AnchorRouter>
              is connected to
              <AnchorRouter to={identity.wallet} className='text-primary'>
                {' '}
                {identity.walletUrlWithoutProtocol}{' '}
              </AnchorRouter>
              who Fynbos have verified is a real person.
              <img
                className='mt-4 max-w-[310px]'
                loading='lazy'
                alt='Identity card'
                src={`https://cdn.fynbos.app/identities/${identity.signatureHash}/domain.png`}
              />
            </p>
            <div className='mt-4 flex w-full flex-col space-y-1'>
              <span className='text-medium'>Hostname</span>
              <span className='font-medium'>_fynbos.{identity.identifier}</span>
            </div>
            <div className='mt-4 flex w-full flex-col space-y-1'>
              <span className='text-medium'>Code</span>
              <span className='font-medium'>{identity.signatureHash}</span>
            </div>
          </CardContent>
        </Card>
      )}
      {isUser && (
        <Button form='me' type='submit'>
          Pay {identity.identifier}
        </Button>
      )}
    </>
  )
}

function removeProtocol(url: string): string {
  return url.replace(/(http(s)?:\/\/)/i, '')
}
