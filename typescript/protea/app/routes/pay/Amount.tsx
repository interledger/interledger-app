import { useFetcher, useLoaderData } from '@remix-run/react'
import type { ChangeEventHandler } from 'react'
import { useCallback, useState } from 'react'
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
import type { loader } from '~/routes/pay/route'
import { useStore } from '~/store'

export function Amount() {
  const { linkedAccounts, publicWalletInfo } = useLoaderData<typeof loader>()
  const fetcher = useFetcher()
  const [showDialog, setShowDialog] = useState<boolean>(false)

  const [address, toLinkedAccountId] = useStore((state) => [
    state.address,
    state.toLinkedAccountId
  ])

  const [linkedAccount, setLinkedAccount] = useState<{
    id: string
    name: string
  }>(linkedAccounts[0])

  const _onChangeLinkedAccount = useCallback((event: SelectOptions) => {
    setLinkedAccount(event)
  }, [])

  const _onChangeInput = useCallback<ChangeEventHandler<HTMLInputElement>>(
    (event) => {
      let amount = event.target.value
      fetcher.submit({ amount: amount, toLinkedAccountId }, { method: 'post' })
    },
    [fetcher, toLinkedAccountId]
  )

  return (
    <>
      <fetcher.Form
        id='amount-form'
        action='/pay'
        method='post'
        className='hidden'
      />
      <Card>
        <CardContent>
          <div className='flex items-center justify-between'>
            <h2 className='text-4xl font-medium text-strong'>
              {flow?.data.displayReceiveAmount || '$ 0.00'}
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
        <CardContent>
          <TextField
            id='amount'
            form='amount-form'
            label='Amount'
            name='amount'
            defaultValue={flow?.data.amount}
            onChange={_onChangeInput}
            prefix='$'
            type='number'
            min='0'
            step='0.01'
            aria-invalid={Boolean(fetcher.data?.errors.amount) || undefined}
            aria-describedby={
              fetcher.data?.errors.amount ? 'amount-error' : undefined
            }
            errorMessage={fetcher.data?.errors.amount || undefined}
            required
          />
        </CardContent>
      </Card>
      <Card>
        <CardContent>
          <span>Select an account to pay from:</span>
          <Select
            id='linkedAccount'
            label='Connected accounts'
            className='mt-4'
            value={linkedAccount}
            options={linkedAccounts}
            onChange={_onChangeLinkedAccount}
            aria-invalid={
              Boolean(fetcher.data?.errors.linkedAccount) || undefined
            }
            aria-describedby={
              fetcher.data?.errors.linkedAccount
                ? 'linkedAccount-error'
                : undefined
            }
            errorMessage={fetcher.data?.errors.linkedAccount}
          />
          <input
            form='amount-form'
            value={linkedAccount.id}
            name='toLinkedAccountId'
            type='hidden'
          />
          <TextField
            id='note'
            label='Reference'
            name='note'
            form='amount-form'
            type='text'
            defaultValue={flow.data.note || ''}
            className='mt-4'
            aria-invalid={Boolean(fetcher.data?.errors.note) || undefined}
            aria-describedby={
              fetcher.data?.errors.note ? 'reference-error' : undefined
            }
            errorMessage={fetcher.data?.errors.note}
          />
        </CardContent>
      </Card>
      <Button form='amount-form' type='submit' name='route-to' value='next'>
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
