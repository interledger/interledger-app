import type { PlainMessage } from '@bufbuild/protobuf'
import type {
  ActionFunctionArgs,
  LoaderFunctionArgs,
  MetaFunction
} from '@remix-run/node'
import { Form, Link, useActionData, useLoaderData } from '@remix-run/react'
import { useEffect, useState } from 'react'
import { route } from 'routes-gen'
import type { ApplicationProps } from '~/components'
import {
  Alert,
  AlertBody,
  AlertTitle,
  AnchorRouter,
  Autocomplete,
  Button,
  Card,
  CardButton,
  CardContent,
  CardHeader,
  CardLink,
  CardTitle,
  Chip,
  ChipColor,
  Dialog,
  DiscordIcon,
  FynbosIcon,
  Icon,
  Layouts,
  LinkedInIcon,
  SlackIcon,
  TextButton,
  TextField,
  TwitterIcon
} from '~/components'
import { Label } from '~/components/Label'
import type {
  Country,
  PublicWalletInfo
} from '~/generated/connect/backend/v1/backend_pb'
import { jsonWithCSRF } from '~/lib/csrf.server'
import { isConnectError } from '~/lib/error.server'
import { grpc } from '~/lib/grpc.server'
import { getClientIP } from '~/lib/ip.server'
import { mergeMeta } from '~/lib/meta'
import { redirectWithSnackbar } from '~/lib/snackbar.server'

export const handle: ApplicationProps = {
  layout: Layouts.Focus,
  scaffold: {
    header: {
      back: route('/'),
      title: 'Collect Payment'
    }
  }
}

export const meta: MetaFunction = mergeMeta(() => [
  {
    title: 'Collect Payment'
  }
])

export async function loader({ request, params }: LoaderFunctionArgs) {
  let link = await grpc.getPaymentLink(request, {
    id: params.linkId
  })
  if (isConnectError(link)) throw link.errorResponse

  let countriesRpc = await grpc.getCountries(request, {})

  if (isConnectError(countriesRpc)) throw countriesRpc.errorResponse

  let publicWalletInfo: PlainMessage<PublicWalletInfo>
  const publicWalletInfoResponse = await grpc.getPublicWalletInfo(request, {
    walletAddress: link.senderWalletUrl
  })

  if (isConnectError(publicWalletInfoResponse)) {
    publicWalletInfo = {
      walletID: 'not-found',
      address: link.senderWalletUrl,
      shortAddress: '',
      publicName: '',
      identities: [],
      canReceive: false
    }
  } else publicWalletInfo = publicWalletInfoResponse

  return jsonWithCSRF(request, {
    ...link,
    publicWalletInfo,
    countries: countriesRpc.countries
  })
}

export default function Page() {
  const { completed, expired } = useLoaderData<typeof loader>()

  if (completed || expired) {
    return (
      <>
        <Alert>
          <Icon>notification_important</Icon>
          <AlertBody>
            <AlertTitle>Payment {expired ? 'Expired' : 'Completed'}</AlertTitle>
            <AlertBody>
              Please contact our support team for any queries.
            </AlertBody>
          </AlertBody>
        </Alert>
        <Card>
          <CardHeader>
            <CardTitle>Support details</CardTitle>
          </CardHeader>
          <CardContent>
            <p>
              Our telephone support lines are open Monday to Friday between 9am
              and 5pm EST.
            </p>
            <div className='mt-4 flex items-center space-x-2 text-medium'>
              <Icon>call</Icon>
              <AnchorRouter
                to='tel:+1 (856) 249-3067'
                className='text-sm text-primary'
              >
                +1 (856) 249-3067
              </AnchorRouter>
            </div>
            <div className='mt-4 flex items-center space-x-2 text-medium'>
              <Icon>mail</Icon>
              <AnchorRouter
                to='mailto:support@fynbos.app'
                className='text-sm text-primary'
              >
                support@fynbos.app
              </AnchorRouter>
            </div>
          </CardContent>
        </Card>
      </>
    )
  } else {
    return <PaymentLink />
  }
}

function PaymentLink() {
  const {
    id,
    note,
    receiverIdentifier,
    publicWalletInfo,
    formattedAmount,
    countries
  } = useLoaderData<typeof loader>()
  const [showDialog, setShowDialog] = useState<boolean>(false)
  const actionData = useActionData<typeof action>()

  const [query, setQuery] = useState<string>('')
  const [filteredCountries, setFilteredCountries] = useState(countries)
  const [country, setCountry] = useState<PlainMessage<Country>>()

  useEffect(() => {
    if (query === '') setFilteredCountries(countries)
    else {
      setFilteredCountries(
        countries.filter((country) => {
          return (
            country.name
              .toLowerCase()
              .replace(/\s+/g, '')
              .includes(query.toLowerCase().replace(/\s+/g, '')) ||
            country.id
              .toLowerCase()
              .replace(/\s+/g, '')
              .includes(query.toLowerCase().replace(/\s+/g, ''))
          )
        })
      )
    }
  }, [query, countries])

  return (
    <>
      <Form
        id='details'
        action={route('/collect/:linkId', { linkId: id })}
        method='post'
        className='hidden'
      />
      <Card>
        <CardContent className='space-y-4 text-medium'>
          <p>Hello {receiverIdentifier},</p>
          <h2 className='text-xl font-medium'>You've been sent a payment</h2>
          <h2 className='text-4xl font-medium'>{formattedAmount}</h2>
        </CardContent>
        <CardContent className='mt-2'>
          <span className='text-weak'>Payment from</span>
        </CardContent>
        <CardButton
          noHover
          onClick={() => {
            setShowDialog(true)
          }}
        >
          <div className='flex w-full items-center justify-between text-medium'>
            <div className='flex space-x-2'>
              <FynbosIcon />
              <span>{publicWalletInfo.publicName}</span>
            </div>
            <Icon>navigate_next</Icon>
          </div>
        </CardButton>

        {note && (
          <CardContent className='mt-2'>
            <div className='flex w-full flex-col space-y-1'>
              <span className='text-weak'>Payment note</span>
              <span className='text-medium'>{note}</span>
            </div>
          </CardContent>
        )}
      </Card>
      <Card>
        <CardContent>
          To collect the payment,{' '}
          <Link className='text-primary' to={route('/login')}>
            log in
          </Link>{' '}
          or provide your details below.
        </CardContent>
        <TextField
          id='firstName'
          name='firstName'
          label='First Name'
          form='details'
          className='mt-4'
          aria-invalid={Boolean(actionData?.errors?.firstName) || undefined}
          aria-describedby={
            actionData?.errors?.firstName ? 'firstName-error' : undefined
          }
          required
          errorMessage={actionData?.errors?.firstName}
        />
        <TextField
          id='lastName'
          name='lastName'
          label='Last Name'
          form='details'
          className='mt-4'
          aria-invalid={Boolean(actionData?.errors?.lastName) || undefined}
          aria-describedby={
            actionData?.errors?.lastName ? 'lastName-error' : undefined
          }
          required
          errorMessage={actionData?.errors?.lastName}
        />
        <TextField
          id='email'
          name='email'
          type='email'
          label='Email address'
          form='details'
          className='mt-4'
          aria-invalid={Boolean(actionData?.errors?.email) || undefined}
          aria-describedby={
            actionData?.errors?.email ? 'email-error' : undefined
          }
          required
          errorMessage={actionData?.errors?.email}
        />
        <Autocomplete
          id='country'
          value={country as Country}
          onChange={setCountry}
          onQuery={setQuery}
          options={filteredCountries}
          label='Country of residence'
          className='mt-4'
          aria-invalid={Boolean(actionData?.errors?.country) || undefined}
          aria-describedby={
            actionData?.errors?.country ? 'country-error' : undefined
          }
          errorMessage={actionData?.errors?.country}
        />
        <input
          form='details'
          value={String(country?.id)}
          name='country'
          type='hidden'
        />
      </Card>
      <Button form='details' type='submit'>
        Continue
      </Button>

      <Dialog
        open={showDialog}
        setOpen={() => {
          setShowDialog(true)
        }}
      >
        <CardHeader>
          <h1 className='text-xl font-medium'>User information</h1>
        </CardHeader>
        {publicWalletInfo.walletID == 'not-found' && (
          <CardContent>
            <span className='text-medium'>
              This person is not a Fynbos user yet.
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
          </>
        )}

        {publicWalletInfo.identities.map((identity) => (
          <div key={identity.id} className='contents'>
            <Label className='mt-2 capitalize'>{identity.platform}</Label>
            <CardLink className='flex w-full' to={publicWalletInfo.address}>
              <div className='flex w-full items-center justify-between text-medium'>
                <div className='flex space-x-2'>
                  {identity.platform == 'twitter' && <TwitterIcon />}
                  {identity.platform == 'linkedin' && <LinkedInIcon />}
                  {identity.platform == 'discord' && <DiscordIcon />}
                  {identity.platform == 'slack' && <SlackIcon />}
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

export async function action({ request, params }: ActionFunctionArgs) {
  const form = await request.formData()
  const firstName = String(form.get('firstName') || '')
  const lastName = String(form.get('lastName') || '')
  const email = String(form.get('email') || '')
  const country = String(form.get('country') || '')

  let errors = {
    firstName: '',
    lastName: '',
    email: '',
    country: ''
  }

  let ipAddress = await getClientIP(request)
  let tokenResponse = await grpc.consumePaymentLink(request, {
    id: params.linkId,
    email,
    firstName,
    lastName,
    ipAddress,
    countryCode: country
  })
  if (isConnectError(tokenResponse)) {
    return tokenResponse.error({ errors }, {}, { action: 'Contact support' })
  }

  return redirectWithSnackbar(
    request,
    `${route('/collect/card')}?token=${tokenResponse.token}`,
    {
      message: 'Personal details provided.',
      icon: 'close'
    }
  )
}
