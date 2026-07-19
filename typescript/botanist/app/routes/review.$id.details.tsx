import type { LoaderFunctionArgs } from 'react-router'
import { data, href, Form, useLoaderData, useSubmit } from 'react-router'
import {
  GetLinkedAccount,
  GetReview,
  GetWalletDetails
} from '~/lib/wallet.server'
import { GridCard, Icon } from '~/components'
import { LinkedAccountReviewState } from '~/lib/types'

export async function loader({ request, params }: LoaderFunctionArgs) {
  const review = await GetReview(request, params.id as string)
  const linkedAccount = await GetLinkedAccount(
    request,
    review.linkedAccountID as string
  )
  const wallet = await GetWalletDetails(request, review.walletID as string)

  return data({
    wallet,
    review,
    linkedAccount
  })
}

export default function Page() {
  const { review, wallet, linkedAccount } = useLoaderData<typeof loader>()
  const submit = useSubmit()

  return (
    <>
      <Form
        id='features-form'
        action={href('/review/:id/details', { id: review.id })}
        method='post'
        className='hidden'
      />
      <GridCard
        className='col-span-full lg:col-span-4'
        title='Review'
        options={review}
      />
      <GridCard
        className='col-span-full lg:col-span-4'
        title='Wallet'
        options={wallet}
      />
      <GridCard
        className='col-span-full lg:col-span-4'
        title='Linked Account'
        options={linkedAccount}
      />
      <div
        className='whitespace-nowrap p-4 text-sm text-gray-500'
        onClick={() => {
          let formData = new FormData()
          formData.append('reviewID', review.id)
          formData.append('newState', LinkedAccountReviewState.Verified)
          formData.append('reason', 'Manually verified.')
          submit(formData, {
            action: href('/reviews'),
            method: 'POST'
          })
        }}
      >
        <span className='flex cursor-pointer items-center space-x-2 font-medium text-primary'>
          <Icon>approval_delegation</Icon>
          <span>Approve</span>
        </span>
      </div>
      <div
        className='whitespace-nowrap p-4 text-sm text-gray-500'
        onClick={() => {
          let formData = new FormData()
          formData.append('reviewID', review.id)
          formData.append('newState', LinkedAccountReviewState.Rejected)
          formData.append('reason', 'Manually rejected.')
          submit(formData, {
            action: href('/reviews'),
            method: 'POST'
          })
        }}
      >
        <span className='flex cursor-pointer items-center space-x-2 font-medium text-primary text-red-500'>
          <Icon>close</Icon>
          <span>Reject</span>
        </span>
      </div>
    </>
  )
}
