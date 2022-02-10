import React from 'react'
import {
  NextPage,
  InferGetServerSidePropsType,
  GetServerSideProps,
  GetServerSidePropsResult
} from 'next'
import { Routes, WalletLayout } from 'components'
import { getSessionOrRedirect } from 'lib/kratos'

const ActivityTransactionPage: NextPage<
  InferGetServerSidePropsType<typeof getServerSideProps>
> = () => {
  return (
    <WalletLayout backRoute={Routes.activity} header='Transaction' hideNav>
      {/* TODO insert content */}
    </WalletLayout>
  )
}

export default ActivityTransactionPage

export const getServerSideProps: GetServerSideProps = async (
  context
): Promise<GetServerSidePropsResult<any>> => {
  const session = await getSessionOrRedirect(context, true)
  if (session && 'redirect' in session) {
    return session
  }

  return {
    props: {}
  }
}
