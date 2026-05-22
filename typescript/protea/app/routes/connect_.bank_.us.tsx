// ORPHANED FROM UI — as of Phase 3, Plaid is the only bank-link path on Home.
// Still reachable by direct URL and still driven by e2e (e2e/pti_deposit.go).
// Full removal tracked in BACKLOG-Bx.
import type { Route } from './+types/connect_.bank_.us'
import { Code } from '@bufbuild/connect'
import { redirect } from 'react-router';
import { Form, useActionData, useLoaderData, useNavigation } from 'react-router';
import { useEffect, useState } from 'react'
import { href } from 'react-router'
import type { ApplicationProps, SelectOptions } from '~/components'
import {
  Button,
  Card,
  CardContent,
  Layouts,
  Select,
  TextField
} from '~/components'
import { jsonWithCSRF, validateCSRFToken } from '~/lib/csrf.server'
import { isConnectError } from '~/lib/error.server'
import { grpc } from '~/lib/grpc.server'
import { mergeMeta } from '~/lib/meta'
import { redirectWithSnackbar } from '~/lib/snackbar.server'
import { useScaffoldStore } from '~/lib/useScaffoldStore'

export async function loader({ request }: Route.LoaderArgs) {
  const balancesResponse = await grpc.getBalances(request, {})
  if (
    isConnectError(balancesResponse) ||
    balancesResponse.balances.filter((bal) => bal.countryCode == 'US').length ==
      0
  )
    throw redirect(href('/'))

  const accountTypes: SelectOptions[] = [
    { id: 'CHECKING', name: 'Checking' },
    { id: 'SAVINGS', name: 'Savings' }
  ]

  return jsonWithCSRF(request, { accountTypes })
}

export const handle: ApplicationProps = {
  layout: Layouts.Focus,
  scaffold: {
    header: {
      back: href('/accounts'),
      title: 'Connect bank account'
    }
  }
}

export const meta = mergeMeta(() => [
  {
    title: 'Connect bank account'
  }
])

export default function Page() {
  const { accountTypes, csrfToken } = useLoaderData()

  const navigation = useNavigation()
  const actionData = useActionData()
  const [accountType, setAccountType] = useState<SelectOptions>(accountTypes[0])

  const [setLoading] = useScaffoldStore((state) => [state.setLoading])

  useEffect(() => {
    if (navigation.state == 'submitting' && navigation.formMethod === 'POST') {
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
        id='connect-bank-us'
        action={href('/connect/bank/us')}
        method='post'
        className='hidden'
      />
      <input
        form='connect-bank-us'
        value={csrfToken}
        name='csrfToken'
        type='hidden'
      />
      <input
        form='connect-bank-us'
        value={accountType.id}
        name='accountType'
        type='hidden'
      />
      <Card>
        <CardContent>
          <p>Connect a bank account.</p>
        </CardContent>
        <TextField
          id='bankName'
          form='connect-bank-us'
          label='Bank Name'
          name='bankName'
          type='text'
          className='mt-2'
          aria-invalid={Boolean(actionData?.errors?.bankName) || undefined}
          aria-describedby={
            actionData?.errors?.bankName ? 'bankName-error' : undefined
          }
          required
          errorMessage={actionData?.errors?.bankName}
        />
        <TextField
          id='accountNumber'
          form='connect-bank-us'
          label='Account number'
          name='accountNumber'
          type='text'
          className='mt-2'
          aria-invalid={Boolean(actionData?.errors?.accountNumber) || undefined}
          aria-describedby={
            actionData?.errors?.accountNumber
              ? 'accountNumber-error'
              : undefined
          }
          required
          errorMessage={actionData?.errors?.accountNumber}
        />
        <TextField
          id='routingNumber'
          form='connect-bank-us'
          label='Routing number'
          name='routingNumber'
          type='text'
          className='mt-2'
          aria-invalid={Boolean(actionData?.errors?.routingNumber) || undefined}
          aria-describedby={
            actionData?.errors?.routingNumber
              ? 'routingNumber-error'
              : undefined
          }
          required
          errorMessage={actionData?.errors?.routingNumber}
        />
        <Select
          id='accountType'
          label='Account Type'
          value={accountType}
          options={accountTypes}
          onChange={setAccountType}
          className='mt-4'
        />
      </Card>
      <Button type='submit' form='connect-bank-us'>
        Continue
      </Button>
    </>
  )
}

export async function action({ request }: Route.ActionArgs) {
  const form = await request.formData()
  const accountType = form.get('accountType') as string
  const accountNumber = form.get('accountNumber') as string
  const routingNumber = form.get('routingNumber') as string
  const bankName = form.get('bankName') as string

  await validateCSRFToken(request, form)

  const errors = {
    form: '',
    accountNumber: '',
    accountTyoe: '',
    routingNumber: '',
    bankName: ''
  }

  let response = await grpc.createPtiBankAccount(request, {
    accountNumber,
    accountType,
    routingNumber,
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

  return redirectWithSnackbar(request, href('/accounts'), {
    message: 'New bank account successfully saved.',
    icon: 'close'
  })
}
