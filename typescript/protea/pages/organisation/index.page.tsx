import {
  GetServerSideProps,
  GetServerSidePropsResult,
  InferGetServerSidePropsType,
  NextPage
} from 'next'
import React, { FC } from 'react'
import { Router, Routes, Logo } from 'components'
import { getSessionOrRedirect } from 'lib/kratos'
import { Session } from '@ory/kratos-client'
import { apolloClient } from 'lib/apollo'
import {
  GetOrganisationsQuery,
  GetOrganisationsQueryVariables,
  GetOrganisationsDocument,
  Organisation
} from 'generated/types'
import { CreateOrganisationForm } from './CreateOrganisationForm'

const OverviewPage: NextPage<
  InferGetServerSidePropsType<typeof getServerSideProps>
> = ({ organisations }) => {
  return (
    <main className='mx-auto flex h-screen max-w-sm flex-col items-start justify-center px-4'>
      <Router href={Routes.home} aria-label='Fynbos logo'>
        <Logo className='h-8' />
      </Router>
      <h1 className='mt-6 mb-1 font-display text-4xl font-medium leading-normal text-strong'>
        {organisations.length == 0 && 'Create an organisation'}
        {organisations.length > 0 && 'Your organisations'}
      </h1>
      <p className='mb-10 text-medium'>
        {organisations.length == 0 &&
          'Create a new organisation to get started.'}
        {(organisations.length == 1 || organisations.length == 2) &&
          'Select an organisation, or create a new one.'}
        {organisations.length > 2 && 'Select an organisation.'}
      </p>
      {organisations.length > 0 && (
        <NavList>
          {organisations.map((option: Organisation) => {
            return (
              <NavListItem
                key={option.id}
                orgId={option.id}
                pathname={Routes.organisationOverview}
              >
                {option.name}
              </NavListItem>
            )
          })}
        </NavList>
      )}
      {organisations.length < 3 && <CreateOrganisationForm />}
    </main>
  )
}

export default OverviewPage

type OverviewPageProps = {
  session: Session
  organisations: GetOrganisationsQuery['organisations']
}

export const getServerSideProps: GetServerSideProps = async (
  context
): Promise<GetServerSidePropsResult<OverviewPageProps>> => {
  const session = await getSessionOrRedirect(context, true)
  if ('redirect' in session) {
    return session
  }

  let data
  try {
    data = await apolloClient
      .query<GetOrganisationsQuery, GetOrganisationsQueryVariables>({
        query: GetOrganisationsDocument,
        context: {
          headers: {
            cookie: context.req?.headers.cookie
          }
        }
      })
      .then((val) => val.data)
  } catch (e) {
    data = { organisations: [] }
  }

  return {
    props: {
      session: session,
      organisations: data.organisations
    }
  }
}

const NavList: FC = ({ children }) => {
  return (
    <ul className='mb-12 flex min-w-full flex-col space-y-2'>{children}</ul>
  )
}

type NavListItemProps = {
  icon?: React.ReactNode
  pathname: Routes
  orgId?: string
  route?: Routes
}

const NavListItem: FC<NavListItemProps> = ({
  children,
  icon,
  route,
  pathname,
  orgId
}) => {
  const href =
    typeof orgId == 'string'
      ? { pathname: pathname, query: { orgId: orgId } }
      : pathname
  return (
    <Router href={href}>
      <li
        className={`flex h-12 cursor-pointer items-center justify-start p-2 hover:bg-container-hover ${
          route == pathname ? 'text-primary' : 'text-medium'
        }`}
      >
        {icon && <div className='mr-2'>{icon}</div>}
        {children}
      </li>
    </Router>
  )
}
