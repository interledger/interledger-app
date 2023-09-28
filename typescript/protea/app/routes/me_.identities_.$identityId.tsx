import type { LoaderFunctionArgs, MetaFunction } from '@remix-run/node'
import { json } from '@remix-run/node'
import type { UIMatch } from '@remix-run/react'
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
import { getIdentityBySignatureHash } from '~/data/identity.server'
import { getPublicWalletDetails } from '~/data/wallet.server'
import { hasUserSession } from '~/lib/kratos.server'
import { mergeMeta } from '~/lib/meta'

export async function loader({ request, params }: LoaderFunctionArgs) {
  const identity = await getIdentityBySignatureHash(
    request,
    params.identityId as string
  )
  const wallet = await getPublicWalletDetails(request, identity.walletId)

  const isUser = hasUserSession(request)
  return json({
    wallet: {
      publicName: wallet.publicName
    },
    identity: {
      ...identity,
      walletUrlWithoutProtocol: removeProtocol(wallet.publicName),
      verifiedAt: DateTime.fromSeconds(
        Number(identity.verifiedAt?.seconds ?? '')
      ).toFormat('dd MMM yyyy')
    },
    isUser
  })
}

export const meta: MetaFunction<typeof loader> = mergeMeta(({ data }) => {
  let metaContent

  switch (data?.identity.platform) {
    case 'twitter':
      metaContent = {
        title: `@${data.identity.identifier} has verified they are a real person`,
        description:
          'Fynbos has verified that this person is real and this is the public proof of their Twitter identity.'
      }
      break
    case 'domain':
      metaContent = {
        title: `${data.identity.identifier} is connected to a real person`,
        description:
          'Fynbos has verified that this domain is connected to a real person and this is the public proof of their domain identity.'
      }
      break
    case 'discord':
      metaContent = {
        title: `@${data.identity.identifier} has verified they are a real person`,
        description:
          'Fynbos has verified that this person is real and this is the public proof of their Discord identity.'
      }
      break
    case 'slack':
      metaContent = {
        title: `@${data.identity.identifier} has verified they are a real person`,
        description:
          'Fynbos has verified that this person is real and this is the public proof of their Slack identity.'
      }
      break
    default:
      metaContent = {
        title: `@${data?.identity.identifier} has verified they are a real person`,
        description:
          'Fynbos has verified that this person is real and this is the public proof of their identity.'
      }
  }

  return [
    { title: metaContent.title },
    {
      property: 'og:title',
      content: metaContent.title
    },
    {
      name: 'twitter:title',
      content: metaContent.title
    },
    {
      name: 'description',
      content: metaContent.description
    },
    {
      property: 'og:description',
      content: metaContent.description
    },
    {
      name: 'twitter:description',
      content: metaContent.description
    },
    {
      property: 'og:image',
      content: `https://cdn.fynbos.app/identities/${data?.identity.signatureHash}/${data?.identity.platform}-og.png`
    },
    {
      name: 'twitter:image',
      content: `https://cdn.fynbos.app/identities/${data?.identity.signatureHash}/${data?.identity.platform}-og.png`
    },
    {
      name: 'og:url',
      content:
        'https://fynbos.app/me/identities/' + data?.identity.signatureHash
    },
    {
      name: 'twitter:url',
      content:
        'https://fynbos.app/me/identities/' + data?.identity.signatureHash
    }
  ]
})

export const handle: ApplicationProps = {
  layout: Layouts.Focus,
  scaffold: {
    header: {
      back: (match: UIMatch<typeof loader>) =>
        `/me/${match.data.identity.walletUrlWithoutProtocol}`,
      title: (match: UIMatch<typeof loader>) => {
        switch (match.data.identity.platform) {
          case 'twitter':
            return `@${match.data.identity.identifier}`
          case 'domain':
            return match.data.identity.identifier
          default:
            return match.data.identity.identifier
        }
      },
      actions: (match: UIMatch<typeof loader>) =>
        match.data.identity.state == 'verified'
          ? {
              key: 'Verified',
              nodes: <Chip color={ChipColor.green}>Verified</Chip>
            }
          : null
    }
  }
}

export default function Page() {
  const { identity, isUser } = useLoaderData<typeof loader>()

  return (
    <>
      {identity.platform == 'twitter' && <Twitter />}
      {identity.platform == 'discord' && <Discord />}
      {identity.platform == 'slack' && <Slack />}
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

function Discord() {
  const { identity, isUser } = useLoaderData<typeof loader>()

  return (
    <>
      {identity.state == 'verified' && (
        <Card>
          <CardHeader>
            <CardTitle>Discord verification</CardTitle>
          </CardHeader>
          <CardContent>
            <p>
              <AnchorRouter to={identity.wallet} className='text-primary'>
                {' '}
                {identity.walletUrlWithoutProtocol}{' '}
              </AnchorRouter>
              has linked their Fynbos wallet to Discord.
            </p>
            <p className='mt-4'>
              This identity card shows that {identity.identifier} is the same
              person as
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
              src={`https://cdn.fynbos.app/identities/${identity.signatureHash}/discord.png`}
            />
            <div className='mt-4 flex w-full flex-col space-y-1'>
              <span className='text-medium'>Verification date</span>
              <span className='font-medium'>{identity.verifiedAt}</span>
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

function Slack() {
  const { identity, isUser } = useLoaderData<typeof loader>()

  return (
    <>
      {identity.state == 'verified' && (
        <Card>
          <CardHeader>
            <CardTitle>Slack verification</CardTitle>
          </CardHeader>
          <CardContent>
            <p>
              <AnchorRouter to={identity.wallet} className='text-primary'>
                {' '}
                {identity.walletUrlWithoutProtocol}{' '}
              </AnchorRouter>
              has linked their Fynbos wallet to Slack.
            </p>
            <p className='mt-4'>
              This identity card shows that {identity.identifier} is the same
              person as
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
              src={`https://cdn.fynbos.app/identities/${identity.signatureHash}/slack.png`}
            />
            <div className='mt-4 flex w-full flex-col space-y-1'>
              <span className='text-medium'>Verification date</span>
              <span className='font-medium'>{identity.verifiedAt}</span>
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
              <AnchorRouter
                to={`https://${identity.identifier}`}
                className='text-primary'
              >
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
              <span className='break-all font-medium'>
                {identity.signatureHash}
              </span>
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
