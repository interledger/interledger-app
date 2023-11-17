import type { PlainMessage } from '@bufbuild/protobuf'
import type { LoaderFunctionArgs, MetaFunction } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { useLoaderData } from '@remix-run/react'
import { useEffect } from 'react'
import { route } from 'routes-gen'
import type { ApplicationProps } from '~/components'
import {
  Alert,
  AlertBody,
  AlertContent,
  AlertTitle,
  ButtonRouter,
  Card,
  CardButton,
  CardContent,
  Icon,
  Layouts
} from '~/components'
import { Label } from '~/components/Label'
import type { PublicWalletInfo } from '~/generated/connect/backend/v1/backend_pb'
import { isConnectError } from '~/lib/error.server'
import { grpc } from '~/lib/grpc.server'
import { mergeMeta } from '~/lib/meta'
import { useScaffoldStore } from '~/lib/useScaffoldStore'

export async function loader({ request, params }: LoaderFunctionArgs) {
  if (process.env.FYNBOS_ENV == 'prod') {
    return redirect(route('/payments'))
  }

  const transaction = await grpc.lookupTransaction(request, {
    id: params.paymentId as string
  })

  if (isConnectError(transaction)) throw transaction.errorResponse

  const walletUrl = transaction.source
  let publicWalletInfo: PlainMessage<PublicWalletInfo>

  const publicWalletInfoResponse = await grpc.getPublicWalletInfo(request, {
    walletAddress: walletUrl
  })

  if (isConnectError(publicWalletInfoResponse)) {
    publicWalletInfo = {
      walletID: 'not-found',
      address: walletUrl,
      shortAddress: '',
      publicName: '',
      identities: [],
      canReceive: false
    }
  } else publicWalletInfo = publicWalletInfoResponse

  let rpc = await grpc.createPaymentLink(request, {
    transactionId: transaction.id
  })
  if (isConnectError(rpc)) throw rpc.errorResponse

  return json({
    transaction,
    shareUrl: rpc.url,
    publicWalletInfo
  })
}

export const handle: ApplicationProps = {
  layout: Layouts.Focus,
  scaffold: {
    header: {
      back: route('/payments'),
      title: 'Share Payment'
    }
  }
}

export const meta: MetaFunction = mergeMeta(() => [
  {
    title: 'Share Payment'
  }
])

export default function Page() {
  const { transaction, shareUrl } = useLoaderData<typeof loader>()

  const [pushSnackbar] = useScaffoldStore((state) => [state.pushSnackbar])

  useEffect(() => {
    // TODO Need to check this actually works
    if (shareUrl && typeof navigator !== 'undefined')
      navigator.clipboard
        .writeText(shareUrl)
        .then(() => {})
        .catch((err) => {
          console.log('Had an error', err)
        })
  }, [shareUrl])

  return (
    <>
      <Card>
        <CardContent>
          <div className='flex items-center justify-between'>
            <h2 className='text-4xl font-medium text-error'>
              {transaction.subtotal}
            </h2>
            <div className='flex flex-col items-end space-y-1'>
              <span className='text-sm font-medium text-medium'>
                {transaction.formattedDate}
              </span>
              <span className='text-xs text-weak'>
                {transaction.formattedTime}
              </span>
            </div>
          </div>
        </CardContent>
        <Label>Share payment</Label>
        <CardButton
          noHover
          type='button'
          onClick={() => {
            if (typeof navigator.share == 'undefined') {
              navigator.clipboard.writeText(shareUrl).then(
                () => {
                  pushSnackbar({
                    id: 'copy-payment-link-success',
                    message: 'The link has been copied to your clipboard.',
                    icon: 'close',
                    canShow: true
                  })
                },
                () => {
                  pushSnackbar({
                    id: 'copy-to-clipboard-fail',
                    message: "Couldn't copy to clipboard.",
                    icon: 'close',
                    canShow: true
                  })
                }
              )
            } else navigator.share({ url: shareUrl })
          }}
          className='items-center justify-between'
        >
          <span className='truncate text-left font-medium text-medium'>
            {shareUrl}
          </span>
          <Icon className='text-medium'>share</Icon>
        </CardButton>
      </Card>
      <Alert>
        <Icon>notification_important</Icon>
        <AlertContent>
          <AlertTitle>
            Only share the payment with the intended receiver
          </AlertTitle>
          <AlertBody>
            Anyone with the link can collect the payment, therefore only share
            it with the intended receiver.
          </AlertBody>
        </AlertContent>
      </Alert>
      <ButtonRouter
        to={route('/payments/:paymentId', { paymentId: transaction.id })}
      >
        View payment details
      </ButtonRouter>
    </>
  )
}
