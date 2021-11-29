import {
  ApolloClient,
  HttpLink,
  InMemoryCache,
  from
} from '@apollo/client'
import { onError } from '@apollo/client/link/error'

const APOLLO_URL =
  process.env.NEXT_PUBLIC_APOLLO_PUBLIC || 'http://fynbos.test/graphql'
const APOLLO_URL_SERVER =
  process.env.NEXT_PUBLIC_APOLLO_SERVER_PUBLIC || 'http://backend/graphql'

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
  uri: typeof window === 'undefined' ? APOLLO_URL_SERVER : APOLLO_URL
})

export const apolloClient = new ApolloClient({
  cache: new InMemoryCache(),
  link: from([errorLink, Link])
})
