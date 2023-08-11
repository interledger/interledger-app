import type { ActionArgs, LoaderArgs } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { Form, useLoaderData, useNavigate } from '@remix-run/react'
import { useEffect } from 'react'
import { route } from 'routes-gen'
import type { ApplicationProps } from '~/components'
import {
  Button,
  Card,
  CardContent,
  Layouts,
  LoadingShapes,
  Shape
} from '~/components'
import { jsonWithCSRF, validateCSRFToken } from '~/lib/csrf.server'
import { error } from '~/lib/error.server'
import {
  StatusError,
  grpcClient,
  httpMapping,
  isGrpcError
} from '~/lib/proto.server'

export const handle: ApplicationProps = {
  layout: Layouts.Focus,
  scaffold: {
    header: {
      back: route('/'),
      title: 'Connect a Twitter identity'
    }
  }
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

    return jsonWithCSRF(request, { id: resp.response.id })
  } else {
    return jsonWithCSRF(request, { id: null })
  }
}

export default function Page() {
  let { id, csrfToken } = useLoaderData<typeof loader>()
  const nav = useNavigate()

  // We do this redirect clientside because the browser removes secure cookies when coming from another domain.
  useEffect(() => {
    if (id != null) nav(`/identities/${id}?back=4`)
  }, [id, nav])

  return (
    <>
      <Form
        id='connect-twitter'
        action={'/connect/twitter'}
        method='post'
        className='hidden'
      />
      <input
        form='connect-twitter'
        value={csrfToken}
        name='csrfToken'
        type='hidden'
      />
      {id && (
        <Card>
          <CardContent>
            <LoadingShapes />
          </CardContent>
        </Card>
      )}
      {!id && (
        <>
          <Card>
            <CardContent>
              <p>To link your Twitter account, simply follow these steps:</p>
              <div className='mt-6 flex items-start'>
                <Shape
                  width='w-8'
                  flex='flex-none'
                  radius='rounded-tl-full'
                  color='bg-yellow-400'
                />
                <Shape
                  width='w-8'
                  flex='flex-none'
                  radius='rounded-l-full'
                  color='bg-rose-300'
                />
                <div className='ml-5'>
                  <h3 className='mb-1 font-medium text-strong'>Log in</h3>
                  <p className='text-medium'>Open and log in to Twitter.</p>
                </div>
              </div>
              <div className='mt-6 flex items-start'>
                <Shape
                  width='w-8'
                  flex='flex-none'
                  radius='rounded-full'
                  color='bg-lime-400'
                />
                <Shape
                  width='w-8'
                  flex='flex-none'
                  radius='rounded-tl-full rounded-br-full'
                  color='bg-indigo-400'
                />
                <div className='ml-5'>
                  <h3 className='mb-1 font-medium text-strong'>Authorize</h3>
                  <p className='text-medium'>
                    Grant Fynbos permission to connect to your Twitter account.
                  </p>
                </div>
              </div>
              <div className='mt-6 flex items-start'>
                <Shape
                  width='w-8'
                  flex='flex-none'
                  radius='rounded-full rounded-bl-none'
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
                  <p className='text-medium'>
                    Keep the generated Tweet on your timeline which contains a
                    signature that proves you are the owner of your Fynbos
                    wallet.
                  </p>
                </div>
              </div>
            </CardContent>
          </Card>
          <Button form='connect-twitter' type='submit'>
            Continue
          </Button>
        </>
      )}
    </>
  )
}

export async function action({ request }: ActionArgs) {
  const form = await request.formData()

  await validateCSRFToken(request, form)

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
    return error(request, {}, true, 'Contact support')
  }

  return redirect(resp.url)
}
