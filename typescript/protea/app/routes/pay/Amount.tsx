import { useFetcher } from '@remix-run/react'
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

export function Amount() {
  // const { linkedAccounts } = useLoaderData<typeof loader>()

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
    setQuoteId,
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
    state.setQuoteId,
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
          note
        },
        { method: 'post' }
      )
    },
    [
      account?.id,
      address?.identifier,
      address?.identifierType,
      address?.walletUrl,
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
          note
        },
        { method: 'post' }
      )
    },
    [
      address?.identifier,
      address?.identifierType,
      address?.walletUrl,
      amount,
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
          note
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
      quoteFetcher,
      setNote
    ]
  )

  useEffect(() => {
    if (quoteFetcher.data?.quoteId) {
      setQuoteId(quoteFetcher.data?.quoteId)
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
    setQuoteId,
    setStep
  ])

  useEffect(() => {
    if (accounts.data?.sendAccounts) {
      setAccount({ ...accounts.data?.sendAccounts[0] })
    }
  }, [accounts.data?.sendAccounts, setAccount])

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
        type: 'submitting'
      },
      { method: 'post' }
    )
  }, [
    account?.id,
    address?.identifier,
    address?.identifierType,
    address?.walletUrl,
    amount,
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
        <CardButton onClick={() => setShowDialog(true)}>
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
          label='Reference'
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
