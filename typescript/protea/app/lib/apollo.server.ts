import {
  ApolloClient,
  createHttpLink,
  InMemoryCache,
  from
} from '@apollo/client'
import type { NormalizedCacheObject } from '@apollo/client'
import { onError } from '@apollo/client/link/error'
import { setContext } from '@apollo/client/link/context'

const token = process.env.DATO_API_TOKEN || ''

const httpLink = createHttpLink({
  uri: 'https://graphql.datocms.com/'
})

const authLink = setContext((_, { headers }) => {
  return {
    headers: Object.assign(headers || {}, {
      'Content-Type': 'application/json',
      Accept: 'application/json',
      Authorization: `Bearer ${token}`,
      'X-Include-Drafts': (process.env.FYNBOS_ENV !== 'prod').toString()
    })
  }
})

const errorLink = onError((err) => {
  if (err.graphQLErrors) {
    err.graphQLErrors.map(({ extensions }) => {
      if (extensions && extensions.code === 'UNAUTHENTICATED') {
        console.error('UNAUTHENTICATED')
      }

      if (extensions && extensions.code === 'FORBIDDEN') {
        console.error('FORBIDDEN')
      }
    })
  }
})

const link = from([authLink, errorLink, httpLink])

let apolloClient: ApolloClient<NormalizedCacheObject>

declare global {
  var __apolloClient: ApolloClient<NormalizedCacheObject> | undefined
}
// this is needed because in development we don't want to restart
// the server with every change, but we want to make sure we don't
// create a new connection to the Client with every change either.
if (
  process.env.NODE_ENV === 'production' &&
  process.env.FYNBOS_ENV === 'prod'
) {
  apolloClient = new ApolloClient({
    cache: new InMemoryCache({}),
    link,
    defaultOptions: {
      query: {
        fetchPolicy: 'no-cache'
      },
      mutate: {
        fetchPolicy: 'no-cache'
      },
      watchQuery: {
        fetchPolicy: 'no-cache'
      }
    }
  })
} else {
  if (!global.__apolloClient) {
    global.__apolloClient = new ApolloClient({
      cache: new InMemoryCache({}),
      link,
      defaultOptions: {
        query: {
          fetchPolicy: 'no-cache'
        },
        mutate: {
          fetchPolicy: 'no-cache'
        },
        watchQuery: {
          fetchPolicy: 'no-cache'
        }
      }
    })
  }
  apolloClient = global.__apolloClient
}

export { apolloClient }
