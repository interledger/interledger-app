import { Routes, redirect } from 'components'
import {
  GetOrgsForDashboardQuery,
  GetOrgsForDashboardQueryVariables,
  GetOrgsForDashboardDocument
} from 'generated/types'
import { GetServerSidePropsContext, PreviewData, Redirect } from 'next'
import { ParsedUrlQuery } from 'querystring'
import { apolloClient } from './apollo'

type CurrentOrg = {
  id: string
  name: string
  verified?: boolean
}

type Organisation =
  | {
      id: string
      name: string
    }
  | undefined

export type OrgsForDashboard = {
  currentOrg: CurrentOrg | null
  organisations: Organisation[]
}

export type DashboardProps = OrgsForDashboard | { redirect: Redirect }

export async function getOrgsForDashboardOrRedirect(
  context: GetServerSidePropsContext<ParsedUrlQuery, PreviewData>
): Promise<DashboardProps> {
  let orgId = context.params?.orgId

  if (typeof orgId == 'undefined' || typeof orgId !== 'string') orgId = ''

  try {
    const data = await apolloClient
      .query<GetOrgsForDashboardQuery, GetOrgsForDashboardQueryVariables>({
        query: GetOrgsForDashboardDocument,
        variables: {
          id: orgId
        },
        context: {
          headers: {
            cookie: context.req?.headers.cookie
          }
        },
        // We allow partial results through.
        // For pages like /profile that don't relate to a specific org.
        errorPolicy: 'all'
      })
      .then((val) => val.data)

    let currentOrg: CurrentOrg | undefined
    if (data.organisation) {
      const { id, name, verified } = data.organisation
      currentOrg = { id: id, name, verified }
    }
    return {
      // Ensure currentOrg is serializable if undefined
      currentOrg: currentOrg || null,
      organisations: data.organisations.map((val: any) => {
        if (val) {
          return { id: val?.id, name: val?.name }
        }
      })
    }
  } catch (e) {
    return redirect(Routes.organisation)
  }
}
