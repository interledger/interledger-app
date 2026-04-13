import type { Route } from './+types/connect_.card'
import { Code } from '@bufbuild/connect'
import { useLoaderData, useSubmit } from 'react-router';
import { useEffect } from 'react'
import { href } from 'react-router'
import type { ApplicationProps } from '~/components'
import { Layouts } from '~/components'
import { jsonWithCSRF, validateCSRFToken } from '~/lib/csrf.server'
import { isConnectError } from '~/lib/error.server'
import type { FiantSdkMessage } from '~/lib/fiant'
import { grpc } from '~/lib/grpc.server'
import { mergeMeta } from '~/lib/meta'
import { redirectWithSnackbar } from '~/lib/snackbar.server'
import { useScaffoldStore } from '~/lib/useScaffoldStore'
import { useScript } from '~/lib/useScript'
import { usePtiConfig } from '~/lib/pti-context'

export async function loader({ request }: Route.LoaderArgs) {
  const response = await grpc.getKYCProviderWidget(request, {
    idempotencyKey: ''
  })

  if (isConnectError(response)) throw response.errorResponse

  return jsonWithCSRF(request, {
    widget: response.ptiWidget
  })
}

export const handle: ApplicationProps = {
  layout: Layouts.Focus,
  scaffold: {
    header: {
      back: href('/accounts'),
      title: 'Connect card'
    }
  }
}

export const meta = mergeMeta(() => [
  {
    title: 'Connect card'
  }
])

export default function Page() {
  const { widget, csrfToken } = useLoaderData()
  const submit = useSubmit()
  // const actionData = useActionData()
  const [setLoading] = useScaffoldStore((state) => [state.setLoading])
  const ptiConfig = usePtiConfig()
  const scriptStatus = useScript(ptiConfig?.sdkUrl ?? '')
  useEffect(() => {
    // This ensures that loading is false when this route is unmounted.
    return () => {
      setLoading(false)
    }
  }, [setLoading])

  useEffect(() => {
    if (scriptStatus == 'ready' && window.PTI !== undefined) {
      const styling = {
        mode: window.matchMedia('(prefers-color-scheme: dark)').matches
          ? 'dark'
          : 'light',
        primaryColor: '#3b82f6',
        backgroundColor: '#f8fafc',
        fontFamily: 'Poppins'
      }

      window.PTI.init({
        clientId: widget?.clientId,
        generateTokenPath: widget?.generateTokenPath,
        ptiFormsUrl: ptiConfig?.formsUrl
      })
      window.PTI.form({
        type: 'ADD_CC',
        requestId: widget?.requestId,
        userId: widget?.userId,
        scenarioId: widget?.scenarioId,
        parentElement: document.getElementById('card_form'),
        lang: 'en',
        styleConfig: styling
      })
      setLoading(false)
    }

    const handleMessage = (message: MessageEvent<FiantSdkMessage>) => {
      if (message.data.name === 'AddCreditCardCompleted') {
        setLoading(true)
        let formData = new FormData()
        formData.append('tokenId', message.data.createdId)
        formData.append('csrfToken', csrfToken)
        submit(formData, {
          action: `/connect/card`,
          method: 'post'
        })
      }
    }
    window.addEventListener('message', handleMessage)

    return () => {
      window.removeEventListener('message', handleMessage)
    }
  }, [scriptStatus, widget, ptiConfig, setLoading, submit, csrfToken])

  return (
    <>
      <div id='card_form' className='h-[750px]' />
    </>
  )
}

export async function action({ request }: Route.ActionArgs) {
  const form = await request.formData()
  const cardToken = form.get('tokenId') as string

  await validateCSRFToken(request, form)

  const errors = {
    form: '',
    number: ''
  }
  const mapping = {
    number: 'CardNumber'
  }

  let response = await grpc.createCard(
    request,
    { tokenID: cardToken },
    {
      timeoutMs: 60 * 1000
    }
  )

  if (isConnectError(response)) {
    if (response.code == Code.InvalidArgument) {
      return response.error({ errors }, mapping)
    } else if (response.code == Code.FailedPrecondition) {
      errors.form = response.violations[0].description
      return response.error({ errors }, mapping)
    } else if (response.code == Code.AlreadyExists) {
      errors.form = 'This card is already connected.'
      return response.error({ errors }, mapping)
    } else {
      if (response.code == Code.Unavailable) {
        errors.form = 'We did not receive a response from our card processor.'
      }
      if (errors.form == '') {
        errors.form = 'There was an error connecting your card.'
      }
      return response.error({ errors }, mapping, { action: 'Contact support' })
    }
  }

  return redirectWithSnackbar(
    request,
    href('/accounts/:accountId', { accountId: response.id }),
    {
      message: 'New card successfully saved.',
      icon: 'close'
    }
  )
}
