import {
  ApolloClient,
  HttpLink,
  ApolloLink,
  InMemoryCache,
  from
} from '@apollo/client'
import getConfig from 'next/config'
import { onError } from '@apollo/client/link/error'


const { publicRuntimeConfig } = getConfig()

const APOLLO_URL = publicRuntimeConfig.apolloClient
const APOLLO_URL_SERVER = publicRuntimeConfig.apolloServer

const errorLink = onError(({ graphQLErrors, networkError }) => {
  if (graphQLErrors)
    graphQLErrors.forEach(({ message, locations, path }) =>
      console.log(
        `[GraphQL error]: Message: ${message}, Location: ${locations}, Path: ${path}`
      )
    )

  if (networkError) console.log(`[Network error]: ${networkError}`)
})

const previewMiddleware = new ApolloLink((operation, forward) => {
  const context = operation.getContext()
  operation.setContext(({ headers = {} }) => ({
    headers: {
      ...headers,
      'fynbos-preview': context.preview || true
    }
  }))

  return forward(operation)
})

const Link = new HttpLink({
  uri: typeof window === 'undefined' ? APOLLO_URL_SERVER : APOLLO_URL
})

export const apolloClient = new ApolloClient({
  cache: new InMemoryCache(),
  link: from([errorLink, previewMiddleware, Link])
})
