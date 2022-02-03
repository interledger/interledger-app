import React from 'react'
import {
  NextPage,
  InferGetServerSidePropsType,
  GetServerSideProps,
  GetServerSidePropsResult
} from 'next'
import { Routes, WalletLayout } from 'components'
import { getSessionOrRedirect } from 'lib/kratos'

const SettingsPage: NextPage<
  InferGetServerSidePropsType<typeof getServerSideProps>
> = () => {
  return (
    <WalletLayout
      route={Routes.settings}
      backRoute={Routes.walletHome}
      header='Settings'
      hideNav
    >
      {/* TODO insert content */}
    </WalletLayout>
  )
}

export default SettingsPage

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
