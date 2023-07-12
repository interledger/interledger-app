import type { ActionArgs, LoaderArgs, MetaFunction } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import {
  Form,
  useActionData,
  useFetcher,
  useLoaderData
} from '@remix-run/react'
import type { ChangeEventHandler } from 'react'
import { useCallback } from 'react'
import { route } from 'routes-gen'
import type { ApplicationProps } from '~/components'
import {
  Avatar,
  Card,
  CardButton,
  CardContent,
  CardHeader,
  CardTitle,
  FynbosIcon,
  Icon,
  Layouts,
  TextField,
  TwitterIcon
} from '~/components'
import { flowType, requireFlow, updateFlow } from '~/lib/flows.server'
import type { GrpcError } from '~/lib/proto.server'
import {
  StatusError,
  grpcClient,
  httpMapping,
  isGrpcError
} from '~/lib/proto.server'
import { generateQR, qrSvg } from '~/lib/qr.server'
import { flashSnackbar } from '~/lib/snackbar.server'
import {
  getKycStatus,
  getWalletContacts,
  getWalletInfo
} from '~/lib/wallet.server'
import { KycStatus } from '~/routes/_index/route'

export async function loader({ request }: LoaderArgs) {
  const flow = await requireFlow(request, flowType.Pay)
  const walletInfo = await getWalletInfo(request)
  const { kycStatus } = await getKycStatus(request)

  if (kycStatus != KycStatus.Verified)
    return redirect(route('/personal-details'))

  const paymentPointerQR = qrSvg(await generateQR(walletInfo.url))

  const contacts = (
    await getWalletContacts(request, {
      pageSize: 3,
      orderBy: 'last_paid_at desc'
    })
  ).contacts

  return json({ contacts, flow, paymentPointerQR })
}

export const handle: ApplicationProps = {
  layout: Layouts.Focus,
  scaffold: {
    header: { title: 'Pay', back: route('/') }
  }
}

export const meta: MetaFunction = () => {
  return {
    title: 'Pay'
  }
}

export default function Page() {
  const { contacts, flow } = useLoaderData<typeof loader>()
  const fetcher = useFetcher()
  const actionData = useActionData<typeof action>()
  const _onChangeInput = useCallback<ChangeEventHandler<HTMLInputElement>>(
    (event) => {
      let term = event.target.value
      console.log(term)
      fetcher.submit({ term: term, formName: 'search' }, { method: 'post' })
    },
    [fetcher]
  )
  return (
    <>
      <fetcher.Form
        id='search-form'
        action={route('/pay')}
        method='post'
        className='hidden'
      />
      <Form
        id='pay-address'
        action={route('/pay')}
        method='post'
        className='hidden'
      />
      <Card>
        <CardContent>
          <TextField
            id='search'
            form='search-form'
            name='search'
            placeholder='Search for a user to pay'
            onChange={_onChangeInput}
            prefixIcon={<Icon>search</Icon>}
            type='text'
          />
        </CardContent>
      </Card>
      <Card>
        <CardHeader>
          <CardTitle>Results</CardTitle>
        </CardHeader>
        {fetcher.data &&
          fetcher.data.results.map((result: any) => {
            return (
              <CardButton
                key={result.walletID}
                name='address'
                // form='pay-address'
                // value={contact.paymentPointer}
                className='items-center space-x-3'
              >
                {/*<Avatar index={index}>{contact.name.charAt(0)}</Avatar>*/}
                {result.identifierType == 'wallet' && <FynbosIcon />}
                {result.identifierType == 'twitter' && <TwitterIcon />}
                <span className='text-medium'>{result.identifier}</span>
              </CardButton>
            )
          })}
        {contacts.length == 0 && (
          <CardContent>
            <p className='text-sm text-medium'>You haven't paid anyone yet.</p>
          </CardContent>
        )}
        {contacts.map((contact, index) => (
          <CardButton
            key={contact.id}
            name='address'
            form='pay-address'
            value={contact.paymentPointer}
            className='items-center space-x-3'
          >
            <Avatar index={index}>{contact.name.charAt(0)}</Avatar>
            <span className='text-medium'>{contact.name}</span>
          </CardButton>
        ))}
      </Card>
    </>
  )
}

// The field names given by the backend for field violations
type fieldErrorsMap = 'url'

function mapper(field: fieldErrorsMap): 'address' | null {
  switch (field) {
    case 'url':
      return 'address'
    default:
      return null
  }
}

export async function action({ request }: ActionArgs) {
  await requireFlow(request, flowType.Pay)
  const form = await request.formData()
  const formName = (await form.get('formName')) as string
  const term = form.get('term') as string
  const walletID = form.get('walletID') as string
  const identifier = form.get('identifier') as string
  const identifierType = form.get('identifierType') as string

  const fieldErrors = {
    form: '',
    address: ''
  }

  switch (formName) {
    case 'search':
      const response = await grpcClient
        .searchWallets(
          { term },
          {
            meta: {
              cookies: String(request.headers.get('cookie')) || ''
            }
          }
        )
        .then((v) => v)
        .catch(StatusError)

      if (isGrpcError(response)) {
        if (response.code == 3) {
          for (let violation of (response as GrpcError).details[0]
            .fieldViolations) {
            const field = mapper(violation.field as fieldErrorsMap)
            if (field != null) fieldErrors[field] = violation.description
          }
          return json(
            { results: [], errors: { ...fieldErrors } },
            { status: 400 }
          )
        } else if (response.code == 5) {
          fieldErrors.address = 'Wallet address not found.'
          return json(
            { results: [], errors: { ...fieldErrors } },
            { status: 400 }
          )
        } else throw json({}, httpMapping(response.code))
      }
      console.log('search results', response.response.results)
      return json(
        { results: response.response.results.filter((v) => v.canSend) },
        {
          headers: {
            'Set-Cookie': await flashSnackbar(request, {
              message: 'Linked identity verification started.',
              icon: 'close'
            })
          }
        }
      )

    case 'submit':
      await updateFlow(request, flowType.Pay, {
        address: { walletID, identifier, identifierType }
      })
      return redirect(route('/pay/amount'))
    default:
      throw json(
        { title: "Submitted a form that doesn't exist" },
        {
          status: 400
        }
      )
  }
}
