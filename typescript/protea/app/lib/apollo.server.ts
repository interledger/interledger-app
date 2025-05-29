import type { NormalizedCacheObject } from '@apollo/client'
import {
  ApolloClient,
  InMemoryCache,
  createHttpLink,
  from
} from '@apollo/client'
import { setContext } from '@apollo/client/link/context'
import { onError } from '@apollo/client/link/error'
import { captureMessage } from '@sentry/remix'

const token = process.env.DATO_API_TOKEN || ''

// Can be removed after we migrate away from the DATO responses
const httpLink = createHttpLink({
  uri: 'https://dato.fynbos.workers.dev/'
})

const authLink = setContext((_, { headers }) => {
  if (process.env.FYNBOS_ENV !== 'prod') {
    headers = Object.assign(headers || {}, {
      'X-Include-Drafts': 'true'
    })
  }
  return {
    headers: Object.assign(headers || {}, {
      'Content-Type': 'application/json',
      Accept: 'application/json',
      Authorization: `Bearer ${token}`
    })
  }
})

const errorLink = onError((err) => {
  captureMessage('Error received in apollo client', {
    extra: {
      operation: err.operation,
      graphQLErrors: err.graphQLErrors,
      networkError: err.networkError
    }
  })
  if (err.graphQLErrors) {
    err.graphQLErrors.map(({ extensions }) => {
      if (extensions && extensions.code === 'UNAUTHENTICATED') {
        console.error('UNAUTHENTICATED')
      }

      if (extensions && extensions.code === 'FORBIDDEN') {
        console.error('FORBIDDEN')
      }
      return null
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
if (process.env.NODE_ENV === 'production') {
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
