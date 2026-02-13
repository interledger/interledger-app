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
import type { FiantSdkMessage } from '~/lib/fiant'
import { exitFlow, flowType, requireFlow } from '~/lib/flows.server'
import { grpc } from '~/lib/grpc.server'
import logger from '~/lib/logger.server'
import { mergeMeta } from '~/lib/meta'
import { redirectWithSnackbar } from '~/lib/snackbar.server'
import { useScaffoldStore } from '~/lib/useScaffoldStore'
import { useScript } from '~/lib/useScript'

const KYCErrors: KYCErrorsType = {
  UnableToPars: 'KYC: unable to parse message data'
}

type KYCErrorsType = {
  UnableToPars: 'KYC: unable to parse message data'
}

export async function loader({ request }: LoaderFunctionArgs) {
  const flow = await requireFlow(request, flowType.PersonalDetails)
  const response = await grpc.getKYCProviderWidget(request, {
    idempotencyKey: flow.data.idempotencyKey
  })

  if (isConnectError(response)) throw response.errorResponse

  logger.info(
    {
      provider: response.provider,
      hasGatehubWidget: !!response.gatehubWidget,
      gatehubWidgetUrl: response.gatehubWidget?.widgetUrl,
      hasPersonaWidget: !!response.personaInquiry,
      hasChimoneyWidget: !!response.chimoneyWidget,
      hasPtiWidget: !!response.ptiWidget,
      flow: 'kyc'
    },
    '[KYC] Personal details page loaded'
  )

  // if (response.provider === 'local') {
  //   // wait 1s on local for the async processes to finish
  //   await new Promise((resolve) => setTimeout(resolve, 1000))
  //   throw redirect('/')
  // }

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
      className='h-[750px] sm:min-w-[400px] md:min-w-[400px]'
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
  const submit = useSubmit()
  useEffect(() => {
    const onKYCComplete = (e: MessageEvent) => {
      console.log('[KYC] Received message from iframe:', {
        origin: e.origin,
        data: e.data,
        hasType: !!e.data?.type,
        hasValue: !!e.data?.value
      })
      if (!e.data?.type || !e.data?.value) {
        console.warn('[KYC] Message missing type or value, ignoring')
        return
      }

      let parsedValue
      try {
        parsedValue = JSON.parse(e.data.value)
        console.log('[KYC] Parsed message value:', parsedValue)
      } catch (parseError) {
        console.error('[KYC] Failed to parse message value:', {
          error: parseError,
          rawValue: e.data.value
        })
        throw new Error(KYCErrors.UnableToPars)
      }

      if (
        e.data.type === 'OnboardingCompleted' &&
        parsedValue?.applicantStatus === 'submitted'
      ) {
        console.log('[KYC] Onboarding completed, submitting form to backend')
        submit(null, {
          action: '/personal-details',
          method: 'post'
        })
      } else {
        console.log(
          '[KYC] Message received but not OnboardingCompleted or wrong status:',
          {
            type: e.data.type,
            applicantStatus: parsedValue?.applicantStatus
          }
        )
      }
    }

    window.addEventListener('message', onKYCComplete)

    return () => {
      window.removeEventListener('message', onKYCComplete)
    }
  }, [submit])

  return (
    <>
      <KycIntro
        onClick={() => {
          console.log('[KYC] Opening gatehub KYC dialog', {
            hasWidgetUrl: !!gatehubWidget?.widgetUrl
          })
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
          className='h-[750px] sm:min-w-[400px] md:min-w-[400px]'
          onLoad={() => console.log('[KYC] Gatehub iframe loaded successfully')}
          onError={(e) => console.error('[KYC] Gatehub iframe load error:', e)}
        />
      </Dialog>
    </>
  )
}

function PtiPage() {
  const { ptiWidget } = useLoaderData<typeof loader>()
  const submit = useSubmit()
  const scriptStatus = useScript(
    ptiWidget?.sdkUrl || 'https://sdk.platform.fiant.io/0.0.23/index.js'
  )
  const [setLoading] = useScaffoldStore((state) => [state.setLoading])

  // Unmount make sure the loading state is set to false
  useEffect(() => {
    return () => {
      setLoading(false)
    }
  }, [setLoading])

  useEffect(() => {
    if (scriptStatus == 'ready' && typeof (window as any).PTI !== 'undefined') {
      ;(window as any).PTI.init({
        clientId: ptiWidget?.clientId,
        generateTokenPath: ptiWidget?.generateTokenPath,
        ptiFormsUrl: ptiWidget?.formsUrl || 'https://forms.platform.fiant.io'
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

    const handleMessage = (message: MessageEvent<FiantSdkMessage>) => {
      if (message.data.name === 'UserAssessmentCompleted') {
        setLoading(true)
        submit(null, {
          action: '/personal-details',
          method: 'post'
        })
      }
    }
    window.addEventListener('message', handleMessage)

    return () => {
      window.removeEventListener('message', handleMessage)
    }
  }, [scriptStatus, ptiWidget, setLoading, submit])

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
  logger.info(
    { flow: 'kyc' },
    '[KYC] Personal details action called - marking KYC status as pending'
  )

  await exitFlow(request, flowType.PersonalDetails)

  const setKycResponse = await grpc.setKYCStatusPending(request, {})
  if (isConnectError(setKycResponse)) {
    logger.error(
      { error: setKycResponse, flow: 'kyc' },
      '[KYC] Failed to set KYC status as pending'
    )
    throw setKycResponse.errorResponse
  }

  logger.info({ flow: 'kyc' }, '[KYC] KYC status set to pending successfully')

  return redirectWithSnackbar(request, route('/'), {
    message: 'Personal details captured.',
    icon: 'close'
  })
}
