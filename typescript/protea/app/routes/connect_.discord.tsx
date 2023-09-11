import type { ActionArgs, LoaderArgs } from '@remix-run/node'
import { redirect } from '@remix-run/node'
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
import { connectClient } from '~/lib/connect.server'
import { jsonWithCSRF, validateCSRFToken } from '~/lib/csrf.server'
import { isConnectError } from '~/lib/error.server'

export const handle: ApplicationProps = {
  layout: Layouts.Focus,
  scaffold: {
    header: {
      back: route('/'),
      title: 'Connect a Discord identity'
    }
  }
}

export async function loader({ request }: LoaderArgs) {
  // Check for state and code if not state create auth url
  let url = new URL(request.url)
  let state = url.searchParams.get('state')
  let code = url.searchParams.get('code')

  if (state && code) {
    let response = await connectClient.discordCallback(request, {
      state: state.toString(),
      code: code.toString()
    })

    if (isConnectError(response)) throw response.errorResponse

    return jsonWithCSRF(request, { id: response.id })
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
        id='connect-discord'
        action={'/connect/discord'}
        method='post'
        className='hidden'
      />
      <input
        form='connect-discord'
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
              <p>To link your Discord account, simply follow these steps:</p>
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
                  <p className='text-medium'>Open and log in to Discord.</p>
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
                    Grant Fynbos permission to connect to your Discord account.
                  </p>
                </div>
              </div>
            </CardContent>
          </Card>
          <Button form='connect-discord' type='submit'>
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

  const errors = {
    form: ''
  }

  let response = await connectClient.createDiscordAuthURL(request, {})

  if (isConnectError(response)) {
    return response.error({ errors }, {}, { action: 'Contact support' })
  }

  return redirect(response.url)
}
