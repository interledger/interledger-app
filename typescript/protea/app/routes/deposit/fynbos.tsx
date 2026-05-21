import { Form, useActionData, useLoaderData, useSearchParams } from 'react-router';
import type { ChangeEventHandler } from 'react'
import { useCallback, useEffect, useState } from 'react'
import { href } from 'react-router'
import {
  Alert,
  AlertBody,
  AlertContent,
  AlertTitle,
  Button,
  Card,
  CardContent,
  CardHeader,
  CardIcon,
  CardTitle,
  Icon,
  Router,
  Select,
  TextField
} from '~/components'
import type { FormattedLinkedAccount } from '~/data/accounts.server'

import { useScaffoldStore } from '~/lib/useScaffoldStore'
import { PaySelect } from '../pay_.$paymentId/PaySelect'
import { fynbosDepositLoader } from './loader.server';


export function FynbosDepositPage() {
  const { depositDetails, showTestDeposit } = useLoaderData<typeof fynbosDepositLoader>()

  const [setLoading] = useScaffoldStore((state) => [state.setLoading])

  useEffect(() => {
    // This ensures that loading is false when this route is unmounted.
    return () => {
      setLoading(false)
    }
  }, [setLoading])

  if (depositDetails) return <DepositDetails />

  return <Amount />
}

const Amount = () => {
  const { balance, balances, balanceAccount, linkedAccounts, csrfToken, provider } =
    useLoaderData<typeof fynbosDepositLoader>()
  const [, setSearchParams] = useSearchParams()
  const actionData = useActionData<any>()

  const [amount, setAmount] = useState<string>('')

  const [linkedAccount, setLinkedAccount] = useState<FormattedLinkedAccount>(
    linkedAccounts[0]
  )

  const _onChangeDepositAmount = useCallback<
    ChangeEventHandler<HTMLInputElement>
  >((event) => {
    setAmount(event.target.value)
  }, [])

  const _onChangeLinkedAccount = useCallback(
    (linkedAccount: FormattedLinkedAccount) => {
      setSearchParams(
        (prev) => {
          prev.set('linkedAccount', linkedAccount.id)
          return prev
        },
        { replace: true }
      )
    },
    [setSearchParams]
  )

  // TODO Need to make this dynamic based on jurisdiction capabilities
  if (linkedAccounts.length === 0)
    return (
      <Card>
        <CardContent>
          <div className='flex items-start space-x-4'>
            <CardIcon>
              <Icon>account_balance</Icon>
            </CardIcon>
            <div className='flex flex-col space-y-4'>
              <p className='text-sm text-medium'>
                To deposit from your balance, first connect a bank account.
              </p>
              <Router
                prefetch='render'
                className='text-sm font-medium text-primary'
                to={href('/accounts')}
              >
                Go to accounts page
              </Router>
            </div>
          </div>
        </CardContent>
      </Card>
    )

  return (
    <>
      <Form
        id='account-deposit'
        action={href('/deposit')}
        method='post'
        className='hidden'
      />
      <input
        form='account-deposit'
        value={csrfToken}
        name='csrfToken'
        type='hidden'
      />
      <input
        form='account-deposit'
        value={linkedAccount.id}
        name='fromLinkedAccount'
        type='hidden'
      />
      <input
        form='account-deposit'
        value={balanceAccount.linkedAccount}
        name='toLinkedAccount'
        type='hidden'
      />
      <input
        form='account-deposit'
        value={provider}
        name='provider'
        type='hidden'
      />
      <Card>
        <Select
          id='bank'
          label='Deposit to'
          value={balance}
          options={balances}
          onChange={_onChangeLinkedAccount}
        />
      </Card>
      <Card>
        <PaySelect
          id='depositAmount'
          label='Deposit amount'
          name='depositAmount'
          form='account-deposit'
          value={amount}
          onChange={_onChangeDepositAmount}
          linkedAccount={linkedAccount}
          linkedAccountOptions={linkedAccounts || []}
          onChangeLinkedAccount={setLinkedAccount}
          placeholder='0.00'
          prefixIcon={
            <div
              className={`flag:${linkedAccount.receiveCurrencyCountryCode}`}
            />
          }
          type='number'
          min='0'
          step='0.01'
          aria-invalid={Boolean(actionData?.errors?.depositAmount) || undefined}
          aria-describedby={
            actionData?.errors?.depositAmount ? 'amount-error' : undefined
          }
          errorMessage={actionData?.errors?.depositAmount || undefined}
        />
        <CardContent className='mt-2 flex flex-col gap-y-4'>
          <div className='flex flex-col gap-y-1'>
            <div className='flex w-full justify-between'>
              <span className='text-weak'>Fees</span>
              <span className='text-medium'>0.00</span>
            </div>
            <span className='text-xs text-weak'>
              For a limited time, the Interledger Wallet will absorb all fees.
            </span>
          </div>
        </CardContent>
      </Card>
      <Card>
        <TextField
          id='note'
          label='Deposit note'
          name='note'
          form='account-deposit'
          type='text'
          aria-invalid={Boolean(actionData?.errors?.note) || undefined}
          aria-describedby={
            actionData?.errors?.note ? 'reference-error' : undefined
          }
          errorMessage={actionData?.errors?.note}
        />
      </Card>
      <Button type='submit' form='account-deposit'>
        Continue
      </Button>
    </>
  )
}

export function DepositDetails() {
  const { depositDetails, csrfToken, showTestDeposit } =
    useLoaderData<typeof fynbosDepositLoader>()

  return (
    <>
      <Card>
        <CardHeader>
          <CardTitle>EFT details</CardTitle>
        </CardHeader>
        <CardContent>
          <span className='mt-4'>Arrives 1-2 business days.</span>
          <div className='mt-4 flex w-full flex-col justify-between space-y-1'>
            <span className='text-weak'>Bank</span>
            <span className='text-medium'>{depositDetails?.bankName}</span>
          </div>
          <div className='mt-4 flex w-full flex-col space-y-1'>
            <span className='text-weak'>Branch code</span>
            <span className='text-medium'>{depositDetails?.branchCode}</span>
          </div>
          <div className='mt-4 flex w-full flex-col space-y-1'>
            <span className='text-weak'>Account number</span>
            <span className='text-medium'>{depositDetails?.accountNumber}</span>
          </div>
          <div className='mt-4 flex w-full flex-col space-y-1'>
            <span className='text-weak'>Reference</span>
            <span className='text-medium'>
              {depositDetails?.depositReference}
            </span>
          </div>
        </CardContent>
      </Card>
      {showTestDeposit ? (
        <Form
          id='xago-test-account-deposit'
          action={href('/deposit')}
          method='post'
        >
          <input
            name='formName'
            value='xago-test-account-deposit'
            type='hidden'
          />
          <input value={csrfToken} name='csrfToken' type='hidden' />
          <Button type='submit'>Test Deposit</Button>
        </Form>
      ) : null}
      <Alert>
        <Icon>error</Icon>
        <AlertContent className='items-start'>
          <AlertTitle>Important</AlertTitle>
          <AlertBody>
            Use the reference above when depositing for secure and faster
            processing.
          </AlertBody>
        </AlertContent>
      </Alert>
    </>
  )
}

export function stringToBigInt(amount: string) {
  if (amount == '') return BigInt(0)
  const dotIndex = amount.lastIndexOf('.')
  if (dotIndex > -1) {
    const amounts = amount.split('.')
    return BigInt(amounts[0] + amounts[1].slice(0, 2).padEnd(2, '0'))
  }
  return BigInt(parseFloat(amount) * 100)
}

