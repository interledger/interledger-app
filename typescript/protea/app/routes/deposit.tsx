import type { LoaderFunctionArgs } from '@remix-run/node'
import { json } from '@remix-run/node'
import { useLoaderData } from '@remix-run/react'
import { route } from 'routes-gen'
import type { ApplicationProps } from '~/components'
import {
  Alert,
  AlertBody,
  AlertContent,
  AlertTitle,
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  Icon,
  Layouts
} from '~/components'
import { isConnectError } from '~/lib/error.server'
import { grpc } from '~/lib/grpc.server'

export async function loader({ request }: LoaderFunctionArgs) {
  const balanceResponse = await grpc.getXagoBalances(request, {})
  if (isConnectError(balanceResponse)) throw balanceResponse.error

  const balanceAccount = balanceResponse.balances.find(
    (balance) => balance.currency == 'ZAR'
  )
  if (!balanceAccount) throw json({}, { status: 404 })

  let details = await grpc.getXagoDepositDetails(request, {
    linkedAccount: balanceAccount.linkedAccount
  })
  if (isConnectError(details)) throw details.errorResponse

  let ret = details.details.filter((d) => d.currency == balanceAccount.currency)

  return json({
    depositDetails: ret[0]
  })
}

export const handle: ApplicationProps = {
  layout: Layouts.Focus,
  scaffold: {
    header: {
      back: route('/accounts'),
      title: 'Deposit to balance'
    }
  }
}

export default function Page() {
  const { depositDetails } = useLoaderData<typeof loader>()

  return (
    <>
      <Card>
        <CardHeader>
          <CardTitle>EFT details</CardTitle>
        </CardHeader>
        <CardContent>
          <span className='mt-4'>Arrives 1-2 business days.</span>
          <div className='mt-4 flex w-full flex-col justify-between space-y-1'>
            <span className='text-weak'>Bank</span>
            <span className='text-medium'>{depositDetails.bankName}</span>
          </div>
          <div className='mt-4 flex w-full flex-col space-y-1'>
            <span className='text-weak'>Branch code</span>
            <span className='text-medium'>{depositDetails.branchCode}</span>
          </div>
          <div className='mt-4 flex w-full flex-col space-y-1'>
            <span className='text-weak'>Account number</span>
            <span className='text-medium'>{depositDetails.accountNumber}</span>
          </div>
          <div className='mt-4 flex w-full flex-col space-y-1'>
            <span className='text-weak'>Reference</span>
            <span className='text-medium'>
              {depositDetails.depositReference}
            </span>
          </div>
        </CardContent>
      </Card>
      <Alert>
        <Icon>error</Icon>
        <AlertContent className='items-start'>
          <AlertTitle>Important</AlertTitle>
          <AlertBody>
            Use the reference above when depositing for secure and faster
            processing.
          </AlertBody>
        </AlertContent>
      </Alert>
    </>
  )
}
