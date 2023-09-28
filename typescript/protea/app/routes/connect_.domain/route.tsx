import type {
  ActionFunctionArgs,
  LoaderFunctionArgs,
  MetaFunction
} from '@remix-run/node'

import { Code } from '@bufbuild/connect'
import { useEffect } from 'react'
import { route } from 'routes-gen'
import type { ApplicationProps } from '~/components'
import { Layouts } from '~/components'
import { jsonWithCSRF, validateCSRFToken } from '~/lib/csrf.server'
import { isConnectError } from '~/lib/error.server'
import { grpc } from '~/lib/grpc.server'
import { mergeMeta } from '~/lib/meta'
import { redirectWithSnackbar } from '~/lib/snackbar.server'
import {
  ConnectDomainStep,
  useConnectDomainStore
} from '~/lib/useConnectDomainStore'
import { Landing } from './Landing'
import { Name } from './Name'

export async function loader({ request }: LoaderFunctionArgs) {
  return jsonWithCSRF(request, {
    fynbosEnv: process.env.FYNBOS_ENV
  })
}

export const handle: ApplicationProps = {
  layout: Layouts.Focus,
  scaffold: {
    header: { title: 'Connect a domain', back: 'connect-domain' }
  }
}

export const meta: MetaFunction = mergeMeta(() => [
  {
    title: 'Connect a domain'
  }
])

export default function Page() {
  const [step, reset] = useConnectDomainStore((state) => [
    state.step,
    state.reset
  ])

  useEffect(() => {
    // This ensures that the state is only cleared when this route is unmounted.
    return () => {
      reset()
    }
  }, [reset])

  return (
    <>
      {step === ConnectDomainStep.LANDING && <Landing />}
      {step === ConnectDomainStep.NAME && <Name />}
    </>
  )
}

export async function action({ request }: ActionFunctionArgs) {
  const form = await request.formData()

  await validateCSRFToken(request, form)

  const errors = {
    form: '',
    domain: ''
  }

  const domain = form.get('domain') as string

  const response = await grpc.createDomainIdentity(request, {
    url: domain
  })

  if (isConnectError(response)) {
    if (response.code == Code.InvalidArgument) {
      return response.error({ errors })
    } else if (response.code == Code.AlreadyExists) {
      errors['domain'] = 'Domain is already connected.'
      return response.error({ errors })
    } else return response.error({ errors }, {}, { action: 'Contact support' })
  }

  return redirectWithSnackbar(
    request,
    route('/identities/:identityId', {
      identityId: response.id
    }),
    {
      message: 'Your domain was connected successfully.',
      icon: 'close'
    }
  )
}
