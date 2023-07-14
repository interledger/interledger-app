import type { ActionArgs, LoaderArgs, MetaFunction } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { Form, useActionData, useLoaderData } from '@remix-run/react'
import { useState } from 'react'
import { route } from 'routes-gen'
import type { ApplicationProps } from '~/components'
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
  Layouts,
  LinkedInIcon,
  TextButton,
  TwitterIcon
} from '~/components'
import { Label } from '~/components/Label'
import { exitFlow, flowType, requireFlow } from '~/lib/flows.server'
import { getClientIP } from '~/lib/ip.server'
import { getUserSession } from '~/lib/kratos.server'
import {
  StatusError,
  httpMapping,
  isGrpcError,
  openPaymentsClient
} from '~/lib/proto.server'
import { getLinkedAccounts, getPublicWalletInfo } from '~/lib/wallet.server'

export async function loader({ request }: LoaderArgs) {
  const session = await getUserSession(request)
  const flow = await requireFlow(request, flowType.Pay)
  const { cardAccounts, bankAccounts } = await getLinkedAccounts(request)

  const publicWalletInfo = await getPublicWalletInfo(
    request,
    flow.data.address.walletUrl
  )
  return json({
    flow,
    publicWalletInfo,
    traits: session.identity.traits,
    // TODO use a lookup for this account rather?
    linkedAccount: [...cardAccounts, ...bankAccounts].find(
      (account) => account.id == flow.data.toLinkedAccountId
    )
  })
}

export const handle: ApplicationProps = {
  layout: Layouts.Focus,
  scaffold: {
    header: {
      back: route('/pay/amount'),
      title: 'Confirm Payment'
    }
  }
}

export const meta: MetaFunction = () => {
  return {
    title: 'Pay | Confirm'
  }
}

export default function Page() {
  const { flow, publicWalletInfo, linkedAccount } =
    useLoaderData<typeof loader>()
  const actionData = useActionData<typeof action>()
  const [showDialog, setShowDialog] = useState<boolean>(false)

  return (
    <>
      <Form
        id='pay-confirm'
        action={route('/pay/confirm')}
        method='post'
        className='hidden'
      />
      <Card>
        <CardContent>
          <div className='flex items-center justify-between'>
            <h2 className='text-4xl font-medium text-strong'>
              {flow?.data.displayReceiveAmount || '$ 0.00'}
            </h2>
            {flow.data.address.identifierType === 'wallet' && (
              <FynbosIcon height='h-12' />
            )}
            {flow.data.address.identifierType === 'twitter' && (
              <TwitterIcon height='h-12' />
            )}
          </div>
        </CardContent>
        <Label className='mt-2'>Payment to</Label>
        <CardButton onClick={() => setShowDialog(true)}>
          <div className='flex w-full items-center justify-between text-medium'>
            <span>{flow.data.address.identifier}</span>
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
            <span className='text-medium'>
              {flow?.data.displayReceiveAmount || '$ 0.00'}
            </span>
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
            <span className='text-medium'>{linkedAccount?.name}</span>
          </div>
          {flow?.data.note && (
            <div className='mt-4 flex w-full flex-col space-y-1'>
              <span className='text-weak'>Reference</span>
              <span className='text-medium'>{flow?.data.note}</span>
            </div>
          )}
        </CardContent>
      </Card>
      <Card>
        <CardContent>
          <Checkbox
            id='service-agreement'
            name='service-agreement'
            form='pay-confirm'
            className='flex'
            aria-invalid={
              Boolean(actionData?.errors.serviceAgreement) || undefined
            }
            aria-describedby={
              actionData?.errors.serviceAgreement
                ? 'serviceAgreement-error'
                : undefined
            }
            errorMessage={actionData?.errors.serviceAgreement}
          >
            I authorize Fynbos to debit
            {linkedAccount?.type == 'card'
              ? ' the card indicated '
              : ' my account '}
            for the amount noted on today’s date. I will not dispute Fynbos
            debiting my account, so long as the transaction corresponds to the
            terms in this online form and my agreement with Fynbos.
          </Checkbox>
          <input
            form='pay-confirm'
            defaultValue={linkedAccount?.type}
            name='linked-account-type'
            type='hidden'
          />
        </CardContent>
      </Card>
      <Button form='pay-confirm' type='submit'>
        Confirm payment
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
          <span className=''>{publicWalletInfo.publicName}</span>
        </div>
        <Label className='mt-2'>Wallet address</Label>
        <CardLink className='flex w-full' to={publicWalletInfo.address}>
          <div className='flex w-full items-center justify-between text-medium'>
            <div className='flex space-x-2'>
              <FynbosIcon />
              <span>{publicWalletInfo.shortAddress}</span>
            </div>
            <Icon>navigate_next</Icon>
          </div>
        </CardLink>
        {publicWalletInfo.identities.map((identity) => (
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

export async function action({ request }: ActionArgs) {
  const flow = await requireFlow(request, flowType.Pay)
  const form = await request.formData()
  const serviceAgreement = form.get('service-agreement') as string
  const linkedAccountType = form.get('linked-account-type') as string

  const fieldErrors = {
    form: '',
    serviceAgreement: ''
  }

  if (serviceAgreement == null) {
    fieldErrors.serviceAgreement = 'You are required to authorize to continue.'
    return json(
      {
        errors: {
          ...fieldErrors
        }
      },
      { status: 400 }
    )
  }

  if (linkedAccountType === 'card') {
    return redirect(route('/pay/3ds'))
  }

  const clientIpAddress = getClientIP(request)

  const response = await openPaymentsClient
    .createOutgoingPayment(
      {
        idempotencyKey: flow.data.idempotencyKey || '',
        quoteID: flow.data.quoteID,
        description: flow.data.note,
        externalRef: '',
        ipAddress: clientIpAddress,
        threeDSID: '',
        identityType: flow.data.address.identifierType,
        identity: flow.data.address.identifier
      },
      {
        meta: {
          cookies: String(request.headers.get('cookie')) || ''
        }
      }
    )
    .then((v) => v)
    .catch(StatusError)
  if (isGrpcError(response)) throw json({}, httpMapping(response.code))

  await exitFlow(request, flowType.Pay)
  return redirect(route('/'))
}
