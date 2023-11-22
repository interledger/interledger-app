import { Code } from '@bufbuild/connect'
import type {
  ActionFunctionArgs,
  LoaderFunctionArgs,
  MetaFunction
} from '@remix-run/node'
import { redirect } from '@remix-run/node'
import { Form, useActionData, useLoaderData } from '@remix-run/react'
import { route } from 'routes-gen'
import type { ApplicationProps } from '~/components'
import { Button, Card, CardContent, Layouts, TextField } from '~/components'
import { jsonWithCSRF, validateCSRFToken } from '~/lib/csrf.server'
import { isConnectError } from '~/lib/error.server'
import { grpc } from '~/lib/grpc.server'
import { mergeMeta } from '~/lib/meta'
import { redirectWithSnackbar } from '~/lib/snackbar.server'

export async function loader({ request }: LoaderFunctionArgs) {
  const getXagoBalancesResponse = await grpc.getXagoBalances(request, {})
  if (isConnectError(getXagoBalancesResponse)) throw redirect(route('/'))

  if (getXagoBalancesResponse.balances.length == 0) {
    throw redirect(route('/'))
  }

  return jsonWithCSRF(request, {})
}

export const handle: ApplicationProps = {
  layout: Layouts.Focus,
  scaffold: {
    header: {
      back: route('/accounts'),
      title: 'Connect bank account'
    }
  }
}

export const meta: MetaFunction = mergeMeta(() => [
  {
    title: 'Connect bank account'
  }
])

export default function Page() {
  const { csrfToken } = useLoaderData<typeof loader>()
  const actionData = useActionData<typeof action>()

  return (
    <>
      <Form
        id='connect-bank-za'
        action={route('/connect/bank/za')}
        method='post'
        className='hidden'
      />
      <input
        form='connect-bank-za'
        value={csrfToken}
        name='csrfToken'
        type='hidden'
      />
      <Card>
        <CardContent>
          <p>Connect a bank account to easily send and receive payments.</p>
        </CardContent>
        <TextField
          id='accountNumber'
          form='connect-bank-za'
          label='Account number'
          name='accountNumber'
          type='text'
          className='mt-2'
          aria-invalid={Boolean(actionData?.errors?.accountNumber) || undefined}
          aria-describedby={
            actionData?.errors?.accountNumber ? 'firstName-error' : undefined
          }
          required
          errorMessage={actionData?.errors?.accountNumber}
        />
        <TextField
          id='bankName'
          form='connect-bank-za'
          label='Bank name'
          name='bankName'
          type='text'
          className='mt-4'
          aria-invalid={Boolean(actionData?.errors?.bankName) || undefined}
          aria-describedby={
            actionData?.errors?.bankName ? 'firstName-error' : undefined
          }
          required
          errorMessage={actionData?.errors?.bankName}
        />
        <TextField
          id='branchCode'
          form='connect-bank-za'
          label='Branch code'
          name='branchCode'
          type='text'
          className='mt-4'
          aria-invalid={Boolean(actionData?.errors?.branchCode) || undefined}
          aria-describedby={
            actionData?.errors?.branchCode ? 'firstName-error' : undefined
          }
          required
          errorMessage={actionData?.errors?.branchCode}
        />
      </Card>
      <Button type='submit' form='connect-bank-za'>
        Continue
      </Button>
    </>
  )
}

export async function action({ request }: ActionFunctionArgs) {
  const form = await request.formData()
  const accountNumber = form.get('accountNumber') as string
  const branchCode = form.get('branchCode') as string
  const bankName = form.get('bankName') as string

  await validateCSRFToken(request, form)

  const errors = {
    form: '',
    accountNumber: '',
    branchCode: '',
    bankName: ''
  }

  let response = await grpc.addXagoBankAccount(request, {
    accountNumber,
    branchCode,
    bankName
  })

  if (isConnectError(response)) {
    if (response.code == Code.InvalidArgument) {
      return response.error({ errors })
    } else {
      errors.form = 'Failed to connect bank account.'
      return response.error({ errors }, {}, { action: 'Contact support' })
    }
  }

  return redirectWithSnackbar(request, route('/accounts'), {
    message: 'New bank account successfully saved.',
    icon: 'close'
  })
}
