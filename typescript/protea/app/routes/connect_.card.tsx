import {
  type ActionFunctionArgs,
  type LoaderFunctionArgs,
  type MetaFunction
} from '@remix-run/node'
import { useActionData, useLoaderData, useSubmit } from '@remix-run/react'
import { useEffect } from 'react'
import { Code } from '@bufbuild/connect'
import { route } from 'routes-gen'
import type { ApplicationProps } from '~/components'
import {
  Layouts,
} from '~/components'
import { jsonWithCSRF, validateCSRFToken } from '~/lib/csrf.server'
import { isConnectError } from '~/lib/error.server'
import { grpc } from '~/lib/grpc.server'
import { mergeMeta } from '~/lib/meta'
import { redirectWithSnackbar } from '~/lib/snackbar.server'
import { useScaffoldStore } from '~/lib/useScaffoldStore'
import { useScript } from '~/lib/useScript'
import { FiantSdkMessage } from '~/lib/fiant'

export async function loader({ request }: LoaderFunctionArgs) {
  const response = await grpc.getKYCProviderWidget(request, {
    idempotencyKey: ''
  })

  if (isConnectError(response)) throw response.errorResponse

  return jsonWithCSRF(request, {
    widget: response.ptiWidget,
  })
}

export const handle: ApplicationProps = {
  layout: Layouts.Focus,
  scaffold: {
    header: {
      back: route('/accounts'),
      title: 'Connect card'
    }
  }
}

export const meta: MetaFunction = mergeMeta(() => [
  {
    title: 'Connect card'
  }
])

export default function Page() {
  const { widget, csrfToken } = useLoaderData<typeof loader>()
  const submit = useSubmit()
  // const actionData = useActionData<typeof action>()
  const [setLoading] = useScaffoldStore((state) => [state.setLoading])
  const scriptStatus = useScript(
    widget?.sdkUrl || 'https://sdk.platform.fiant.io/0.0.23/index.js'
  )
  useEffect(() => {
    // This ensures that loading is false when this route is unmounted.
    return () => {
      setLoading(false)
    }
  }, [setLoading])

  useEffect(() => {
    if (scriptStatus == 'ready' && typeof (window as any).PTI !== 'undefined') {
      ; (window as any).PTI.init({
        clientId: widget?.clientId,
        generateTokenPath: widget?.generateTokenPath,
        ptiFormsUrl: widget?.formsUrl || 'https://forms.platform.fiant.io'
      })
        ; (window as any).PTI.form({
          type: 'ADD_CC',
          requestId: widget?.requestId,
          userId: widget?.userId,
          scenarioId: widget?.scenarioId,
          parentElement: document.getElementById('card_form'),
          lang: 'en'
        })
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
  }, [scriptStatus, widget, setLoading, submit, csrfToken])

  return (
    <>
      <div id='card_form' className='h-[750px]' />
    </>
  )
}

export async function action({ request }: ActionFunctionArgs) {
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
      errors.form = 'This card is already connected to Fynbos.'
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
    route('/accounts/:accountId', { accountId: response.id }),
    {
      message: 'New card successfully saved.',
      icon: 'close'
    }
  )
}
