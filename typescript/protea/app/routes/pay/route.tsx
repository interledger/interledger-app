import type { ActionArgs, LoaderArgs, MetaFunction } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { Form, useFetcher } from '@remix-run/react'
import { route } from 'routes-gen'
import type { ApplicationProps } from '~/components'
import { Layouts } from '~/components'
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
import { PayStep, useStore } from '~/lib/useStore'
import {
  getKycStatus,
  getWalletContacts,
  getWalletInfo
} from '~/lib/wallet.server'
import { KycStatus } from '~/routes/_index/route'
import { SearchPage } from '~/routes/pay/searchPage'

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
  const fetcher = useFetcher()

  const step = useStore((state) => state.step)

  return (
    <>
      <fetcher.Form
        id='pay-form'
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
      {step === PayStep.SEARCH && <SearchPage />}
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
