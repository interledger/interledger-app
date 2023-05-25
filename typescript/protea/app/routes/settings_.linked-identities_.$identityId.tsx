import type { ActionArgs, LoaderArgs, MetaFunction } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import type { RouteMatch } from '@remix-run/react'
import {
  Form,
  useFetcher,
  useActionData,
  useLoaderData,
  useRevalidator
} from '@remix-run/react'
import {
  AnchorRouter,
  Button,
  Card,
  Dialog,
  Icon,
  Layouts,
  OutlineButton,
  Router,
  Snackbar,
  Switch,
  TextButton
} from '~/components'
import { route } from 'routes-gen'
import { flashSnackbar, getSnackbar } from '~/lib/snackbar.server'
import {
  deleteTwitterIdentity,
  getIdentity,
  getPublicWalletDetails,
  getWalletPaymentPointer,
  setTwitterIdentityPublic,
  verifyTwitterIdentity
} from '~/lib/wallet.server'
import { useCallback, useEffect, useState } from 'react'
import { DateTime } from 'luxon'

export async function loader({ request, params }: LoaderArgs) {
  const paymentPointer = await getWalletPaymentPointer(request)
  const { publicName } = await getPublicWalletDetails(
    request,
    paymentPointer.walletID
  )
  const identity = await getIdentity(request, params.identityId as string)
  const snackbar = await getSnackbar(request)
  return json({
    snackbar,
    paymentPointer,
    publicName,
    identity: {
      ...identity,
      verifiedAt: DateTime.fromSeconds(
        parseInt(identity.verifiedAt?.seconds ?? '')
      ).toFormat('dd MMM yyyy')
    }
  })
}

export const handle = {
  title: (match: RouteMatch) => `@${match.data.identity.identifier}`,
  layout: Layouts.FocusLayout
}

export const meta: MetaFunction = () => {
  return {
    title: 'Settings | Edit public name'
  }
}

export default function Page() {
  const { identity, publicName, paymentPointer, snackbar } =
    useLoaderData<typeof loader>()
  const response = useActionData<typeof action>()

  const { revalidate } = useRevalidator()

  const [showSnackbar, setSnackbar] = useState<boolean>(
    (snackbar.show || response?.show) ?? false
  )
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

  useEffect(() => {
    setSnackbar((snackbar.show || response?.show) ?? false)
  }, [snackbar, response])

  // Polling for now until we have pusher for this
  useEffect(() => {
    const interval = setInterval(() => {
      if (identity.state == 'pending') revalidate()
    }, 5000)
    return () => clearInterval(interval)
  }, [identity.state, revalidate])

  return (
    <>
      <Form
        id='identity'
        action={`/settings/linked-identities/${identity.id}`}
        method='post'
        className='hidden'
      />
      {identity.state == 'verified' && (
        <>
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
                className='font-medium text-primary'
              >
                {identity.proof}
              </AnchorRouter>
            </Card.Item>
          </Card>
          <Card className='!bg-rose-100 dark:!bg-rose-950'>
            <Card.Item variant='col'>
              <span className='text-medium font-medium text-xs'>
                Please note
              </span>
              <span className='text-xs text-medium'>
                Do not delete the public proof from your Twitter timeline. Doing
                so will result in your identity no longer being verified by
                Fynbos.
              </span>
            </Card.Item>
          </Card>
          <Card>
            <div className='flex justify-between items-center'>
              <span className='text-sm'>
                Show on your Fynbos public profile.
              </span>
              <Switch
                checked={identity.public}
                disabled={false}
                onChange={() => _onChangeSwitch('publish', !identity.public)}
              />
            </div>
            <h2 className='text-sm font-medium mt-6'>Public profile</h2>
            <Router
              to={`/me/${paymentPointer.formatted}`}
              className='mt-2 flex items-center justify-between rounded-xl bg-nav p-3 text-medium hover:bg-nav-hover'
            >
              <div className='flex space-x-3'>
                <Icon>contact_page</Icon>
                <span>{publicName}</span>
              </div>
              <Icon>navigate_next</Icon>
            </Router>
          </Card>
          <OutlineButton
            className='!text-error outline-error focus-visible:outline-red-800 hover:!text-red-800 hover:outline-red-800'
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
            <p>
              To prove that you own this Twitter handle (and that you are a real
              person) we're going to post this Twitter identity card onto your
              timeline and link it back to your wallet.
            </p>
            <img
              className='mt-4 max-w-[310px]'
              loading='lazy'
              alt='Identity card'
              src='https://cdn.fynbos.app/identities/template.png'
            />
            <Router
              className='mt-4 text-sm font-medium text-primary rounded focus-visible:outline focus-visible:outline-2 focus-visible:outline-focus'
              to={route('/connect/twitter')}
            >
              Read more about how we verify your Twitter handle.
            </Router>
          </Card>
          <Button name='formName' value='verify' form='identity' type='submit'>
            Continue
          </Button>
        </>
      )}
      {identity.state == 'failed' && (
        <>
          <Card>
            <p>
              Your Twitter identity verification has failed. Please try again.
            </p>
            <img
              className='mt-4 max-w-[310px]'
              loading='lazy'
              alt='Identity card'
              src='https://cdn.fynbos.app/identities/template.png'
            />
          </Card>
          <Button name='formName' value='retry' form='identity' type='submit'>
            Retry
          </Button>
        </>
      )}
      {identity.state == 'pending' && (
        <Card>
          <p>
            Your Twitter identity verification is pending. We will notify you
            once verified.
          </p>
          <img
            className='mt-4 max-w-[310px]'
            loading='lazy'
            alt='Identity card'
            src='https://cdn.fynbos.app/identities/template.png'
          />
        </Card>
      )}
      <Snackbar
        message={snackbar.message}
        action={snackbar.action}
        icon={snackbar.icon}
        show={showSnackbar}
        id='cookie-snackbar'
        dismissAfter={3000}
        onClose={() => setSnackbar(false)}
      />
      <Dialog open={showDialog} setOpen={setShowDialog}>
        <h1 className='font-display text-2xl'>Remove Twitter ID card</h1>
        <span className='text-medium'>
          Are you sure you want to remove the Twitter identity card? This action
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
          <Form
            id='signup-phone-otp-validation'
            action='/signup/phone'
            method='post'
            className='hidden'
          />
        </div>
      </Dialog>
    </>
  )
}

export async function action({ request, params }: ActionArgs) {
  const form = await request.formData()
  const formName = (await form.get('formName')) as string
  const identityId = params.identityId as string
  const publish = (await form.get('publish')) as string

  switch (formName) {
    case 'verify':
      await verifyTwitterIdentity(request, identityId)
      return json(
        { show: true },
        {
          headers: {
            'Set-Cookie': await flashSnackbar(request, {
              message: 'Linked identity verification started.',
              icon: 'close'
            })
          }
        }
      )
    case 'retry':
      await verifyTwitterIdentity(request, identityId)
      return json(
        { show: true },
        {
          headers: {
            'Set-Cookie': await flashSnackbar(request, {
              message: 'Retrying identity verification.',
              icon: 'close'
            })
          }
        }
      )
    case 'publish':
      await setTwitterIdentityPublic(request, identityId, publish === 'true')
      return json(
        { show: true },
        {
          headers: {
            'Set-Cookie': await flashSnackbar(request, {
              message: 'Identity visibility updated.',
              icon: 'close'
            })
          }
        }
      )
    case 'delete':
      await deleteTwitterIdentity(request, identityId)
      return redirect(route('/settings/linked-identities'), {
        headers: {
          'Set-Cookie': await flashSnackbar(request, {
            message: 'Identity deleted successfully.',
            icon: 'close'
          })
        }
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
