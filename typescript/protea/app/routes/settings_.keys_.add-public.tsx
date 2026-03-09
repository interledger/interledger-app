import { Code } from '@bufbuild/connect'
import type {
  ActionFunctionArgs,
  LoaderFunctionArgs,
  MetaFunction
} from '@remix-run/node'
import { Form, useActionData, useLoaderData } from '@remix-run/react'
import { route } from 'routes-gen'
import type { ApplicationProps } from '~/components'
import logger from '~/lib/logger.server'
import {
  Button,
  Card,
  CardContent,
  Layouts,
  TextArea,
  TextField
} from '~/components'
import { jsonWithCSRF, validateCSRFToken } from '~/lib/csrf.server'
import { isConnectError } from '~/lib/error.server'
import { grpc } from '~/lib/grpc.server'
import { mergeMeta } from '~/lib/meta'
import { redirectWithSnackbar } from '~/lib/snackbar.server'

export const handle: ApplicationProps = {
  layout: Layouts.Focus,
  scaffold: {
    header: {
      back: route('/settings/keys'),
      title: 'Add a public key'
    }
  }
}

export const meta: MetaFunction = mergeMeta(() => [
  {
    title: 'Add a public key'
  }
])

export async function loader({ request }: LoaderFunctionArgs) {
  return jsonWithCSRF(request, {})
}

export default function Page() {
  const actionData = useActionData<typeof action>()
  const { csrfToken } = useLoaderData<typeof loader>()

  return (
    <>
      <Form
        id='add-public-key'
        action='/settings/keys/add-public'
        method='post'
        className='hidden'
      />
      <input
        form='add-public-key'
        value={csrfToken}
        name='csrfToken'
        type='hidden'
      />
      <Card>
        <CardContent>
          <p>Add a public key to connect to your wallet.</p>
        </CardContent>
        <TextField
          id='applicationName'
          label='Nickname'
          name='applicationName'
          form='add-public-key'
          type='text'
          defaultValue=''
          className='mt-2'
          aria-invalid={
            Boolean(actionData?.errors.applicationName) || undefined
          }
          aria-describedby={
            actionData?.errors.applicationName
              ? 'applicationName-error'
              : undefined
          }
          required
          errorMessage={actionData?.errors.applicationName}
        />

        <TextArea
          id='publicKey'
          form='add-public-key'
          label='Public key'
          name='publicKey'
          className='mt-4'
          placeholder='ewogICJrdHkiOiAiT0tQIiwKICAiY3J2IjogIkVkMjU1MTkiLAogICJraWQiOiAidGVzdC1rZXktZWQyNTUxOSIsCiAgImQiOiAibjROaS1IcElTcFZPYm5RTVcwd09oQ0tST2FJS3FLdFdfMlpZYjJwOUtjVSIsCiAgIngiOiAiSnJRTGo1UF84OWlYRVM5LXZGZ3JJeTI5Y2xGOUNDX29QUHN3M2M1RDBicyIKfQ=='
          aria-invalid={Boolean(actionData?.errors.publicKey) || undefined}
          aria-describedby={
            actionData?.errors.publicKey ? 'publicKey-error' : undefined
          }
          required
          errorMessage={actionData?.errors.publicKey}
        />
      </Card>

      <Button form='add-public-key' type='submit'>
        Save
      </Button>
    </>
  )
}

export async function action({ request }: ActionFunctionArgs) {
  const form = await request.formData()

  await validateCSRFToken(request, form)

  const errors = {
    form: '',
    applicationName: '',
    publicKey: ''
  }

  const response = await grpc.createConnection(request, {
    applicationName: form.get('applicationName') as string,
    publicKey: form.get('publicKey') as string
  })

  if (isConnectError(response)) {
    logger.error({ response }, 'Failed to create connection')
    if (response.code == Code.InvalidArgument) {
      return response.error({ errors })
    } else return response.error({ errors }, {}, { action: 'Contact support' })
  }

  return redirectWithSnackbar(request, route('/settings/keys'), {
    message: 'Public key added successfully.',
    icon: 'close'
  })
}
