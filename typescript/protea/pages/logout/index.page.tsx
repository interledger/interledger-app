import { NextPage } from 'next'
import { useEffect, useState } from 'react'
import { useRouter } from 'next/router'
import { kratos } from '../../lib/kratos'
import { AxiosError } from 'axios'
import { Routes } from '../../components'

const LogoutPage: NextPage = () => {
  const router = useRouter()

  useEffect(() => {
    kratos
      .createSelfServiceLogoutFlowUrlForBrowsers()
      .then(({ data }) => {
        // This is a workaround until https://github.com/ory/kratos/pull/1758 lands.
        const lo = new URL(String(data.logout_url))
        const token = String(lo.searchParams.get('token'))
        kratos
          .submitSelfServiceLogoutFlow(token)
          .then(() => router.push(Routes.login))
          .then(() => router.reload())
      })
      .catch((err: AxiosError) => {
        switch (err.response?.status) {
          case 401:
            // do nothing, the user is not logged in
            return router.push(Routes.login)
        }

        // Something else happened!
        return Promise.reject(err)
      })
  }, [router])

  // Shouldn't need to render anything.
  return null
}

export default LogoutPage
