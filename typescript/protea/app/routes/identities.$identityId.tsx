import type {
  ActionFunctionArgs,
  LoaderFunctionArgs,
  MetaFunction
} from '@remix-run/node'
import { json } from '@remix-run/node'
import type { UIMatch } from '@remix-run/react'
import { Form, useFetcher, useLoaderData } from '@remix-run/react'
import { DateTime } from 'luxon'
import { useCallback, useState } from 'react'
import { route } from 'routes-gen'
import type { ApplicationProps } from '~/components'
import {
  AnchorRouter,
  Button,
  Card,
  CardContent,
  CardCopy,
  CardHeader,
  CardIcon,
  CardLink,
  CardTitle,
  Chip,
  ChipColor,
  Dialog,
  Icon,
  Layouts,
  OutlineButton,
  Router,
  Switch,
  TextButton
} from '~/components'
import { Label } from '~/components/Label'
import { getIdentity } from '~/data/identity.server'
import { getPublicWalletDetails, getWalletInfo } from '~/data/wallet.server'
import { jsonWithCSRF, validateCSRFToken } from '~/lib/csrf.server'
import { isConnectError } from '~/lib/error.server'
import { grpc } from '~/lib/grpc.server'
import { mergeMeta } from '~/lib/meta'
import { getPusherArgs } from '~/lib/pusher.server'
import { jsonWithSnackbar, redirectWithSnackbar } from '~/lib/snackbar.server'
import { usePusher } from '~/lib/usePusher'
import { useScaffoldStore } from '~/lib/useScaffoldStore'

export async function loader({ request, params }: LoaderFunctionArgs) {
  const walletInfo = await getWalletInfo(request)
  const { publicName } = await getPublicWalletDetails(
    request,
    walletInfo.walletID
  )
  const identity = await getIdentity(request, params.identityId as string)
  const pusherArgs = await getPusherArgs(request)

  return jsonWithCSRF(request, {
    walletInfo,
    publicName,
    identity: {
      ...identity,
      verifiedAt: DateTime.fromSeconds(
        Number(identity.verifiedAt?.seconds ?? '')
      ).toFormat('dd MMM yyyy')
    },
    pusherArgs
  })
}

export const handle: ApplicationProps = {
  layout: Layouts.Wallet,
  scaffold: {
    header: {
      back: route('/identities'),
      title: (match: UIMatch<typeof loader>) =>
        `@${match.data.identity.identifier}`,
      actions: (match: UIMatch<typeof loader>) => {
        const state = match.data.identity.state
        switch (state) {
          case 'verified':
            return {
              key: 'Verified',
              nodes: <Chip color={ChipColor.green}>Verified</Chip>
            }
          case 'unverified':
            return {
              key: 'Unverified',
              nodes: <Chip color={ChipColor.yellow}>Unverified</Chip>
            }
          case 'failed':
            return {
              key: 'Failed',
              nodes: <Chip color={ChipColor.red}>Failed</Chip>
            }
          case 'pending':
            return {
              key: 'Pending',
              nodes: <Chip color={ChipColor.orange}>Pending</Chip>
            }
          default:
            return null
        }
      }
    },
    isNested: true
  }
}

export const meta: MetaFunction = mergeMeta(() => [
  {
    title: 'Identities'
  }
])

export default function Page() {
  const { identity, csrfToken, pusherArgs } = useLoaderData<typeof loader>()

  usePusher(pusherArgs, ['identity'])

  return (
    <>
      <Form
        id='identity'
        action={route('/identities/:identityId', { identityId: identity.id })}
        method='post'
        className='hidden'
      />
      <input form='identity' value={csrfToken} name='csrfToken' type='hidden' />
      {identity.platform == 'twitter' && <Twitter />}
      {identity.platform == 'domain' && <Domain />}
      {identity.platform == 'discord' && <Discord />}
      {identity.platform == 'slack' && <Slack />}
    </>
  )
}

function Twitter() {
  const { identity, publicName, walletInfo } = useLoaderData<typeof loader>()

  const [showDialog, setShowDialog] = useState<boolean>(false)

  const fetcher = useFetcher()
  const _onChangeSwitch = useCallback<{
    (formName: string, publish: boolean): void
  }>(
    (formName, publish) => {
      fetcher.submit(
        { formName, publish: publish.toString() },
        { method: 'post' }
      )
    },
    [fetcher]
  )

  return (
    <>
      {identity.state == 'verified' && (
        <>
          <Card>
            <CardHeader>
              <CardTitle>Twitter details</CardTitle>
            </CardHeader>
            <CardContent>
              <img
                className='max-w-[310px]'
                loading='lazy'
                alt='Identity card'
                src={`https://cdn.fynbos.app/identities/${identity.signatureHash}/twitter.png`}
              />
              <div className='mt-4 flex w-full flex-col space-y-1'>
                <span className='text-weak'>Verification date</span>
                <span className='font-medium'>{identity.verifiedAt}</span>
              </div>
              <div className='mt-4 flex w-full flex-col space-y-1'>
                <span className='text-weak'>Public proof</span>
                <AnchorRouter
                  to={identity.proof}
                  className='break-all font-medium text-primary'
                >
                  {identity.proof}
                </AnchorRouter>
              </div>
            </CardContent>
          </Card>
          <Card>
            <CardContent>
              <div className='flex items-start space-x-4'>
                <CardIcon className='!bg-error'>
                  <Icon className='text-error'>exclamation</Icon>
                </CardIcon>
                <div className='flex flex-col space-y-1'>
                  <h3 className='font-medium text-medium'>Please note</h3>
                  <p className='text-sm text-medium'>
                    Do not delete the public proof from your Twitter timeline.
                    Doing so will result in your identity no longer being
                    verified by Fynbos.
                  </p>
                </div>
              </div>
            </CardContent>
          </Card>
          <Card>
            <Label className='mt-2'>Public profile</Label>
            <CardLink
              to={`/me/${walletInfo.formattedURL}`}
              className='items-center justify-between bg-nav'
            >
              <div className='flex space-x-3'>
                <Icon>contact_page</Icon>
                <span>{publicName}</span>
              </div>
              <Icon>navigate_next</Icon>
            </CardLink>
            <CardContent>
              <div className='flex items-center justify-between'>
                <span className='text-sm'>
                  Show Twitter handle on your Fynbos public profile
                </span>
                <Switch
                  checked={identity.public}
                  disabled={false}
                  onChange={() => _onChangeSwitch('publish', !identity.public)}
                />
              </div>
            </CardContent>
          </Card>
          <OutlineButton
            className='!text-error outline-error hover:!text-red-800 hover:outline-red-800 focus-visible:outline-red-800'
            type='button'
            onClick={() => setShowDialog(true)}
          >
            Delete
          </OutlineButton>
        </>
      )}
      {identity.state === 'unverified' && (
        <>
          <Card>
            <CardContent>
              <p>
                To prove that you own this Twitter handle (and that you are a
                real person) we're going to post this Twitter identity card onto
                your timeline and link it back to your wallet.
              </p>
              <img
                className='mt-4 max-w-[310px]'
                loading='lazy'
                alt='Identity card'
                src={`https://cdn.fynbos.app/identities/${identity.signatureHash}/twitter.png`}
              />
              <Router
                className='mt-4 rounded text-sm font-medium text-primary focus-visible:outline focus-visible:outline-2 focus-visible:outline-focus'
                to={route('/connect/twitter')}
              >
                Read more about how we verify your Twitter handle.
              </Router>
            </CardContent>
          </Card>
          <div className='flex w-full space-x-2'>
            <OutlineButton
              shrink
              className='!text-error outline-error hover:!text-red-800 hover:outline-red-800 focus-visible:outline-red-800'
              type='button'
              onClick={() => setShowDialog(true)}
            >
              Delete
            </OutlineButton>
            <Button
              name='formName'
              value='verify'
              form='identity'
              type='submit'
            >
              Send Tweet
            </Button>
          </div>
        </>
      )}
      {identity.state == 'failed' && (
        <>
          <Card>
            <CardContent>
              <p>
                Your Twitter identity verification has failed. Please try again.
              </p>
              <img
                className='mt-4 max-w-[310px]'
                loading='lazy'
                alt='Identity card'
                src={`https://cdn.fynbos.app/identities/${identity.signatureHash}/twitter.png`}
              />
            </CardContent>
          </Card>
          <div className='flex w-full space-x-2'>
            <OutlineButton
              shrink
              className='!text-error outline-error hover:!text-red-800 hover:outline-red-800 focus-visible:outline-red-800'
              type='button'
              onClick={() => setShowDialog(true)}
            >
              Delete
            </OutlineButton>
            <Button name='formName' value='retry' form='identity' type='submit'>
              Retry
            </Button>
          </div>
        </>
      )}
      {identity.state == 'pending' && (
        <>
          <Card>
            <CardContent>
              <p>
                Your Twitter identity verification is pending. We will notify
                you once verified.
              </p>
              <img
                className='mt-4 max-w-[310px]'
                loading='lazy'
                alt='Identity card'
                src={`https://cdn.fynbos.app/identities/${identity.signatureHash}/twitter.png`}
              />
            </CardContent>
          </Card>
          <OutlineButton
            className='!text-error outline-error hover:!text-red-800 hover:outline-red-800 focus-visible:outline-red-800'
            type='button'
            onClick={() => setShowDialog(true)}
          >
            Delete
          </OutlineButton>
        </>
      )}
      <Dialog open={showDialog} setOpen={setShowDialog}>
        <CardHeader>
          <h1 className='text-xl font-medium'>Remove Twitter ID card</h1>
        </CardHeader>
        <CardContent>
          <span className='text-medium'>
            Are you sure you want to remove the Twitter identity card? This
            action cannot be undone.
          </span>

          <div className='flex w-full justify-end space-x-6 pt-2'>
            <TextButton
              type='button'
              className='!text-medium'
              onClick={() => setShowDialog(false)}
            >
              Cancel
            </TextButton>
            <TextButton
              name='formName'
              className='!text-error'
              value='delete'
              form='identity'
              type='submit'
            >
              Remove card
            </TextButton>
          </div>
        </CardContent>
      </Dialog>
    </>
  )
}

function Domain() {
  const { identity, publicName, walletInfo } = useLoaderData<typeof loader>()

  const hostName =
    identity.txtRecord?.substring(0, identity.txtRecord.indexOf('=')) || ''
  const code =
    identity.txtRecord?.substring(identity.txtRecord.indexOf('=') + 1) || ''
  const [showDialog, setShowDialog] = useState<boolean>(false)

  const [pushSnackbar] = useScaffoldStore((state) => [state.pushSnackbar])

  const fetcher = useFetcher()
  const _onChangeSwitch = useCallback<{
    (formName: string, publish: boolean): void
  }>(
    (formName, publish) => {
      fetcher.submit(
        { formName, publish: publish.toString() },
        { method: 'post' }
      )
    },
    [fetcher]
  )

  return (
    <>
      {identity.state == 'verified' && (
        <>
          <Card>
            <CardHeader>
              <CardTitle>Domain details</CardTitle>
            </CardHeader>
            <CardContent>
              <img
                className='max-w-[310px]'
                loading='lazy'
                alt='Identity card'
                src={`https://cdn.fynbos.app/identities/${identity.signatureHash}/domain.png`}
              />
              <div className='mt-4 flex w-full flex-col space-y-1'>
                <span className='text-weak'>Hostname</span>
                <span className='font-medium'>{hostName}</span>
              </div>
              <div className='mt-4 flex w-full flex-col space-y-1'>
                <span className='text-weak'>Code</span>
                <span className='font-medium'>{code}</span>
              </div>
            </CardContent>
          </Card>
          <Card>
            <CardContent>
              <div className='flex items-start space-x-4'>
                <CardIcon className='!bg-error'>
                  <Icon className='text-error'>exclamation</Icon>
                </CardIcon>
                <div className='flex flex-col space-y-1'>
                  <h3 className='font-medium text-medium'>Please note</h3>
                  <p className='text-sm text-medium'>
                    Do not delete the TXT record at your DNS provider. Doing so
                    will result in your identity no longer being verified by
                    Fynbos.
                  </p>
                </div>
              </div>
            </CardContent>
          </Card>
          <Card>
            <Label className='mt-2'>Public profile</Label>
            <CardLink
              to={`/me/${walletInfo.formattedURL}`}
              className='items-center justify-between bg-nav'
            >
              <div className='flex space-x-3'>
                <Icon>contact_page</Icon>
                <span>{publicName}</span>
              </div>
              <Icon>navigate_next</Icon>
            </CardLink>
            <CardContent>
              <div className='flex items-center justify-between'>
                <span className='text-sm'>
                  Show domain on your Fynbos public profile
                </span>
                <Switch
                  checked={identity.public}
                  disabled={false}
                  onChange={() => _onChangeSwitch('publish', !identity.public)}
                />
              </div>
            </CardContent>
          </Card>
          <OutlineButton
            className='!text-error outline-error hover:!text-red-800 hover:outline-red-800 focus-visible:outline-red-800'
            type='button'
            onClick={() => setShowDialog(true)}
          >
            Delete
          </OutlineButton>
        </>
      )}
      {identity.state === 'unverified' && (
        <>
          <Card>
            <CardHeader>
              <CardTitle>Domain details</CardTitle>
            </CardHeader>
            <CardContent>
              <p>
                To prove ownership of the domain, please create a TXT record in
                your DNS configuration using the following details:
              </p>
              <p>
                It may take up to 72 hours to propagate, we will notify you once
                complete.
              </p>
              <img
                className='mt-4 max-w-[310px]'
                loading='lazy'
                alt='Identity card'
                src={`https://cdn.fynbos.app/identities/${identity.signatureHash}/domain.png`}
              />
            </CardContent>
            <Label className='mt-2'>Hostname</Label>
            <CardCopy
              content={hostName}
              success='Hostname copied to clipboard.'
              copyError="Couldn't copy to clipboard."
              shareError="Couldn't share hostname."
            >
              {hostName}
            </CardCopy>
            <Label className='mt-2'>Code</Label>
            <CardCopy
              content={code}
              success='Code copied to clipboard.'
              copyError="Couldn't copy to clipboard."
              shareError="Couldn't share code."
            >
              {code}
            </CardCopy>
          </Card>
          <div className='flex w-full space-x-2'>
            <OutlineButton
              shrink
              className='!text-error outline-error hover:!text-red-800 hover:outline-red-800 focus-visible:outline-red-800'
              type='button'
              onClick={() => setShowDialog(true)}
            >
              Delete
            </OutlineButton>
            <Button
              name='formName'
              value='verify'
              form='identity'
              type='submit'
            >
              Verify domain
            </Button>
          </div>
        </>
      )}
      {identity.state == 'failed' && (
        <>
          <Card>
            <CardContent>
              <p>Your domain verification has failed. Please try again.</p>
              <img
                className='mt-4 max-w-[310px]'
                loading='lazy'
                alt='Identity card'
                src={`https://cdn.fynbos.app/identities/${identity.signatureHash}/domain.png`}
              />
            </CardContent>
          </Card>
          <div className='flex w-full space-x-2'>
            <OutlineButton
              shrink
              className='!text-error outline-error hover:!text-red-800 hover:outline-red-800 focus-visible:outline-red-800'
              type='button'
              onClick={() => setShowDialog(true)}
            >
              Delete
            </OutlineButton>
            <Button name='formName' value='retry' form='identity' type='submit'>
              Retry
            </Button>
          </div>
        </>
      )}
      {identity.state == 'pending' && (
        <>
          <Card>
            <CardContent>
              <p>
                Your domain verification is pending. We will notify you once
                verified.
              </p>
              <img
                className='mt-4 max-w-[310px]'
                loading='lazy'
                alt='Identity card'
                src={`https://cdn.fynbos.app/identities/${identity.signatureHash}/domain.png`}
              />
            </CardContent>
          </Card>
          <OutlineButton
            className='!text-error outline-error hover:!text-red-800 hover:outline-red-800 focus-visible:outline-red-800'
            type='button'
            onClick={() => setShowDialog(true)}
          >
            Delete
          </OutlineButton>
        </>
      )}
      <Dialog open={showDialog} setOpen={setShowDialog}>
        <CardHeader>
          <h1 className='text-xl font-medium'>Remove domain card</h1>
        </CardHeader>
        <CardContent>
          <span className='text-medium'>
            Are you sure you want to remove the domain card? This action cannot
            be undone.
          </span>

          <div className='flex w-full justify-end space-x-6 pt-2'>
            <TextButton
              type='button'
              className='!text-medium'
              onClick={() => setShowDialog(false)}
            >
              Cancel
            </TextButton>
            <TextButton
              name='formName'
              className='!text-error'
              value='delete'
              form='identity'
              type='submit'
            >
              Remove card
            </TextButton>
          </div>
        </CardContent>
      </Dialog>
    </>
  )
}

function Discord() {
  const { identity, publicName, walletInfo } = useLoaderData<typeof loader>()
  const [showDialog, setShowDialog] = useState<boolean>(false)

  const fetcher = useFetcher()
  const _onChangeSwitch = useCallback<{
    (formName: string, publish: boolean): void
  }>(
    (formName, publish) => {
      fetcher.submit(
        { formName, publish: publish.toString() },
        { method: 'post' }
      )
    },
    [fetcher]
  )

  return (
    <>
      <>
        <Card>
          <CardHeader>
            <CardTitle>Discord</CardTitle>
          </CardHeader>
          <CardContent>
            <img
              className='max-w-[310px]'
              loading='lazy'
              alt='Identity card'
              src={`https://cdn.fynbos.app/identities/${identity.signatureHash}/discord.png`}
            />
            <div className='mt-4 flex w-full flex-col space-y-1'>
              <span className='text-weak'>Verification date</span>
              <span className='font-medium'>{identity.verifiedAt}</span>
            </div>
          </CardContent>
        </Card>
        <Card>
          <Label className='mt-2'>Public profile</Label>
          <CardLink
            to={`/me/${walletInfo.formattedURL}`}
            className='items-center justify-between bg-nav'
          >
            <div className='flex space-x-3'>
              <Icon>contact_page</Icon>
              <span>{publicName}</span>
            </div>
            <Icon>navigate_next</Icon>
          </CardLink>
          <CardContent>
            <div className='flex items-center justify-between'>
              <span className='text-sm'>
                Show Discord handle on your Fynbos public profile
              </span>
              <Switch
                checked={identity.public}
                disabled={false}
                onChange={() => _onChangeSwitch('publish', !identity.public)}
              />
            </div>
          </CardContent>
        </Card>
        <OutlineButton
          className='!text-error outline-error hover:!text-red-800 hover:outline-red-800 focus-visible:outline-red-800'
          type='button'
          onClick={() => setShowDialog(true)}
        >
          Delete
        </OutlineButton>
      </>
      <Dialog open={showDialog} setOpen={setShowDialog}>
        <CardHeader>
          <h1 className='text-xl font-medium'>Remove Discord ID card</h1>
        </CardHeader>
        <CardContent>
          <span className='text-medium'>
            Are you sure you want to remove the Discord identity card? This
            action cannot be undone.
          </span>

          <div className='flex w-full justify-end space-x-6 pt-2'>
            <TextButton
              type='button'
              className='!text-medium'
              onClick={() => setShowDialog(false)}
            >
              Cancel
            </TextButton>
            <TextButton
              name='formName'
              className='!text-error'
              value='delete'
              form='identity'
              type='submit'
            >
              Remove card
            </TextButton>
          </div>
        </CardContent>
      </Dialog>
    </>
  )
}

function Slack() {
  const { identity, publicName, walletInfo } = useLoaderData<typeof loader>()
  const [showDialog, setShowDialog] = useState<boolean>(false)

  const fetcher = useFetcher()
  const _onChangeSwitch = useCallback<{
    (formName: string, publish: boolean): void
  }>(
    (formName, publish) => {
      fetcher.submit(
        { formName, publish: publish.toString() },
        { method: 'post' }
      )
    },
    [fetcher]
  )

  return (
    <>
      <>
        <Card>
          <CardHeader>
            <CardTitle>Slack</CardTitle>
          </CardHeader>
          <CardContent>
            <img
              className='max-w-[310px]'
              loading='lazy'
              alt='Identity card'
              src={`https://cdn.fynbos.app/identities/${identity.signatureHash}/slack.png`}
            />
            <div className='mt-4 flex w-full flex-col space-y-1'>
              <span className='text-weak'>Verification date</span>
              <span className='font-medium'>{identity.verifiedAt}</span>
            </div>
          </CardContent>
        </Card>
        <Card>
          <Label className='mt-2'>Public profile</Label>
          <CardLink
            to={`/me/${walletInfo.formattedURL}`}
            className='items-center justify-between bg-nav'
          >
            <div className='flex space-x-3'>
              <Icon>contact_page</Icon>
              <span>{publicName}</span>
            </div>
            <Icon>navigate_next</Icon>
          </CardLink>
          <CardContent>
            <div className='flex items-center justify-between'>
              <span className='text-sm'>
                Show Slack handle on your Fynbos public profile
              </span>
              <Switch
                checked={identity.public}
                disabled={false}
                onChange={() => _onChangeSwitch('publish', !identity.public)}
              />
            </div>
          </CardContent>
        </Card>
        <OutlineButton
          className='!text-error outline-error hover:!text-red-800 hover:outline-red-800 focus-visible:outline-red-800'
          type='button'
          onClick={() => setShowDialog(true)}
        >
          Delete
        </OutlineButton>
      </>
      <Dialog open={showDialog} setOpen={setShowDialog}>
        <CardHeader>
          <h1 className='text-xl font-medium'>Remove Slack ID card</h1>
        </CardHeader>
        <CardContent>
          <span className='text-medium'>
            Are you sure you want to remove the Slack identity card? This action
            cannot be undone.
          </span>

          <div className='flex w-full justify-end space-x-6 pt-2'>
            <TextButton
              type='button'
              className='!text-medium'
              onClick={() => setShowDialog(false)}
            >
              Cancel
            </TextButton>
            <TextButton
              name='formName'
              className='!text-error'
              value='delete'
              form='identity'
              type='submit'
            >
              Remove card
            </TextButton>
          </div>
        </CardContent>
      </Dialog>
    </>
  )
}

export async function action({ request, params }: ActionFunctionArgs) {
  const form = await request.formData()
  const formName = form.get('formName') as string
  const identityId = params.identityId as string
  const publish = form.get('publish') as string

  await validateCSRFToken(request, form)

  let response
  switch (formName) {
    case 'verify':
      response = await grpc.verifyIdentity(request, { id: identityId })
      if (isConnectError(response)) throw response.errorResponse
      return jsonWithSnackbar(
        request,
        {},
        {
          message: 'Connected identity verification started.',
          icon: 'close'
        }
      )
    case 'retry':
      response = await grpc.verifyIdentity(request, { id: identityId })
      if (isConnectError(response)) throw response.errorResponse
      return jsonWithSnackbar(
        request,
        {},
        {
          message: 'Retrying identity verification.',
          icon: 'close'
        }
      )
    case 'publish':
      response = await grpc.setIdentityPublic(request, {
        id: identityId,
        public: publish === 'true'
      })
      if (isConnectError(response)) throw response.errorResponse
      return jsonWithSnackbar(
        request,
        {},
        {
          message: 'Identity visibility updated.',
          icon: 'close'
        }
      )
    case 'delete':
      response = await grpc.deleteIdentity(request, { id: identityId })
      if (isConnectError(response)) throw response.errorResponse
      return redirectWithSnackbar(request, route('/identities'), {
        message: 'Identity deleted successfully.',
        icon: 'close'
      })
    default:
      throw json(
        { title: "Submitted a form that doesn't exist" },
        {
          status: 400
        }
      )
  }
}
