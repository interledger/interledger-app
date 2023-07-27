import { useFetcher } from '@remix-run/react'
import { useEffect, useState } from 'react'
import { route } from 'routes-gen'
import {
  Button,
  Card,
  CardButton,
  CardContent,
  CardHeader,
  CardLink,
  Checkbox,
  Chip,
  ChipColor,
  Dialog,
  FynbosIcon,
  Icon,
  LinkedInIcon,
  TextButton,
  TextField,
  TwitterIcon
} from '~/components'
import { Label } from '~/components/Label'
import { usePayStore } from '~/lib/usePayStore'

export function Confirm() {
  const confirm = useFetcher()
  const otpFetcher = useFetcher()

  const phoneMask = useFetcher()

  const [showOTPDialog, setShowOTPDialog] = useState<boolean>(false)
  const [showPublicWalletDialog, setShowPublicWalletDialog] =
    useState<boolean>(false)

  const [
    address,
    account,
    displayAmount,
    note,
    publicWalletInfo,
    quoteId,
    requiresOTP
  ] = usePayStore((state) => [
    state.address,
    state.account,
    state.displayAmount,
    state.note,
    state.publicWalletInfo,
    state.quoteId,
    state.requiresOTP
  ])

  useEffect(() => {
    if (phoneMask.state == 'idle' && phoneMask.data == null && requiresOTP) {
      phoneMask.load(`/pay?phone=true`)
    }
  }, [phoneMask, requiresOTP])

  useEffect(() => {
    if (
      !showOTPDialog &&
      otpFetcher.state == 'loading' &&
      otpFetcher?.data?.success
    ) {
      setShowOTPDialog(true)
    }
  }, [otpFetcher?.data, otpFetcher.state, showOTPDialog])

  return (
    <>
      <otpFetcher.Form
        id='pay-phone-otp'
        action='/api/sendOtp'
        method='post'
        className='hidden'
      />
      <confirm.Form
        id='pay-confirm'
        action={route('/pay')}
        method='post'
        className='hidden'
      />
      <input
        form='pay-confirm'
        defaultValue={quoteId}
        name='quoteId'
        type='hidden'
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
        <CardButton noHover onClick={() => setShowPublicWalletDialog(true)}>
          <div className='flex w-full items-center justify-between text-medium'>
            <span>{address?.identifier}</span>
            <Icon>navigate_next</Icon>
          </div>
        </CardButton>
      </Card>
      <Card>
        <CardContent>
          <div className='flex w-full justify-between'>
            <span className='text-weak'>Total fees</span>
            <span className='text-medium'>
              Free <sup>*</sup>
            </span>
          </div>
          <div className='mt-2 flex w-full justify-between'>
            <span className='text-weak'>They receive</span>
            <span className='text-medium'>{displayAmount}</span>
          </div>
          <div className='mt-4 flex w-full space-x-2'>
            <span className='text-xs text-medium'>*</span>
            <span className='text-xs text-medium'>
              For a limited time, Fynbos will absorb the fees associated with
              making a payment.
            </span>
          </div>
        </CardContent>
      </Card>
      <Card>
        <CardContent>
          <div className='flex w-full flex-col justify-between space-y-1'>
            <span className='text-weak'>Source</span>
            <span className='text-medium'>{account?.name}</span>
          </div>
          {note && (
            <div className='mt-4 flex w-full flex-col space-y-1'>
              <span className='text-weak'>Reference</span>
              <span className='text-medium'>{note}</span>
            </div>
          )}
        </CardContent>
      </Card>
      <Card>
        <CardContent>
          <Checkbox
            id='service-agreement'
            name='serviceAgreement'
            form='pay-confirm'
            className='flex'
            aria-invalid={
              Boolean(confirm.data?.errors?.serviceAgreement) || undefined
            }
            aria-describedby={
              confirm.data?.errors?.serviceAgreement
                ? 'serviceAgreement-error'
                : undefined
            }
            errorMessage={confirm.data?.errors?.serviceAgreement}
          >
            I authorize Fynbos to debit
            {account?.type == 'card' ? ' the card indicated ' : ' my account '}
            for the amount noted on today’s date. I will not dispute Fynbos
            debiting my account, so long as the transaction corresponds to the
            terms in this online form and my agreement with Fynbos.
          </Checkbox>
        </CardContent>
      </Card>
      {requiresOTP && (
        <Button form='pay-phone-otp' type='submit'>
          Confirm payment
        </Button>
      )}
      {!requiresOTP && (
        <Button
          form='pay-confirm'
          name='formName'
          value='confirm'
          type='submit'
        >
          Confirm payment
        </Button>
      )}
      <Dialog open={showPublicWalletDialog} setOpen={setShowPublicWalletDialog}>
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
        <CardLink
          className='flex w-full'
          to={publicWalletInfo?.address as string}
        >
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
          <TextButton
            type='button'
            onClick={() => setShowPublicWalletDialog(false)}
          >
            Close
          </TextButton>
        </CardContent>
      </Dialog>
      <Dialog open={showOTPDialog} setOpen={setShowOTPDialog}>
        <CardHeader>
          <h1 className='text-xl font-medium'>Two-step verification</h1>
        </CardHeader>
        <CardContent>
          <span className='text-medium'>
            Enter the six digit code sent to your mobile number.
          </span>
        </CardContent>
        <Label className='mt-2'>Your mobile phone number</Label>
        <div className='mt-1 flex space-x-2 rounded-xl bg-nav p-3 text-medium'>
          <Icon>phone_android</Icon>
          <span>{phoneMask?.data?.phoneMask}</span>
        </div>

        <TextField
          id='otp'
          form='pay-confirm'
          label='Verification code'
          name='otp'
          type='number'
          className='mt-4'
          aria-invalid={Boolean(confirm.data?.errors.otp) || undefined}
          aria-describedby={
            confirm.data?.errors.otp ? 'email-error' : undefined
          }
          required
          errorMessage={confirm.data?.errors.otp}
        />
        <CardContent className='mt-2 flex w-full justify-end space-x-6'>
          <TextButton type='submit' form='pay-phone-otp'>
            Resend code
          </TextButton>
          <TextButton
            form='pay-confirm'
            name='formName'
            value='confirm'
            type='submit'
          >
            Verify
          </TextButton>
        </CardContent>
      </Dialog>
    </>
  )
}
