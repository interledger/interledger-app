import { HttpMethod } from './cards/types'

export const cardProcessorClient = {
  card: {
    getSensitiveData: async ({
      jwtToken,
      cardProcessorUrl,
      httpMethod
    }: {
      jwtToken: string
      cardProcessorUrl: string
      httpMethod: HttpMethod
    }) => {
      if (!jwtToken) throw new Error('JWT token required')

      return fetch(cardProcessorUrl, {
        method: httpMethod,
        headers: {
          Authorization: `Bearer ${jwtToken}`
        }
      }).then(async (res) => (await res.json()) as { cypher: string })
    },

    getPin: async ({
      jwtToken,
      cardProcessorUrl,
      httpMethod
    }: {
      jwtToken: string
      cardProcessorUrl: string
      httpMethod: HttpMethod
    }) => {
      if (!jwtToken) throw new Error('JWT token required')

      return fetch(cardProcessorUrl, {
        method: httpMethod,
        headers: {
          Authorization: `Bearer ${jwtToken}`
        }
      }).then(async (res) => (await res.json()) as { cypher: string })
    },

    changePin: async ({
      newPinCypher,
      jwtToken,
      cardProcessorUrl,
      httpMethod
    }: {
      newPinCypher: string
      jwtToken: string
      cardProcessorUrl: string
      httpMethod: HttpMethod
    }) => {
      if (!jwtToken) throw new Error('JWT token required')
      if (!newPinCypher) throw new Error('New PIN cypher required')

      return fetch(cardProcessorUrl, {
        method: httpMethod,
        headers: {
          Authorization: `Bearer ${jwtToken}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({ cypher: newPinCypher })
      })
    }
  }
}
