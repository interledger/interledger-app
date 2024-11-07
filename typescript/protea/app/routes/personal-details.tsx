import type {
  ActionFunctionArgs,
  LoaderFunctionArgs,
  MetaFunction
} from '@remix-run/node'
import { json } from '@remix-run/node'
import { useLoaderData, useSubmit } from '@remix-run/react'
import { useEffect, useRef, useState } from 'react'
import { route } from 'routes-gen'
import { Button, Card, CardContent, Dialog, Layouts, Shape } from '~/components'
import { isConnectError } from '~/lib/error.server'
import { exitFlow, flowType, requireFlow } from '~/lib/flows.server'
import { grpc } from '~/lib/grpc.server'
import { mergeMeta } from '~/lib/meta'
import { redirectWithSnackbar } from '~/lib/snackbar.server'
import { useScaffoldStore } from '~/lib/useScaffoldStore'
import { useScript } from '~/lib/useScript'

export async function loader({ request }: LoaderFunctionArgs) {
  const flow = await requireFlow(request, flowType.PersonalDetails)
  const response = await grpc.getKYCProviderWidget(request, {
    idempotencyKey: flow.data.idempotencyKey
  })

  if (isConnectError(response)) throw response.errorResponse

  return json({
    provider: response.provider,
    gatehubWidget: response.gatehubWidget,
    personaWidget: response.personaInquiry,
    chimoneyWidget: response.chimoneyWidget,
    ptiWidget: response.ptiWidget
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

export const meta: MetaFunction = mergeMeta(() => [
  {
    title: 'Activate wallet'
  }
])

function ChimoneyPage() {
  const submit = useSubmit()
  const { chimoneyWidget } = useLoaderData<typeof loader>()

  useEffect(() => {
    const onKYCComplete = (e: MessageEvent) => {
      if (!e.data) return

      if (e.data.kyc) {
        submit(null, {
          action: '/personal-details',
          method: 'post'
        })
      }
    }

    window.addEventListener('message', onKYCComplete)

    return () => {
      window.removeEventListener('message', onKYCComplete)
    }
  }, [submit])

  return (
    <iframe
      title='Activate wallet'
      src={chimoneyWidget}
      sandbox='allow-top-navigation allow-forms allow-same-origin allow-popups allow-scripts'
      scrolling='yes'
      allow='camera;microphone'
      className='h-[750px]'
    />
  )
}

function PersonaPage() {
  const submit = useSubmit()
  const { personaWidget } = useLoaderData<typeof loader>()
  const [ready, setReady] = useState(false)
  const status = useScript(
    'https://cdn.withpersona.com/dist/persona-v4.8.0-alpha.js'
  )
  let personaRef = useRef<any>(null)

  const [setLoading] = useScaffoldStore((state) => [state.setLoading])

  useEffect(() => {
    setLoading(!ready)
  }, [ready, setLoading])

  // Unmount make sure the loading state is set to false
  useEffect(() => {
    return () => {
      setLoading(false)
    }
  }, [setLoading])

  useEffect(() => {
    if (typeof window !== 'undefined' && status == 'ready') {
      personaRef.current = (window as any).Persona
      personaRef.current = new (window as any).Persona.Client({
        inquiryId: personaWidget?.id,
        sessionToken: personaWidget?.sessionToken,
        onReady: () => setReady(true),
        onComplete: ({ inquiryId, status, fields }: any) => {
          setReady(false)
          submit(null, {
            action: '/personal-details',
            method: 'post'
          })
        },
        onCancel: ({ inquiryId, sessionToken }: any) => console.log('onCancel'),
        onError: (error: any) => console.log(error)
      })
    }
  }, [personaWidget, status, submit])

  return <KycIntro onClick={() => personaRef.current.open()} ready={ready} />
}

function GatehubPage() {
  const { gatehubWidget } = useLoaderData<typeof loader>()
  const [dialogOpen, setDialogOpen] = useState(false)

  return (
    <>
      <KycIntro
        onClick={() => {
          setDialogOpen(true)
        }}
        ready
      />
      <Dialog open={dialogOpen} setOpen={setDialogOpen} grow>
        <iframe
          title='Activate wallet'
          src={gatehubWidget?.widgetUrl}
          sandbox='allow-top-navigation allow-forms allow-same-origin allow-popups allow-scripts'
          scrolling='yes'
          allow='camera;microphone'
          className='h-[750px]'
        />
      </Dialog>
    </>
  )
}

function PtiPage() {
  const { ptiWidget } = useLoaderData<typeof loader>()
  const scriptStatus = useScript(
    'https://sdk.platform.fiant.io/0.0.21/index.js'
  )

  useEffect(() => {
    if (scriptStatus == 'ready' && typeof (window as any).PTI !== 'undefined') {
      ;(window as any).PTI.init({
        clientId: ptiWidget?.clientId,
        generateTokenPath: ptiWidget?.generateTokenPath
      })
      ;(window as any).PTI.form({
        type: 'KYC',
        requestId: ptiWidget?.requestId,
        userId: ptiWidget?.userId,
        scenarioId: ptiWidget?.scenarioId,
        parentElement: document.getElementById('kyc_form'),
        lang: 'en'
      })
    }
  }, [scriptStatus, ptiWidget])

  return (
    <>
      <div id='kyc_form' className='h-[750px]' />
    </>
  )
}

type KycIntroProps = {
  onClick: () => void
  ready: Boolean
}

function KycIntro({ onClick, ready }: KycIntroProps) {
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
      <Button disabled={!ready} type='button' onClick={onClick}>
        Continue
      </Button>
    </>
  )
}

export default function Page() {
  const { provider } = useLoaderData<typeof loader>()

  if (provider == 'persona') {
    return <PersonaPage />
  } else if (provider == 'chimoney') {
    return <ChimoneyPage />
  } else if (provider == 'pti') {
    return <PtiPage />
  } else return <GatehubPage />
}

export async function action({ request }: ActionFunctionArgs) {
  await exitFlow(request, flowType.PersonalDetails)

  await grpc.setKYCStatusPending(request, {})

  return redirectWithSnackbar(request, route('/'), {
    message: 'Personal details captured.',
    icon: 'close'
  })
}
