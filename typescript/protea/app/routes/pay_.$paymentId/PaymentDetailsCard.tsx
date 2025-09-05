import { useLoaderData } from '@remix-run/react'
import { useState } from 'react'
import {
  Alert,
  AlertBody,
  Card,
  CardButton,
  CardContent,
  CardHeader,
  CardLink,
  Chip,
  ChipColor,
  Dialog,
  DiscordIcon,
  Icon,
  InterledgerIcon,
  LinkedInIcon,
  SlackIcon,
  TextButton,
  TwitterIcon
} from '~/components'
import { Label } from '~/components/Label'
import type { loader } from './route'
import { PaymentIdentityType } from './route'

export const PaymentDetailsCard = () => {
  const { publicWalletInfo, payment } = useLoaderData<typeof loader>()

  const [showDialog, setShowDialog] = useState<boolean>(false)

  return (
    <>
      <Card>
        <Label className='mt-2'>Payment to</Label>
        <CardButton noHover onClick={() => setShowDialog(true)}>
          <div className='flex w-full items-center justify-between text-medium'>
            <div className='flex space-x-2'>
              {(payment.receiverIdentityType ===
                PaymentIdentityType.WalletURL ||
                payment.receiverIdentityType ===
                  PaymentIdentityType.WalletID) && <InterledgerIcon />}
              {payment.receiverIdentityType === PaymentIdentityType.Twitter && (
                <TwitterIcon />
              )}
              {payment.receiverIdentityType === PaymentIdentityType.Discord && (
                <DiscordIcon />
              )}
              {payment.receiverIdentityType === PaymentIdentityType.Slack && (
                <SlackIcon />
              )}
              <span>{publicWalletInfo.publicName}</span>
            </div>
            <Icon>navigate_next</Icon>
          </div>
        </CardButton>
      </Card>
      {publicWalletInfo.walletID == 'not-found' && (
        <>
          <Alert>
            <Icon>error</Icon>
            <AlertBody>
              This person is not an Interledger Wallet user yet. Confirming the
              payment will prompt them to sign up to the Interledger Wallet to
              receive it.
            </AlertBody>
          </Alert>
        </>
      )}

      <Dialog open={showDialog} setOpen={setShowDialog}>
        <CardHeader>
          <h1 className='text-xl font-medium'>User information</h1>
        </CardHeader>
        {publicWalletInfo.walletID == 'not-found' && (
          <CardContent>
            <span className='text-medium'>
              This person is not an Interledger Wallet user yet.
            </span>
          </CardContent>
        )}
        {publicWalletInfo.walletID !== 'not-found' && (
          <>
            <CardContent>
              <span className='text-medium'>
                You are viewing public information about the person you intend
                to pay.
              </span>
            </CardContent>
            <Label className='mt-4'>Public name</Label>
            <div className='mt-1 flex rounded-xl bg-nav p-3 text-medium'>
              <span className=''>{publicWalletInfo?.publicName}</span>
            </div>
            <Label className='mt-2'>Wallet address</Label>
            <CardLink
              className='flex w-full'
              to={publicWalletInfo?.address ?? ''}
            >
              <div className='flex w-full items-center justify-between text-medium'>
                <div className='flex space-x-2'>
                  <InterledgerIcon />
                  <span>{publicWalletInfo?.shortAddress}</span>
                </div>
                <Icon>navigate_next</Icon>
              </div>
            </CardLink>
          </>
        )}
        {publicWalletInfo?.identities.map((identity) => (
          <div key={identity.id} className='contents'>
            <Label className='mt-2 capitalize'>{identity.platform}</Label>
            <CardLink className='flex w-full' to={publicWalletInfo.address}>
              <div className='flex w-full items-center justify-between text-medium'>
                <div className='flex space-x-2'>
                  {identity.platform == 'twitter' && <TwitterIcon />}
                  {identity.platform == 'linkedin' && <LinkedInIcon />}
                  {identity.platform == 'discord' && <DiscordIcon />}
                  {identity.platform == 'slack' && <SlackIcon />}
                  {identity.platform == 'domain' && <Icon>captive_portal</Icon>}
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
