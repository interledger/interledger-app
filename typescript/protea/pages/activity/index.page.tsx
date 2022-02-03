import React from 'react'
import {
  NextPage,
  InferGetServerSidePropsType,
  GetServerSideProps,
  GetServerSidePropsResult
} from 'next'
import { FilterIcon, Routes, WalletLayout } from 'components'
import { getSessionOrRedirect } from 'lib/kratos'

const ActivityPage: NextPage<
  InferGetServerSidePropsType<typeof getServerSideProps>
> = () => {
  return (
    <WalletLayout
      backRoute={Routes.walletHome}
      header='Activity'
      hideNav
      actionButton={{
        text: 'Filter',
        route: Routes.activityFilter,
        icon: <FilterIcon />
      }}
    >
      {/* TODO insert content */}
    </WalletLayout>
  )
}

export default ActivityPage

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
