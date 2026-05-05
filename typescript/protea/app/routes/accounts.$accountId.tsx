import type { Route } from './+types/accounts.$accountId'
import { data } from 'react-router';
import type { UIMatch } from 'react-router';
import { Form, useFetcher, useLoaderData, useParams } from 'react-router';
import { useCallback, useState } from 'react'
import { href } from 'react-router'
import type { ApplicationProps } from '~/components'
import {
  Alert,
  AlertBody,
  AlertContent,
  AlertTitle,
  Button,
  Card,
  CardContent,
  CardHeader,
  CardLink,
  CardTitle,
  Checkbox,
  Chip,
  ChipColor,
  Dialog,
  Icon,
  Layouts,
  Router,
  TextButton
} from '~/components'
import { Label } from '~/components/Label'
import type { FormattedLinkedAccount } from '~/data/accounts.server'
import { getLinkedAccount } from '~/data/accounts.server'
import { jsonWithCSRF, validateCSRFToken } from '~/lib/csrf.server'
import { isConnectError } from '~/lib/error.server'
import { grpc } from '~/lib/grpc.server'
import { jsonWithSnackbar, redirectWithSnackbar } from '~/lib/snackbar.server'

export async function loader(args: Route.LoaderArgs) {
  const account = await getLinkedAccount(
    args.request,
    args.params.accountId as string
  )
  if (isConnectError(account)) throw account.errorResponse

  if (account.type == 'card' || account.type == 'sendCard') {
    return cardLoader(args, account)
  }

  if (account.type == 'bank') {
    return bankLoader(args, account)
  }

  if (account.type == 'interac') {
    return bankLoader(args, account)
  }

  throw data({}, { status: 404 })
}

async function bankLoader(
  { request, params }: Route.LoaderArgs,
  account: FormattedLinkedAccount
) {
  return jsonWithCSRF(request, {
    account
  })
}

async function cardLoader(
  { request, params }: Route.LoaderArgs,
  account: FormattedLinkedAccount
) {
  const card = await grpc.getCardDetails(request, {
    id: params.accountId as string
  })
  if (isConnectError(card)) throw card.errorResponse

  return jsonWithCSRF(request, {
    card,
    account
  })
}

export const handle: ApplicationProps = {
  layout: Layouts.Wallet,
  scaffold: {
    header: {
      back: href('/accounts'),
      actions: (match: UIMatch<Route.ComponentProps['loaderData']>) => {
        if (!match.loaderData) return null
        const { account } = match.loaderData
        const state = account.state
        if (state == 'Verified') {
          const canSend = account.canSend
          const canReceive = account.canReceive
          if (canSend && canReceive) {
            return {
              key: 'send+receive',
              nodes: [
                <Chip key='send' color={ChipColor.indigo}>
                  Send
                </Chip>,
                <Chip key='receive' color={ChipColor.purple}>
                  Receive
                </Chip>
              ]
            }
          } else if (!canSend && canReceive) {
            return {
              key: 'Receive only',
              nodes: <Chip color={ChipColor.purple}>Receive only</Chip>
            }
          } else if (canSend && !canReceive) {
            return {
              key: 'Send only',
              nodes: <Chip color={ChipColor.indigo}>Send only</Chip>
            }
          } else return null
        } else if (state == 'OwnershipReviewRequired') {
          return {
            key: 'In review',
            nodes: <Chip color={ChipColor.orange}>In review</Chip>
          }
        } else if (state == 'Rejected') {
          return {
            key: 'Rejected',
            nodes: <Chip color={ChipColor.red}>Rejected</Chip>
          }
        } else return null
      }
    },
    isNested: true
  }
}

export default function Page() {
  const { account } = useLoaderData()

  return (
    <>
      {account.type == 'card' && <CardDetailsPage />}
      {account.type == 'bank' && <BankDetailsPage />}
      {account.type == 'interac' && <InteracDetailsPage />}
    </>
  )
}

function InteracDetailsPage() {
  const { account } = useLoaderData()
  const params = useParams()
  return (
    <>
      <Card>
        <CardHeader>
          <CardTitle>Interac account details</CardTitle>
        </CardHeader>
        <CardContent>
          <div className='flex w-full flex-col justify-between space-y-1'>
            <span className='text-weak'>Email</span>
            <span className='text-medium'>{account.mask}</span>
          </div>
        </CardContent>
      </Card>
      <Card>
        <Label>Nickname</Label>
        <CardLink
          className='flex items-center justify-between'
          to={href('/accounts/:accountId/name', {
            accountId: params.accountId as string
          })}
        >
          {account?.nickname}
          <Icon>navigate_next</Icon>
        </CardLink>
      </Card>
    </>
  )
}

function BankDetailsPage() {
  const { account } = useLoaderData<typeof bankLoader>()
  const params = useParams()
  return (
    <>
      <Card>
        <CardHeader>
          <CardTitle>Bank account details</CardTitle>
        </CardHeader>
        <CardContent>
          <div className='flex w-full flex-col justify-between space-y-1'>
            <span className='text-weak'>Account number</span>
            <span className='text-medium'>{account.mask}</span>
          </div>
        </CardContent>
      </Card>
      <Card>
        <Label>Bank nickname</Label>
        <CardLink
          className='flex items-center justify-between'
          to={href('/accounts/:accountId/name', {
            accountId: params.accountId as string
          })}
        >
          {account?.nickname}
          <Icon>navigate_next</Icon>
        </CardLink>
      </Card>
    </>
  )
}

function CardDetailsPage() {
  const { card, account, csrfToken } = useLoaderData<typeof cardLoader>()
  const params = useParams()
  const [limitationsDialog, setLimitationsDialog] = useState<boolean>(false)
  const [showDialog, setShowDialog] = useState<boolean>(false)

  const updateAccountFetcher = useFetcher<typeof action>()

  const _onChangeLinkedAccount = useCallback(
    (updateType: 'defaultSend' | 'defaultReceive') => {
      updateAccountFetcher.submit(
        {
          formName: updateType,
          accountType: 'card',
          csrfToken
        },
        { method: 'post' }
      )
    },
    [csrfToken, updateAccountFetcher]
  )

  return (
    <>
      <Form
        id='account'
        action={href('/accounts/:accountId', {
          accountId: params.accountId as string
        })}
        method='post'
        className='hidden'
      />
      {card.state == 'OwnershipReviewRequired' && (
        <Alert>
          <Icon>error</Icon>
          <AlertBody>
            This card is currently in review, and you will be unable to send or
            receive payments with it. We will notify you once verified.
          </AlertBody>
        </Alert>
      )}
      {card.state == 'Rejected' && (
        <Alert>
          <Icon>error</Icon>
          <AlertContent>
            <AlertBody>
              This card cannot process payments because of an address mismatch.
            </AlertBody>
            <Router className='text-primary' to={href('/support')}>
              Contact support
            </Router>
          </AlertContent>
        </Alert>
      )}
      {card.state == 'Verified' && card.canSend && !card.canReceive && (
        <Alert>
          <Icon>error</Icon>
          <AlertContent className='items-start'>
            <AlertTitle>Send only</AlertTitle>
            <AlertBody>
              This card is enabled to send payments, but unable to receive
              payments.
            </AlertBody>
            <TextButton
              className='text-primary'
              onClick={() => setLimitationsDialog(true)}
            >
              More information
            </TextButton>
          </AlertContent>
        </Alert>
      )}
      {card.state == 'Verified' && !card.canSend && card.canReceive && (
        <Alert>
          <Icon>error</Icon>
          <AlertContent className='items-start'>
            <AlertTitle>Receive only</AlertTitle>
            <AlertBody>
              This card is enabled to receive payments, but unable to make
              payments.
            </AlertBody>
            <TextButton
              className='text-primary'
              onClick={() => setLimitationsDialog(true)}
            >
              More information
            </TextButton>
          </AlertContent>
        </Alert>
      )}
      <Card>
        <CardHeader>
          <CardTitle>Card details</CardTitle>
        </CardHeader>
        <CardContent>
          <div className='flex w-full flex-col justify-between space-y-1'>
            <span className='text-weak'>Card number</span>
            <span className='text-medium'>
              {card?.bin.replace(' ', '').slice(0, 4)}{' '}
              {card?.bin.replace(' ', '').slice(4).padEnd(4, '*')} ****{' '}
              {card?.last4}
            </span>
          </div>
          <div className='mt-4 flex w-full flex-col space-y-1'>
            <span className='text-weak'>Card type</span>
            <span className='text-medium'>{card.type}</span>
          </div>
          <div className='mt-4 flex w-full flex-col space-y-1'>
            <span className='text-weak'>Expiration date</span>
            <span className='text-medium'>{card.expiration}</span>
          </div>
        </CardContent>
      </Card>
      <Card>
        <Label>Card nickname</Label>
        <CardLink
          className='flex items-center justify-between'
          to={href('/accounts/:accountId/name', {
            accountId: params.accountId as string
          })}
        >
          {card?.nickname}
          <Icon>navigate_next</Icon>
        </CardLink>
      </Card>
      <Card>
        <CardContent>
          <p>Set as the default for sending or receiving payments.</p>
          {account.canSend && (
            <Checkbox
              id='service-agreement'
              name='serviceAgreement'
              form='pay-confirm'
              className='mt-6 flex items-center'
              disabled={account.defaultSend}
              checked={account.defaultSend}
              onChange={() => _onChangeLinkedAccount('defaultSend')}
            >
              Default send
            </Checkbox>
          )}
          {account.canReceive && (
            <Checkbox
              id='service-agreement'
              name='serviceAgreement'
              form='pay-confirm'
              className='mt-6 flex items-center'
              disabled={account.defaultReceive}
              checked={account.defaultReceive}
              onChange={() => _onChangeLinkedAccount('defaultReceive')}
            >
              Default receive
            </Checkbox>
          )}
        </CardContent>
      </Card>
      <Button error type='button' onClick={() => setShowDialog(true)}>
        Remove card
      </Button>
      <Dialog open={showDialog} setOpen={setShowDialog}>
        <CardHeader>
          <h1 className='text-xl font-medium'>Remove card</h1>
        </CardHeader>
        <CardContent>
          <span className='text-medium'>
            Are you sure you want to remove the card? This action cannot be
            undone.
          </span>

          <div className='flex w-full justify-end space-x-6 pt-4'>
            <TextButton
              type='button'
              className='!text-medium'
              onClick={() => setShowDialog(false)}
            >
              Cancel
            </TextButton>
            <TextButton
              name='formName'
              className='!text-error'
              value='delete'
              form='account'
              type='submit'
            >
              Remove card
            </TextButton>
          </div>
        </CardContent>
      </Dialog>
      <Dialog open={limitationsDialog} setOpen={setLimitationsDialog}>
        <CardHeader>
          <h1 className='text-xl font-medium'>Card limitations</h1>
        </CardHeader>
        <CardContent>
          <p className='text-medium'>
            Limitations are placed on some cards enabling sending or receiving
            money only. Please contact your card issuer to remove limitations,
            or connect another card.
          </p>

          <div className='flex w-full justify-end space-x-6 pt-4'>
            <TextButton
              type='button'
              onClick={() => setLimitationsDialog(false)}
            >
              Close
            </TextButton>
          </div>
        </CardContent>
      </Dialog>
    </>
  )
}

export async function action({ request, params }: Route.ActionArgs) {
  const form = await request.formData()
  const formName = form.get('formName') as string
  const accountId = params.accountId as string

  await validateCSRFToken(request, form)

  let accountType = 'Card'
  if (form.get('accountType') == 'wallet') {
    accountType = 'Wallet'
  }

  let response
  switch (formName) {
    case 'defaultSend':
      response = await grpc.setDefaultSendLinkedAccount(request, {
        id: accountId
      })
      if (isConnectError(response)) throw response.errorResponse
      return jsonWithSnackbar(
        request,
        {},
        {
          message: accountType + ' set as default send.',
          icon: 'close'
        }
      )
    case 'defaultReceive':
      response = await grpc.setDefaultReceiveLinkedAccount(request, {
        id: accountId
      })
      if (isConnectError(response)) throw response.errorResponse
      return jsonWithSnackbar(
        request,
        {},
        {
          message: accountType + ' set as default receive.',
          icon: 'close'
        }
      )
    case 'delete':
      response = await grpc.deleteLinkedAccount(request, { id: accountId })
      if (isConnectError(response)) throw response.errorResponse
      return redirectWithSnackbar(request, href('/accounts'), {
        message: 'Card deleted successfully.',
        icon: 'close'
      })
    default:
      throw data(
        { title: "Submitted a form that doesn't exist" },
        {
          status: 400
        }
      )
  }
}
