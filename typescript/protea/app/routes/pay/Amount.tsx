import { useFetcher, useLoaderData } from '@remix-run/react'
import type { ChangeEventHandler } from 'react'
import { useCallback, useEffect, useState } from 'react'
import type { SelectOptions } from '~/components'
import {
  Button,
  Card,
  CardButton,
  CardContent,
  CardHeader,
  CardLink,
  Chip,
  ChipColor,
  Dialog,
  FynbosIcon,
  Icon,
  LinkedInIcon,
  Select,
  TextButton,
  TextField,
  TwitterIcon
} from '~/components'
import { Label } from '~/components/Label'
import { PayStep, usePayStore } from '~/lib/usePayStore'
import type { FormattedLinkedAccount } from '~/lib/wallet.server'
import type { loader, updatePaymentAction } from './route'

export function AmountWithOpenPayments() {
  const { csrfToken } = useLoaderData<typeof loader>()

  const accounts = useFetcher()
  const quoteFetcher = useFetcher()
  const publicWalletInfoFetcher = useFetcher()

  const [showDialog, setShowDialog] = useState<boolean>(false)

  const [
    amount,
    address,
    account,
    displayAmount,
    note,
    publicWalletInfo,
    setAmount,
    setStep,
    setNote,
    setQuote,
    setAccount,
    setPublicWalletInfo
  ] = usePayStore((state) => [
    state.amount,
    state.address,
    state.account,
    state.displayAmount,
    state.note,
    state.publicWalletInfo,
    state.setAmount,
    state.setStep,
    state.setNote,
    state.setQuote,
    state.setAccount,
    state.setPublicWalletInfo
  ])

  useEffect(() => {
    if (accounts.state == 'idle' && accounts.data == null) {
      accounts.load(`/pay?accounts=test`)
    }
  }, [accounts])

  useEffect(() => {
    if (
      !publicWalletInfo &&
      address &&
      publicWalletInfoFetcher.state == 'idle'
    ) {
      publicWalletInfoFetcher.load(`/pay?walletUrl=${address.walletUrl}`)
    }
  }, [address, publicWalletInfo, publicWalletInfoFetcher])

  const _onChangeAmount = useCallback<ChangeEventHandler<HTMLInputElement>>(
    (event) => {
      let amount = event.target.value
      setAmount(amount)
      quoteFetcher.submit(
        {
          formName: 'quote',
          amount,
          accountId: account?.id as string,
          walletUrl: address?.walletUrl as string,
          identity: address?.identifier as string,
          identityType: address?.identifierType as string,
          note,
          csrfToken
        },
        { method: 'post' }
      )
    },
    [
      account?.id,
      address?.identifier,
      address?.identifierType,
      address?.walletUrl,
      csrfToken,
      note,
      quoteFetcher,
      setAmount
    ]
  )

  const _onChangeLinkedAccount = useCallback(
    (event: SelectOptions) => {
      const accountId = event.id
      setAccount(event as FormattedLinkedAccount)
      quoteFetcher.submit(
        {
          formName: 'quote',
          amount,
          accountId: accountId,
          walletUrl: address?.walletUrl as string,
          identity: address?.identifier as string,
          identityType: address?.identifierType as string,
          note,
          csrfToken
        },
        { method: 'post' }
      )
    },
    [
      address?.identifier,
      address?.identifierType,
      address?.walletUrl,
      amount,
      csrfToken,
      note,
      quoteFetcher,
      setAccount
    ]
  )

  const _onChangeNote = useCallback<ChangeEventHandler<HTMLInputElement>>(
    (event) => {
      let note = event.target.value
      setNote(note)
      quoteFetcher.submit(
        {
          formName: 'quote',
          amount,
          accountId: account?.id as string,
          walletUrl: address?.walletUrl as string,
          identity: address?.identifier as string,
          identityType: address?.identifierType as string,
          note,
          csrfToken
        },
        { method: 'post' }
      )
    },
    [
      account?.id,
      address?.identifier,
      address?.identifierType,
      address?.walletUrl,
      amount,
      csrfToken,
      quoteFetcher,
      setNote
    ]
  )

  useEffect(() => {
    if (quoteFetcher.data?.quoteId) {
      setQuote(quoteFetcher.data?.quoteId, quoteFetcher.data?.requiresOTP)
    }
    if (
      quoteFetcher.data?.type == 'submitting' &&
      quoteFetcher.state == 'idle'
    ) {
      setStep(PayStep.CONFIRM)
    }
  }, [
    quoteFetcher,
    quoteFetcher.data?.quoteId,
    quoteFetcher.data?.type,
    setQuote,
    setStep
  ])

  useEffect(() => {
    if (accounts.data?.sendAccounts && !account) {
      setAccount({ ...accounts.data?.sendAccounts[0] })
    }
  }, [account, accounts.data?.sendAccounts, setAccount])

  useEffect(() => {
    if (publicWalletInfoFetcher.data?.publicWalletInfo) {
      setPublicWalletInfo(publicWalletInfoFetcher.data.publicWalletInfo)
    }
  }, [publicWalletInfoFetcher.data?.publicWalletInfo, setPublicWalletInfo])

  const _onClick = useCallback<{
    (): void
  }>(() => {
    quoteFetcher.submit(
      {
        formName: 'quote',
        amount,
        accountId: account?.id as string,
        walletUrl: address?.walletUrl as string,
        note,
        identity: address?.identifier as string,
        identityType: address?.identifierType as string,
        type: 'submitting',
        csrfToken
      },
      { method: 'post' }
    )
  }, [
    account?.id,
    address?.identifier,
    address?.identifierType,
    address?.walletUrl,
    amount,
    csrfToken,
    note,
    quoteFetcher
  ])

  return (
    <>
      <quoteFetcher.Form
        id='amount-form'
        action='/pay'
        method='post'
        className='hidden'
      />
      <Card>
        <CardContent>
          <div className='flex items-center justify-between'>
            <h2 className='text-4xl font-medium text-strong'>
              {displayAmount}
            </h2>
            {address?.identifierType === 'wallet' && (
              <FynbosIcon height='h-12' />
            )}
            {address?.identifierType === 'twitter' && (
              <TwitterIcon height='h-12' />
            )}
          </div>
        </CardContent>
        <Label className='mt-2'>Payment to</Label>
        <CardButton noHover onClick={() => setShowDialog(true)}>
          <div className='flex w-full items-center justify-between text-medium'>
            <span>{address?.identifier}</span>
            <Icon>navigate_next</Icon>
          </div>
        </CardButton>
      </Card>
      <Card>
        <TextField
          id='amount'
          label='Amount'
          name='amount'
          defaultValue={amount}
          onChange={_onChangeAmount}
          prefix='$'
          type='number'
          min='0'
          step='0.01'
          aria-invalid={Boolean(quoteFetcher.data?.errors?.amount) || undefined}
          aria-describedby={
            quoteFetcher.data?.errors?.amount ? 'amount-error' : undefined
          }
          errorMessage={quoteFetcher.data?.errors?.amount || undefined}
          required
        />
      </Card>
      <Card>
        <CardContent>
          <span>Select an account to pay from:</span>
        </CardContent>
        <Select
          id='linkedAccount'
          label='Connected accounts'
          className='mt-2'
          value={account as SelectOptions}
          options={accounts.data?.sendAccounts || []}
          onChange={_onChangeLinkedAccount}
          aria-invalid={
            Boolean(quoteFetcher.data?.errors?.linkedAccount) || undefined
          }
          aria-describedby={
            quoteFetcher.data?.errors?.linkedAccount
              ? 'linkedAccount-error'
              : undefined
          }
          errorMessage={quoteFetcher.data?.errors?.linkedAccount}
        />
        <TextField
          id='note'
          label='Note'
          name='note'
          type='text'
          defaultValue={note}
          onChange={_onChangeNote}
          className='mt-4'
          aria-invalid={Boolean(quoteFetcher.data?.errors?.note) || undefined}
          aria-describedby={
            quoteFetcher.data?.errors?.note ? 'reference-error' : undefined
          }
          errorMessage={quoteFetcher.data?.errors?.note}
        />
      </Card>
      <Button type='button' onClick={_onClick}>
        Continue
      </Button>
      <Dialog open={showDialog} setOpen={setShowDialog}>
        <CardHeader>
          <h1 className='text-xl font-medium'>User information</h1>
        </CardHeader>
        <CardContent>
          <span className='text-medium'>
            You are viewing public information about the person you intend to
            pay.
          </span>
        </CardContent>
        <Label className='mt-4'>Public name</Label>
        <div className='mt-1 flex rounded-xl bg-nav p-3 text-medium'>
          <span className=''>{publicWalletInfo?.publicName}</span>
        </div>
        <Label className='mt-2'>Wallet address</Label>
        <CardLink className='flex w-full' to={publicWalletInfo?.address ?? ''}>
          <div className='flex w-full items-center justify-between text-medium'>
            <div className='flex space-x-2'>
              <FynbosIcon />
              <span>{publicWalletInfo?.shortAddress}</span>
            </div>
            <Icon>navigate_next</Icon>
          </div>
        </CardLink>
        {publicWalletInfo?.identities.map((identity) => (
          <div key={identity.id} className='contents'>
            <Label className='mt-2 capitalize'>{identity.platform}</Label>
            <CardLink className='flex w-full' to={publicWalletInfo.address}>
              <div className='flex w-full items-center justify-between text-medium'>
                <div className='flex space-x-2'>
                  {identity.platform == 'twitter' && <TwitterIcon />}
                  {identity.platform == 'linkedin' && <LinkedInIcon />}
                  <span>{identity.identifier}</span>
                </div>
                {identity.state == 'verified' && (
                  <Chip color={ChipColor.green}>Verified</Chip>
                )}
              </div>
            </CardLink>
          </div>
        ))}

        <CardContent className='flex w-full justify-end space-x-6'>
          <TextButton type='button' onClick={() => setShowDialog(false)}>
            Close
          </TextButton>
        </CardContent>
      </Dialog>
    </>
  )
}

export const Amount = () => {
  const { csrfToken } = useLoaderData<typeof loader>()

  const accounts = useFetcher<typeof loader>()
  const paymentFetcher = useFetcher<typeof updatePaymentAction>()
  const publicWalletInfoFetcher = useFetcher<typeof loader>()

  const [showDialog, setShowDialog] = useState<boolean>(false)

  const [
    paymentId,
    amount,
    address,
    account,
    displayAmount,
    note,
    publicWalletInfo,
    setAmount,
    setStep,
    setNote,
    setAccount,
    setPublicWalletInfo,
    setPayment
  ] = usePayStore((state) => [
    state.paymentId,
    state.amount,
    state.address,
    state.account,
    state.displayAmount,
    state.note,
    state.publicWalletInfo,
    state.setAmount,
    state.setStep,
    state.setNote,
    state.setAccount,
    state.setPublicWalletInfo,
    state.setPayment
  ])

  const _onChangeAmount = useCallback<ChangeEventHandler<HTMLInputElement>>(
    (event) => {
      let amount = event.target.value
      setAmount(amount)
      paymentFetcher.submit(
        {
          formName: 'updatePayment',
          paymentId,
          amount,
          accountId: account?.id as string,
          identity: address?.identifier as string,
          identityType: address?.identifierType as string,
          walletUrl: address?.walletUrl as string,
          csrfToken
        },
        { method: 'post' }
      )
    },
    [
      paymentFetcher,
      setAmount,
      account?.id,
      address?.identifier,
      address?.identifierType,
      address?.walletUrl,
      paymentId,
      csrfToken
    ]
  )

  const _onChangeLinkedAccount = useCallback(
    (event: SelectOptions) => {
      const accountId = event.id
      setAccount(event as FormattedLinkedAccount)
      paymentFetcher.submit(
        {
          formName: 'updatePayment',
          paymentId,
          amount,
          accountId: accountId,
          identity: address?.identifier as string,
          identityType: address?.identifierType as string,
          walletUrl: address?.walletUrl as string,
          csrfToken
        },
        { method: 'post' }
      )
    },
    [
      csrfToken,
      paymentFetcher,
      setAccount,
      address?.identifier,
      address?.identifierType,
      address?.walletUrl,
      amount,
      paymentId
    ]
  )

  const _onChangeNote = useCallback<ChangeEventHandler<HTMLInputElement>>(
    (event) => {
      let note = event.target.value
      setNote(note)
      paymentFetcher.submit(
        {
          formName: 'updatePayment',
          paymentId,
          note,
          amount,
          accountId: account?.id as string,
          identity: address?.identifier as string,
          identityType: address?.identifierType as string,
          walletUrl: address?.walletUrl as string,
          csrfToken
        },
        { method: 'post' }
      )
    },
    [
      csrfToken,
      paymentFetcher,
      setNote,
      account?.id,
      address?.identifier,
      address?.identifierType,
      address?.walletUrl,
      amount,
      paymentId
    ]
  )

  useEffect(() => {
    if (accounts.state == 'idle' && accounts.data == null) {
      accounts.load(`/pay?accounts=test`)
    }
  }, [accounts])

  useEffect(() => {
    if (
      !publicWalletInfo &&
      address &&
      publicWalletInfoFetcher.state == 'idle'
    ) {
      publicWalletInfoFetcher.load(`/pay?walletUrl=${address.walletUrl}`)
    }
  }, [address, publicWalletInfo, publicWalletInfoFetcher])

  useEffect(() => {
    if (accounts.data?.sendAccounts && !account) {
      setAccount({ ...accounts.data?.sendAccounts[0] })
    }
  }, [account, accounts.data?.sendAccounts, setAccount])

  useEffect(() => {
    if (publicWalletInfoFetcher.data?.publicWalletInfo) {
      setPublicWalletInfo(publicWalletInfoFetcher.data.publicWalletInfo)
    }
  }, [publicWalletInfoFetcher.data?.publicWalletInfo, setPublicWalletInfo])

  const _onClick = useCallback<{
    (): void
  }>(() => {
    paymentFetcher.submit(
      {
        formName: 'updatePayment',
        amount,
        accountId: account?.id as string,
        note,
        identity: address?.identifier as string,
        identityType: address?.identifierType as string,
        walletUrl: address?.walletUrl as string,
        type: 'submitting',
        csrfToken
      },
      { method: 'post' }
    )
  }, [
    account?.id,
    address?.identifier,
    address?.identifierType,
    address?.walletUrl,
    amount,
    csrfToken,
    note,
    paymentFetcher
  ])

  useEffect(() => {
    if (paymentFetcher.data?.payment) {
      let requiredActions = paymentFetcher.data.payment
        ?.requiredActions as Array<number>
      const requiredActionOTP = 7
      setPayment(
        paymentFetcher.data?.payment?.id,
        requiredActions.includes(requiredActionOTP)
      )
    }
    if (
      paymentFetcher.data?.type == 'submitting' &&
      paymentFetcher.state == 'idle'
    ) {
      setStep(PayStep.CONFIRM)
    }
  }, [
    paymentFetcher,
    paymentFetcher.data?.payment?.id,
    paymentFetcher.data?.type,
    setPayment,
    setStep
  ])

  return (
    <>
      <paymentFetcher.Form
        id='amount-form'
        action='/pay'
        method='post'
        className='hidden'
      />
      <Card>
        <CardContent>
          <div className='flex items-center justify-between'>
            <h2 className='text-4xl font-medium text-strong'>
              {displayAmount}
            </h2>
            {address?.identifierType === 'wallet' && (
              <FynbosIcon height='h-12' />
            )}
            {address?.identifierType === 'twitter' && (
              <TwitterIcon height='h-12' />
            )}
          </div>
        </CardContent>
        <Label className='mt-2'>Payment to</Label>
        <CardButton noHover onClick={() => setShowDialog(true)}>
          <div className='flex w-full items-center justify-between text-medium'>
            <span>{address?.identifier}</span>
            <Icon>navigate_next</Icon>
          </div>
        </CardButton>
      </Card>
      <Card>
        <TextField
          id='amount'
          label='Amount'
          name='amount'
          defaultValue={amount}
          onChange={_onChangeAmount}
          prefix='$'
          type='number'
          min='0'
          step='0.01'
          aria-invalid={
            Boolean(paymentFetcher.data?.errors?.amount) || undefined
          }
          aria-describedby={
            paymentFetcher.data?.errors?.amount ? 'amount-error' : undefined
          }
          errorMessage={paymentFetcher.data?.errors?.amount || undefined}
          required
        />
      </Card>
      <Card>
        <CardContent>
          <span>Select an account to pay from:</span>
        </CardContent>
        <Select
          id='linkedAccount'
          label='Connected accounts'
          className='mt-2'
          value={account as SelectOptions}
          options={accounts.data?.sendAccounts || []}
          onChange={_onChangeLinkedAccount}
          aria-invalid={
            Boolean(paymentFetcher.data?.errors?.linkedAccount) || undefined
          }
          aria-describedby={
            paymentFetcher.data?.errors?.linkedAccount
              ? 'linkedAccount-error'
              : undefined
          }
          errorMessage={paymentFetcher.data?.errors?.linkedAccount}
        />
        <TextField
          id='note'
          label='Note'
          name='note'
          type='text'
          defaultValue={note}
          onChange={_onChangeNote}
          className='mt-4'
          aria-invalid={Boolean(paymentFetcher.data?.errors?.note) || undefined}
          aria-describedby={
            paymentFetcher.data?.errors?.note ? 'reference-error' : undefined
          }
          errorMessage={paymentFetcher.data?.errors?.note}
        />
      </Card>
      <Button type='button' onClick={_onClick}>
        Continue
      </Button>
      <Dialog open={showDialog} setOpen={setShowDialog}>
        <CardHeader>
          <h1 className='text-xl font-medium'>User information</h1>
        </CardHeader>
        <CardContent>
          <span className='text-medium'>
            You are viewing public information about the person you intend to
            pay.
          </span>
        </CardContent>
        <Label className='mt-4'>Public name</Label>
        <div className='mt-1 flex rounded-xl bg-nav p-3 text-medium'>
          <span className=''>{publicWalletInfo?.publicName}</span>
        </div>
        <Label className='mt-2'>Wallet address</Label>
        <CardLink className='flex w-full' to={publicWalletInfo?.address ?? ''}>
          <div className='flex w-full items-center justify-between text-medium'>
            <div className='flex space-x-2'>
              <FynbosIcon />
              <span>{publicWalletInfo?.shortAddress}</span>
            </div>
            <Icon>navigate_next</Icon>
          </div>
        </CardLink>
        {publicWalletInfo?.identities.map((identity) => (
          <div key={identity.id} className='contents'>
            <Label className='mt-2 capitalize'>{identity.platform}</Label>
            <CardLink className='flex w-full' to={publicWalletInfo.address}>
              <div className='flex w-full items-center justify-between text-medium'>
                <div className='flex space-x-2'>
                  {identity.platform == 'twitter' && <TwitterIcon />}
                  {identity.platform == 'linkedin' && <LinkedInIcon />}
                  <span>{identity.identifier}</span>
                </div>
                {identity.state == 'verified' && (
                  <Chip color={ChipColor.green}>Verified</Chip>
                )}
              </div>
            </CardLink>
          </div>
        ))}

        <CardContent className='flex w-full justify-end space-x-6'>
          <TextButton type='button' onClick={() => setShowDialog(false)}>
            Close
          </TextButton>
        </CardContent>
      </Dialog>
    </>
  )
}
