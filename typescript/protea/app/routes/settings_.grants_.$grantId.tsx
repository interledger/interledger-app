import type {
  ActionFunctionArgs,
  LoaderFunctionArgs,
  MetaFunction
} from '@remix-run/node'
import type { UIMatch } from '@remix-run/react'
import { Form, useLoaderData } from '@remix-run/react'
import { route } from 'routes-gen'
import type { ApplicationProps } from '~/components'
import { Card, CardContent, Layouts } from '~/components'
import { jsonWithCSRF, validateCSRFToken } from '~/lib/csrf.server'
import { isConnectError } from '~/lib/error.server'
import { grpc } from '~/lib/grpc.server'
import { mergeMeta } from '~/lib/meta'
import { redirectWithSnackbar } from '~/lib/snackbar.server'

export const handle: ApplicationProps = {
  layout: Layouts.Focus,
  scaffold: {
    header: {
      back: '/settings/keys',
      title: (match: UIMatch<typeof loader>) => match.data.grant.client
    }
  }
}

export const meta: MetaFunction<typeof loader> = mergeMeta(({ data }) => [
  {
    title: data?.grant.client || 'Grant'
  }
])

export async function loader({ request, params }: LoaderFunctionArgs) {
  const grant = await grpc.getRafikiGrant(request, {
    id: params.keyId as string
  })

  if (isConnectError(grant)) throw grant.errorResponse

  return jsonWithCSRF(request, {
    grant
  })
}

export default function Page() {
  const { grant, csrfToken } = useLoaderData<typeof loader>()

  return (
    <>
      <Form
        id='key-id'
        action={`/settings/keys/${grant.id}`}
        method='post'
        className='hidden'
      />
      <input form='key-id' value={csrfToken} name='csrfToken' type='hidden' />
      <Card>
        <code className='flex items-center justify-between break-all rounded-xl bg-nav p-2 font-mono text-medium'>
          {/*{grant.publicKeyFingerprint}*/}
        </code>
        <CardContent>
          {/*<p className='mt-4 text-sm text-medium'>Added {grant.createdAt}</p>*/}
          {/*TODO: implement last used*/}
          {/*<p className='mt-1 text-sm text-purple-500'>Last used {connection.lastUsedAt}</p>*/}
        </CardContent>
      </Card>
      {/*{grant.applicationName !== 'Fynbos Managed' && (*/}
      {/*  <OutlineButton*/}
      {/*    // TODO error token colors*/}
      {/*    className='text-red-700 outline-red-700 focus-visible:outline-red-800'*/}
      {/*    form='key-id'*/}
      {/*    type='submit'*/}
      {/*  >*/}
      {/*    Delete*/}
      {/*  </OutlineButton>*/}
      {/*)}*/}
    </>
  )
}

export async function action({ request, params }: ActionFunctionArgs) {
  const form = await request.formData()

  await validateCSRFToken(request, form)

  const errors = {
    form: ''
  }

  const response = await grpc.deleteConnection(request, {
    id: params.keyId as string
  })

  if (isConnectError(response)) {
    return response.error({ errors })
  }

  return redirectWithSnackbar(request, route('/settings/keys'), {
    message: 'Public key was deleted.',
    icon: 'close'
  })
}
