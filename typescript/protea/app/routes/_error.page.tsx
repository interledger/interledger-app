import { NextPage, NextPageContext } from 'next'
import React from 'react'
import { Error } from 'components'

interface ErrorPageProps {
  statusCode?: number
}

const ErrorPage: NextPage<ErrorPageProps> = ({ statusCode }) => {
  return <Error statusCode={statusCode} />
}

ErrorPage.getInitialProps = ({ res, err }: NextPageContext) => {
  const statusCode = res ? res.statusCode : err ? err.statusCode : 404
  return { statusCode }
}

export default ErrorPage
