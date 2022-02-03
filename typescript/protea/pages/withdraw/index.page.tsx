import React from 'react'
import {
  NextPage,
  InferGetServerSidePropsType,
  GetServerSideProps,
  GetServerSidePropsResult
} from 'next'
import { Routes, WalletLayout } from 'components'
import { getSessionOrRedirect } from 'lib/kratos'

const WithdrawPage: NextPage<
  InferGetServerSidePropsType<typeof getServerSideProps>
> = () => {
  return (
    <WalletLayout backRoute={Routes.transact} header='Withdraw' hideNav>
      {/* TODO insert content */}
    </WalletLayout>
  )
}

export default WithdrawPage

export const getServerSideProps: GetServerSideProps = async (
  context
): Promise<GetServerSidePropsResult<any>> => {
  const session = await getSessionOrRedirect(context, false)
  if (session && 'redirect' in session) {
    return session
  }

  return {
    props: {}
  }
}
