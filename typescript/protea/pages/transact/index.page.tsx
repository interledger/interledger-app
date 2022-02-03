import React from 'react'
import {
  NextPage,
  InferGetServerSidePropsType,
  GetServerSideProps,
  GetServerSidePropsResult
} from 'next'
import { Routes, WalletLayout } from 'components'
import { getSessionOrRedirect } from 'lib/kratos'

const TransactPage: NextPage<
  InferGetServerSidePropsType<typeof getServerSideProps>
> = () => {
  return (
    <WalletLayout route={Routes.transact} header='Transact' settings>
      {/* TODO insert content */}
    </WalletLayout>
  )
}

export default TransactPage

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
