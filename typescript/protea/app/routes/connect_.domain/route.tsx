import type { ActionArgs, LoaderArgs, MetaFunction } from '@remix-run/node'

import { useEffect } from 'react'
import { route } from 'routes-gen'
import type { ApplicationProps } from '~/components'
import { Layouts } from '~/components'
import { Code } from '~/generated/protobuf-ts/google/rpc/code'
import { jsonWithCSRF, validateCSRFToken } from '~/lib/csrf.server'
import { error } from '~/lib/error.server'
import type { GrpcError } from '~/lib/proto.server'
import { StatusError, grpcClient, isGrpcError } from '~/lib/proto.server'
import { redirectWithSnackbar } from '~/lib/snackbar.server'
import {
  ConnectDomainStep,
  useConnectDomainStore
} from '~/lib/useConnectDomainStore'
import { Landing } from './Landing'
import { Name } from './Name'

export async function loader({ request }: LoaderArgs) {
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

export const meta: MetaFunction = () => {
  return {
    title: 'Connect a domain'
  }
}

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

// The field names given by the backend for field violations
type fieldErrorsMap = 'Domain'

function mapper(field: fieldErrorsMap): 'domainName' | null {
  switch (field) {
    case 'Domain':
      return 'domainName'
    default:
      return null
  }
}
export async function action({ request }: ActionArgs) {
  const form = await request.formData()

  await validateCSRFToken(request, form)

  const fieldErrors = {
    form: '',
    domainName: ''
  }

  const domainName = form.get('domainName') as string

  const response = await grpcClient
    .createDomainIdentity(
      {
        url: domainName
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
    if (response.code == Code.INVALID_ARGUMENT) {
      for (let violation of (response as GrpcError).details[0]
        .fieldViolations) {
        const field = mapper(violation.field as fieldErrorsMap)
        if (field != null) fieldErrors[field] = violation.description
      }
      return error(request, { errors: { ...fieldErrors } })
    } else if (response.code == Code.ALREADY_EXISTS) {
      fieldErrors['domainName'] = 'Domain is already connected.'
      return error(request, { errors: { ...fieldErrors } })
    } else
      return error(
        request,
        { errors: { ...fieldErrors } },
        { action: 'Contact support' }
      )
  }

  return redirectWithSnackbar(
    request,
    route('/identities/:identityId', {
      identityId: response.response.id
    }),
    {
      message: 'Your domain was connected successfully.',
      icon: 'close'
    }
  )
}
