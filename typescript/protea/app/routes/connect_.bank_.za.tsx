import type { Route } from './+types/connect_.bank_.za'
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
    balancesResponse.balances.filter((bal) => bal.countryCode == 'ZA').length ==
      0
  )
    throw redirect(href('/'))

  const banks: SelectOptions[] = [
    { id: '632005', name: 'Absa Bank' },
    { id: '430000', name: 'African Bank Limited' },
    { id: '410506', name: 'Bank of Athens' },
    { id: '590000', name: 'Barclays Bank' },
    { id: '679000', name: 'Bidvest Bank' },
    { id: '470010', name: 'Capitec Bank Limited' },
    { id: '679000', name: 'Discovery Bank Limited' },
    { id: '250655', name: 'First National Bank' },
    { id: '201419', name: 'FirstRand Bank Limited' },
    { id: '587000', name: 'HSBC Bank' },
    { id: '580105', name: 'Investec Bank Limited' },
    { id: '450905', name: 'Mercantile Bank Limited' },
    { id: '198765', name: 'Nedbank' },
    { id: '462005', name: 'Old Mutual' },
    { id: '261251', name: 'Rand Merchant Bank ' },
    { id: '222026', name: 'RMB Private Bank' },
    { id: '683000', name: 'Sasfin Bank Limited' },
    { id: '460005', name: 'SA Post Bank (Post Office)' },
    { id: '410506', name: 'South African Bank of Athens Limited' },
    { id: '051001', name: 'Standard Bank' },
    { id: '730020', name: 'Standard Chartered Bank' },
    { id: '678910', name: 'Tyme Bank' }
  ]

  return jsonWithCSRF(request, { banks })
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
  const { banks, csrfToken } = useLoaderData()

  const navigation = useNavigation()
  const actionData = useActionData()
  const [bank, setBank] = useState<SelectOptions>(banks[0])

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
        id='connect-bank-za'
        action={href('/connect/bank/za')}
        method='post'
        className='hidden'
      />
      <input
        form='connect-bank-za'
        value={csrfToken}
        name='csrfToken'
        type='hidden'
      />
      <input
        form='connect-bank-za'
        value={bank.name}
        name='bankName'
        type='hidden'
      />
      <input
        form='connect-bank-za'
        value={bank.id}
        name='branchCode'
        type='hidden'
      />
      <Card>
        <CardContent>
          <p>Connect a bank account to easily withdraw from your balance.</p>
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
        <Select
          id='bank'
          label='Bank'
          value={bank}
          options={banks}
          onChange={setBank}
          className='mt-4'
        />
      </Card>
      <Button type='submit' form='connect-bank-za'>
        Continue
      </Button>
    </>
  )
}

export async function action({ request }: Route.ActionArgs) {
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

  return redirectWithSnackbar(request, href('/accounts'), {
    message: 'New bank account successfully saved.',
    icon: 'close'
  })
}
