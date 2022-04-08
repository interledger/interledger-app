import { ApolloClient, from, HttpLink, InMemoryCache } from '@apollo/client'
import { onError } from '@apollo/client/link/error'

const errorLink = onError(({ graphQLErrors, networkError }) => {
  if (graphQLErrors)
    graphQLErrors.forEach(({ message, locations, path }) =>
      console.log(
        `[GraphQL error]: Message: ${message}, Location: ${locations}, Path: ${path}`
      )
    )

  if (networkError) console.log(`[Network error]: ${networkError}`)
})

const Link = new HttpLink({
  uri: 'http://backend/graphql'
})

export const apolloClient = new ApolloClient({
  cache: new InMemoryCache(),
  link: from([errorLink, Link]),
  defaultOptions: {
    watchQuery: {
      errorPolicy: 'all',
      fetchPolicy: 'network-only'
    },
    query: {
      errorPolicy: 'all',
      fetchPolicy: 'network-only'
    }
  }
})
