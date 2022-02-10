import React from 'react'
import {
  NextPage,
  InferGetServerSidePropsType,
  GetServerSideProps,
  GetServerSidePropsResult
} from 'next'
import { Session } from '@ory/kratos-client'
import { redirect, Routes, WalletLayout } from 'components'
import { getSessionOrRedirect } from 'lib/kratos'

const SettingsPage: NextPage<
  InferGetServerSidePropsType<typeof getServerSideProps>
> = ({ session }) => {
  return (
    <WalletLayout
      route={Routes.settings}
      backRoute={Routes.walletHome}
      header='Settings'
      hideNav
    >
      {/* TODO insert content */}
      {session?.identity.traits.email}
      &#9679;&#9679;&#9679;&#9679;&#9679;&#9679;&#9679;&#9679;
    </WalletLayout>
  )
}

export default SettingsPage

type SettingsPageProps = {
  session: Session
}

export const getServerSideProps: GetServerSideProps = async (
  context
): Promise<GetServerSidePropsResult<SettingsPageProps>> => {
  const session = await getSessionOrRedirect(context, true)
  if (session && 'redirect' in session) {
    return session
  }

  const { flow: flowId } = context.query
  if (flowId) return redirect(`${Routes.settingsPassword}?flow=${flowId}`)

  return {
    props: {
      session
    }
  }
}
