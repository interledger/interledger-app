import { Code } from '@bufbuild/connect'
import type {
  ActionFunctionArgs,
  LoaderFunctionArgs,
  MetaFunction
} from '@remix-run/node'
import { redirect } from '@remix-run/node'
import {
  Form,
  useActionData,
  useLoaderData,
  useNavigation
} from '@remix-run/react'
import { useEffect, useState } from 'react'
import { route } from 'routes-gen'
import type { ApplicationProps } from '~/components'
import { Button, Card, CardContent, Layouts, TextField } from '~/components'
import { jsonWithCSRF, validateCSRFToken } from '~/lib/csrf.server'
import { isConnectError } from '~/lib/error.server'
import { grpc } from '~/lib/grpc.server'
import { mergeMeta } from '~/lib/meta'
import { redirectWithSnackbar } from '~/lib/snackbar.server'
import { useScaffoldStore } from '~/lib/useScaffoldStore'

export async function loader({ request }: LoaderFunctionArgs) {
  const balancesResponse = await grpc.getBalances(request, {})
  if (
    isConnectError(balancesResponse) ||
    balancesResponse.balances.filter((bal) => bal.countryCode == 'CAD')
      .length == 0
  )
    throw redirect(route('/'))

  return jsonWithCSRF(request, {})
}

export const handle: ApplicationProps = {
  layout: Layouts.Focus,
  scaffold: {
    header: {
      back: route('/accounts'),
      title: 'Connect interac account'
    }
  }
}

export const meta: MetaFunction = mergeMeta(() => [
  {
    title: 'Connect interac account'
  }
])

export default function Page() {
  const { csrfToken } = useLoaderData<typeof loader>()

  const navigation = useNavigation()
  const actionData = useActionData<typeof action>()
  const [email, setEmail] = useState<string>('')

  const [setLoading] = useScaffoldStore((state) => [state.setLoading])

  useEffect(() => {
    if (navigation.state == 'submitting' && navigation.formMethod === 'post') {
      setLoading(true)
    } else if (navigation.state == 'loading' || navigation.state == 'idle') {
      setLoading(false)
    }
    // This ensures that loading is false when this route is unmounted.
    return () => setLoading(false)
  }, [navigation.formMethod, navigation.state, setLoading])

  return (
    <>
      <Form
        id='connect-interc'
        action={route('/connect/interac')}
        method='post'
        className='hidden'
      />
      <input
        form='connect-interac'
        value={csrfToken}
        name='csrfToken'
        type='hidden'
      />
      <Card>
        <CardContent>
          <p>
            Connect an interac account to easily withdraw from your balance.
          </p>
        </CardContent>
        <TextField
          id='email'
          form='connect-interac'
          label='Email'
          name='email'
          type='email'
          className='mt-2'
          value={email}
          onChange={(e) => {
            setEmail(e.target.value)
          }}
          aria-invalid={Boolean(actionData?.errors?.email) || undefined}
          aria-describedby={
            actionData?.errors?.email ? 'email-error' : undefined
          }
          required
          errorMessage={actionData?.errors?.email}
        />
      </Card>
      <Button type='submit' form='connect-interac'>
        Continue
      </Button>
    </>
  )
}

export async function action({ request }: ActionFunctionArgs) {
  const form = await request.formData()
  const email = String(form.get('accountNumber') || '')

  await validateCSRFToken(request, form)

  const errors = {
    form: '',
    email: ''
  }

  let response = await grpc.setChimoneyInterlocEmail(request, { email })
  if (isConnectError(response)) {
    if (response.code == Code.InvalidArgument) {
      return response.error({ errors })
    } else {
      errors.form = 'Failed to connect interac account.'
      return response.error({ errors }, {}, { action: 'Contact support' })
    }
  }

  return redirectWithSnackbar(request, route('/accounts'), {
    message: 'New interac account successfully saved.',
    icon: 'close'
  })
}
