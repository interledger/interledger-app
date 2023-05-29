import type { ActionArgs, LoaderArgs } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import {
  StatusError,
  grpcClient,
  httpMapping,
  isGrpcError
} from '~/lib/proto.server'
import { Form, useLoaderData, useNavigate } from '@remix-run/react'
import { Button, Card, Layouts, Shape } from '~/components'
import { useEffect } from 'react'

export const handle = {
  title: 'Connect a Twitter identity',
  layout: Layouts.FocusLayout
}

export async function loader({ request }: LoaderArgs) {
  // Check for state and code if not state create auth url
  let url = new URL(request.url)
  let state = url.searchParams.get('state')
  let code = url.searchParams.get('code')

  if (state && code) {
    let resp = await grpcClient.twitterCallback(
      {
        state: state.toString(),
        code: code.toString()
      },
      {
        meta: {
          cookies: String(request.headers.get('cookie')) || ''
        }
      }
    )

    if (isGrpcError(resp)) {
      throw json({}, httpMapping(resp.code))
    }

    return json({
      id: resp.response.id
    })
  } else {
    return json({ id: null })
  }
}

export default function Page() {
  let { id } = useLoaderData<typeof loader>()
  const nav = useNavigate()

  // We do this redirect clientside because the browser removes secure cookies when coming from another domain.
  useEffect(() => {
    if (id != null) nav(`/settings/linked-identities/${id}`)
  }, [id, nav])

  return (
    <Card>
      <span>
        We will guide you through the following steps in order to link your
        Twitter account:
      </span>
      <div className='mt-10 flex items-start'>
        <Shape
          width='w-8'
          flex='flex-none'
          radius='rounded-tl-full'
          color='bg-yellow-400'
        />
        <Shape
          width='w-8'
          flex='flex-none'
          radius='rounded-tl-full'
          color='bg-rose-300'
        />
        <div className='ml-5'>
          <h3 className='mb-1 font-medium text-strong'>Log in</h3>
          <p className='text-sm text-medium'>Open and log in to Twitter.</p>
        </div>
      </div>
      <div className='mt-10 flex items-start'>
        <Shape
          width='w-8'
          flex='flex-none'
          radius='rounded-full'
          color='bg-lime-400'
        />
        <Shape
          width='w-8'
          flex='flex-none'
          radius='rounded-tl-full'
          color='bg-slate-300'
        />
        <div className='ml-5'>
          <h3 className='mb-1 font-medium text-strong'>Authorize</h3>
          <p className='text-sm text-medium'>
            Grant Fynbos permission to connect to your Twitter account.
          </p>
        </div>
      </div>
      <div className='mt-10 flex items-start'>
        <Shape
          width='w-8'
          flex='flex-none'
          radius='rounded-full'
          color='bg-yellow-300'
        />
        <Shape
          width='w-8'
          flex='flex-none'
          radius='rounded-r-full'
          color='bg-sky-300'
        />
        <div className='ml-5'>
          <h3 className='mb-1 font-medium text-strong'>Tweet</h3>
          <p className='text-sm text-medium'>
            Keep the generated Tweet on your timeline which contains a signature
            that proves you are the owner of your Fynbos wallet.
          </p>
        </div>
      </div>

      <Form
        id='connect-twitter'
        action={'/connect/twitter'}
        method='post'
        className='hidden'
      />
      <div className='mt-12'>
        <Button form='connect-twitter' type='submit'>
          Continue
        </Button>
      </div>
    </Card>
  )
}

export async function action({ request }: ActionArgs) {
  let resp = await grpcClient
    .createTwitterAuthURL(
      {},
      {
        meta: {
          cookies: String(request.headers.get('cookie')) || ''
        }
      }
    )
    .then((resp) => resp.response)
    .catch(StatusError)

  if (isGrpcError(resp)) {
    throw json({}, httpMapping(resp.code))
  }

  return redirect(resp.url)
}
