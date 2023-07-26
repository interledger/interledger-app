import type { ActionArgs, LoaderArgs, MetaFunction } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { useLoaderData, useRouteLoaderData, useSubmit } from '@remix-run/react'
import { useEffect, useRef, useState } from 'react'
import { route } from 'routes-gen'
import { Button, Card, CardContent, Layouts, Shape } from '~/components'
import { exitFlow, flowType, requireFlow } from '~/lib/flows.server'
import {
  StatusError,
  grpcClient,
  httpMapping,
  isGrpcError
} from '~/lib/proto.server'
import { flashSnackbar } from '~/lib/snackbar.server'
import { useScaffoldStore } from '~/lib/useScaffoldStore'
import { useScript } from '~/lib/useScript'
import { getSession, validateCSRFToken } from '~/session.server'

export async function loader({ request }: LoaderArgs) {
  const flow = await requireFlow(request, flowType.PersonalDetails)
  const response = await grpcClient
    .getPersonaInquiry(
      {
        idempotencyKey: flow.data.idempotencyKey
      },
      {
        meta: {
          cookies: String(request.headers.get('cookie')) || ''
        }
      }
    )
    .then((v) => v)
    .catch(StatusError)
  if (isGrpcError(response)) {
    throw json({}, httpMapping(response.code))
  }

  return json({
    inquiryId: response.response.id,
    sessionToken: response.response.sessionToken
  })
}

export const handle = {
  layout: Layouts.Focus,
  scaffold: {
    header: {
      title: 'Activate wallet',
      back: route('/')
    }
  }
}

export const meta: MetaFunction = () => {
  return {
    title: 'Activate wallet'
  }
}

export default function Page() {
  const submit = useSubmit()
  const { inquiryId, sessionToken } = useLoaderData<typeof loader>()
  const { csrfToken } = useRouteLoaderData('root') as ReturnType<
    () => { csrfToken: string }
  >
  const [ready, setReady] = useState(false)
  const status = useScript(
    'https://cdn.withpersona.com/dist/persona-v4.8.0-alpha.js'
  )
  let personaRef = useRef<any>(null)

  const [setLoading] = useScaffoldStore((state) => [state.setLoading])

  useEffect(() => {
    setLoading(!ready)
  }, [ready, setLoading])

  useEffect(() => {
    if (typeof window !== 'undefined' && status == 'ready') {
      personaRef.current = (window as any).Persona
      personaRef.current = new (window as any).Persona.Client({
        inquiryId,
        sessionToken,
        onReady: () => setReady(true),
        onComplete: ({ inquiryId, status, fields }: any) => {
          let formData = new FormData()
          formData.append('csrfToken', csrfToken)
          submit(formData, {
            action: '/personal-details',
            method: 'post'
          })
        },
        onCancel: ({ inquiryId, sessionToken }: any) => console.log('onCancel'),
        onError: (error: any) => console.log(error)
      })
    }
  }, [inquiryId, sessionToken, status, submit, csrfToken])

  return (
    <>
      <Card>
        <CardContent>
          <span>
            Complete theses steps to confirm your identity and activate your
            wallet:
          </span>
          <div className='mt-6 flex items-start'>
            <Shape
              flex='flex-none'
              width='w-8'
              radius='rounded-full'
              color='bg-yellow-300'
            />
            <Shape
              flex='flex-none'
              width='w-8'
              radius='rounded-tl-full'
              color='bg-rose-400'
            />
            <div className='ml-5'>
              <h3 className='mb-1 font-medium text-strong'>Photo ID</h3>
              <p className='text-medium'>
                We require a photo of a government ID to verify your identity.
              </p>
            </div>
          </div>
          <div className='mt-10 flex items-start'>
            <Shape
              flex='flex-none'
              width='w-8'
              radius='rounded-t-full'
              color='bg-purple-200'
            />
            <Shape
              flex='flex-none'
              width='w-8'
              radius='rounded-tr-full'
              color='bg-purple-400'
            />
            <div className='ml-5'>
              <h3 className='mb-1 font-medium text-strong'>Personal details</h3>
              <p className='text-medium'>
                Confirmation of your personal details and your address.
              </p>
            </div>
          </div>
          <div className='mt-10 flex items-start'>
            <Shape
              flex='flex-none'
              width='w-8'
              radius='rounded-br-full'
              color='bg-slate-300'
            />
            <Shape
              flex='flex-none'
              width='w-8'
              radius='rounded-l-full'
              color='bg-lime-400'
            />
            <div className='ml-5'>
              <h3 className='mb-1 font-medium text-strong'>
                Selfie verification
              </h3>
              <p className='text-medium'>
                A picture of yourself taken using your smartphone, webcam or
                tablet.
              </p>
            </div>
          </div>
        </CardContent>
      </Card>
      <Button
        disabled={!ready}
        type='button'
        onClick={() => personaRef.current.open()}
      >
        Continue
      </Button>
    </>
  )
}

export async function action({ request }: ActionArgs) {
  const form = await request.formData()
  const csrfToken = form.get('csrfToken') as string
  if (
    !validateCSRFToken(
      csrfToken,
      await getSession(request.headers.get('Cookie') as string)
    )
  ) {
    throw json(
      {
        action: {
          route: route('/personal-details'),
          text: 'Try again'
        }
      },
      { status: 422, statusText: 'Invalid CSRF token.' }
    )
  }

  await exitFlow(request, flowType.PersonalDetails)

  return redirect(route('/'), {
    headers: {
      'Set-Cookie': await flashSnackbar(request, {
        message: 'Personal details captured.',
        icon: 'close'
      })
    }
  })
}
