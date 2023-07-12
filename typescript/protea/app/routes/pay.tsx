import type { ActionArgs, LoaderArgs, MetaFunction } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { Form, useFetcher, useLoaderData } from '@remix-run/react'
import type { ChangeEventHandler } from 'react'
import { useCallback, useEffect, useState } from 'react'
import { route } from 'routes-gen'
import type { ApplicationProps } from '~/components'
import {
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
import type { SearchResult } from '~/generated/protobuf-ts/backend/v1/backend'
import { flowType, requireFlow, updateFlow } from '~/lib/flows.server'
import type { GrpcError } from '~/lib/proto.server'
import {
  StatusError,
  grpcClient,
  httpMapping,
  isGrpcError
} from '~/lib/proto.server'
import { generateQR, qrSvg } from '~/lib/qr.server'
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

  // Handle returning to the pay page after routing away
  let results: SearchResult[] = []
  if (flow.data.term) {
    const response = await grpcClient
      .searchWallets(
        { term: flow.data.term },
        {
          meta: {
            cookies: String(request.headers.get('cookie')) || ''
          }
        }
      )
      .then((v) => v)
      .catch(StatusError)

    if (!isGrpcError(response)) {
      results = response.response.results.filter((v) => v.canSend)
    }
  }

  return json({ results, contacts, flow, paymentPointerQR })
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
  const { flow } = useLoaderData<typeof loader>()
  const fetcher = useFetcher()
  const [term, setTerm] = useState<string>(flow.data.term || '')
  // const actionData = useActionData<typeof action>()

  const _onChangeInput = useCallback<ChangeEventHandler<HTMLInputElement>>(
    (event) => {
      let term = event.target.value
      setTerm(term)
      fetcher.submit({ term: term, formName: 'search' }, { method: 'post' })
    },
    [fetcher]
  )

  // TODO: should merge these into local state rather.
  // This will also improve the UX when returning to the page
  // and when the query is changed, and becomes shorter that the lower bound.
  useEffect(() => {
    if (fetcher.state === 'idle' && fetcher.data == null) {
      fetcher.load('/pay')
    }
  }, [fetcher])

  const _onClickResult = useCallback<{
    (result: SearchResult): void
  }>(
    (result) => {
      console.log('result', result)
      fetcher.submit(
        {
          term: term,
          walletUrl: result.walletUrl,
          identifier: result.identifier,
          identifierType: result.identifierType,
          formName: 'submit'
        },
        { method: 'post' }
      )
    },
    [fetcher, term]
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
            defaultValue={term}
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
        {(!fetcher.data || fetcher.data.results.length == 0) && (
          <CardContent>Your search returned no results.</CardContent>
        )}
        {fetcher.data &&
          fetcher.data.results.map((result: SearchResult) => {
            return (
              <CardButton
                key={result.walletID}
                onClick={() => _onClickResult(result)}
                name='address'
                type='button'
                className='items-center space-x-3'
              >
                {result.identifierType == 'wallet' && <FynbosIcon />}
                {result.identifierType == 'twitter' && <TwitterIcon />}
                <span className='text-medium'>{result.identifier}</span>
              </CardButton>
            )
          })}
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
  const walletUrl = form.get('walletUrl') as string
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
      return json({
        results: response.response.results.filter((v) => v.canSend)
      })

    case 'submit':
      console.log(
        'submit',
        formName,
        term,
        walletUrl,
        identifier,
        identifierType
      )
      await updateFlow(request, flowType.Pay, {
        term,
        address: { walletUrl, identifier, identifierType }
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
