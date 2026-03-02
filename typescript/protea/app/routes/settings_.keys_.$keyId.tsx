import type { ActionFunctionArgs, LoaderFunctionArgs, MetaFunction } from 'react-router';
import type { UIMatch } from 'react-router';
import { Form, useLoaderData } from 'react-router';
import { href } from 'react-router'
import type { ApplicationProps } from '~/components'
import { Card, CardContent, Layouts, OutlineButton } from '~/components'
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
      title: (match: UIMatch<Awaited<ReturnType<typeof loader>>['data']>) =>
        match.data!.connection.applicationName
    }
  }
}

export const meta: MetaFunction<typeof loader> = mergeMeta(({ data }) => [
  {
    title: data?.connection.applicationName || 'Public key'
  }
])

export async function loader({ request, params }: LoaderFunctionArgs) {
  const connection = await grpc.getConnection(request, {
    id: params.keyId as string
  })

  if (isConnectError(connection)) throw connection.errorResponse

  return jsonWithCSRF(request, {
    connection
  })
}

export default function Page() {
  const { connection, csrfToken } = useLoaderData<typeof loader>()

  return (
    <>
      <Form
        id='key-id'
        action={`/settings/keys/${connection.id}`}
        method='post'
        className='hidden'
      />
      <input form='key-id' value={csrfToken} name='csrfToken' type='hidden' />
      <Card>
        <code className='flex items-center justify-between break-all rounded-xl bg-nav p-2 font-mono text-medium'>
          {connection.publicKeyFingerprint}
        </code>
        <CardContent>
          <p className='mt-4 text-sm text-medium'>
            Added {connection.createdAt}
          </p>
          {/*TODO: implement last used*/}
          {/*<p className='mt-1 text-sm text-purple-500'>Last used {connection.lastUsedAt}</p>*/}
        </CardContent>
      </Card>
      {connection.applicationName !== 'Interledger Wallet Managed' && (
        <OutlineButton
          // TODO error token colors
          className='text-red-700 outline-red-700 focus-visible:outline-red-800'
          form='key-id'
          type='submit'
        >
          Delete
        </OutlineButton>
      )}
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

  return redirectWithSnackbar(request, href('/settings/keys'), {
    message: 'Public key was deleted.',
    icon: 'close'
  })
}
