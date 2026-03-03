import type { Route } from './+types/settings_.grants_.$grantId'
import type { PlainMessage } from '@bufbuild/protobuf'
import type { UIMatch } from 'react-router';
import { Form, useLoaderData } from 'react-router';
import { href } from 'react-router'
import type { ApplicationProps } from '~/components'
import {
  Card,
  CardContent,
  Chip,
  ChipColor,
  Layouts,
  OutlineButton
} from '~/components'
import { Label } from '~/components/Label'
import { jsonWithCSRF, validateCSRFToken } from '~/lib/csrf.server'
import { isConnectError } from '~/lib/error.server'
import { grpc } from '~/lib/grpc.server'
import type { RafikiAccess } from '~/generated/connect/backend/v1/backend_pb'
import { mergeMeta } from '~/lib/meta'
import { redirectWithSnackbar } from '~/lib/snackbar.server'

export const handle: ApplicationProps = {
  layout: Layouts.Focus,
  scaffold: {
    header: {
      back: '/settings/grants',
      title: (match: UIMatch<Awaited<ReturnType<typeof loader>>['data']>) => match.data!.grant.client,
      actions: (match: UIMatch<Awaited<ReturnType<typeof loader>>['data']>) => {
        const { grant } = match.data!
        switch (grant.state) {
          case 'GRANTED':
            return {
              key: 'Approved',
              nodes: <Chip color={ChipColor.green}>Approved</Chip>
            }
          case 'PENDING':
            return {
              key: 'Pending',
              nodes: <Chip color={ChipColor.orange}>Pending</Chip>
            }
          case 'REJECTED':
            return {
              key: 'Rejected',
              nodes: <Chip color={ChipColor.red}>Rejected</Chip>
            }
          case 'REVOKED':
            return {
              key: 'Revoked',
              nodes: <Chip color={ChipColor.red}>Revoked</Chip>
            }
          default:
            return null
        }
      }
    }
  }
}

export const meta = mergeMeta(({ data }) => {
  const d = data as Awaited<ReturnType<typeof loader>>['data'] | undefined
  return [{ title: d?.grant.client || 'Grant' }]
})

export async function loader({ request, params }: Route.LoaderArgs) {
  const grant = await grpc.getRafikiGrant(request, {
    id: params.grantId as string
  })

  if (isConnectError(grant)) throw grant.errorResponse

  return jsonWithCSRF(request, {
    grant
  })
}

export default function Page() {
  const { grant, csrfToken } = useLoaderData<typeof loader>()

  return (
    <>
      <Form id='key-id' method='post' className='hidden' />
      <input form='key-id' value={csrfToken} name='csrfToken' type='hidden' />
      <Card>
        <CardContent className='mt-2 flex flex-col gap-y-4'>
          <div className='flex w-full justify-between'>
            <span className='text-weak'>Client</span>
            <span className='text-medium'>{grant.client}</span>
          </div>
          <div className='flex w-full justify-between'>
            <span className='text-weak'>Created</span>
            <span className='text-medium'>{grant.createdAt}</span>
          </div>
        </CardContent>
      </Card>

      {grant.access.map((a: PlainMessage<RafikiAccess>) => (
        <Card key={a.id}>
          <Label>{a.type}</Label>
          <CardContent className='mt-2 flex flex-col gap-y-4'>
            <div className='flex w-full justify-between'>
              <span className='text-weak'>Identifier</span>
              <span className='text-medium'>{a.identifier}</span>
            </div>
            <div className='flex w-full justify-between'>
              <span className='text-weak'>Actions</span>
              <span className='text-medium'>
                {a.actions.reduce((prev, curr) => `${prev}, ${curr}`)}
              </span>
            </div>
            <div className='flex w-full justify-between'>
              <span className='text-weak'>Amount</span>
              <span className='text-medium'>
                {a.limits?.formattedDebitAmount}
              </span>
            </div>
          </CardContent>
        </Card>
      ))}

      <OutlineButton
        className='text-red-700 outline-red-700 focus-visible:outline-red-800'
        form='key-id'
        type='submit'
      >
        Revoke grant
      </OutlineButton>
    </>
  )
}

export async function action({ request, params }: Route.ActionArgs) {
  const form = await request.formData()

  await validateCSRFToken(request, form)

  const errors = {
    form: ''
  }

  const response = await grpc.revokeRafikiGrant(request, {
    id: params.grantId as string
  })

  if (isConnectError(response)) {
    return response.error({ errors })
  }

  return redirectWithSnackbar(request, href('/settings/grants'), {
    message: 'Grant was revoked.',
    icon: 'close'
  })
}
