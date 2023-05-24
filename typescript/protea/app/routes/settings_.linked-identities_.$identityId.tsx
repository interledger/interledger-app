import type { ActionArgs, LoaderArgs, MetaFunction } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import type { RouteMatch } from '@remix-run/react'
import { Form, useFetcher, useLoaderData } from '@remix-run/react'
import {
  Button,
  Card,
  Chip,
  ChipColor,
  Icon,
  Layouts,
  OutlineButton,
  Router,
  Snackbar,
  Switch
} from '~/components'
import { route } from 'routes-gen'
import { flashSnackbar, getSnackbar } from '~/lib/snackbar.server'
import { getIdentity } from '~/lib/wallet.server'
import { KycStatus } from '~/routes/_index/route'
import { useCallback } from 'react'

export async function loader({ request, params }: LoaderArgs) {
  const identity = await getIdentity(request, params.identityId as string)
  const snackbar = await getSnackbar(request)
  return json({
    snackbar,
    identity
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
  const { identity, snackbar } = useLoaderData<typeof loader>()
  // const [showSnackbar, setSnackbar] = useState<boolean>(snackbar.show ?? false)

  const fetcher = useFetcher()
  const _onChangeSwitch = useCallback<{
    (formName: string, val: boolean): void
  }>(
    (formName, val) => {
      fetcher.submit({ formName, val: val.toString() }, { method: 'post' })
    },
    [fetcher]
  )
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
            {/*TODO Use for handling fallback of card*/}
            {/*<object data="https://stackoverflow.com/does-not-exist.png" type="image/png">*/}
            {/*  <img src="https://cdn.fynbos.app/identities/template.png" alt="Fallback card image">*/}
            {/*</object>*/}
            <img
              className='mt-4 max-w-[310px]'
              loading='lazy'
              alt='Identity card'
              src='https://cdn.fynbos.app/identities/template.png'
            />
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
            <h2 className='text-sm font-medium'>Profile</h2>
            <Router
              to={route('/settings/profile-personal')}
              className='mt-2 flex items-center justify-between rounded-xl bg-nav p-3 text-medium hover:bg-nav-hover'
            >
              <div className='flex space-x-3'>
                <Icon>contact_page</Icon>
                <span>Public name</span>
              </div>
              <Icon>navigate_next</Icon>
            </Router>
          </Card>
          <OutlineButton
            name='formName'
            value='delete'
            form='identity'
            type='submit'
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
              The link to your Twitter identity has failed. Please try again.
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
            The link to your Twitter identity is pending. We will notify you
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
        show={snackbar.show}
        id='cookie-snackbar'
        dismissAfter={3000}
        offset
        onClose={() => false}
      />
    </>
  )
}

export async function action({ request, params }: ActionArgs) {
  const form = await request.formData()
  const formName = (await form.get('formName')) as string
  console.log('formName', formName)
  const identityId = params.identityId as string

  switch (formName) {
    case 'verify':
      console.log('HERE')
      // await verifyTwitterIdentity(request, identityId)
      return redirect(`/settings/linked-identities/${identityId}`, {
        headers: {
          'Set-Cookie': await flashSnackbar(request, {
            message: 'Linked identity verification in progress',
            icon: 'close'
          })
        }
      })
    // return redirect(route('/settings/linked-identities'))

    default:
      throw json(
        { title: "Submitted a form that doesn't exist" },
        {
          status: 400
        }
      )
  }
}
