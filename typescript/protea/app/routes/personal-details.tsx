import type { ActionArgs, LoaderArgs, MetaFunction } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { useLoaderData, useSubmit } from '@remix-run/react'
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
import {SetKYCStatusPending} from "~/lib/wallet.server";

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
          submit(null, {
            action: '/personal-details',
            method: 'post'
          })
        },
        onCancel: ({ inquiryId, sessionToken }: any) => console.log('onCancel'),
        onError: (error: any) => console.log(error)
      })
    }
  }, [inquiryId, sessionToken, status, submit])

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
  await exitFlow(request, flowType.PersonalDetails)

  await SetKYCStatusPending(request)

  return redirect(route('/'), {
    headers: {
      'Set-Cookie': await flashSnackbar(request, {
        message: 'Personal details captured.',
        icon: 'close'
      })
    }
  })
}
