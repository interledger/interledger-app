import { HttpMethod } from '../types'

export const useCardProcessorApi = (baseUrl?: string) => {
  const api = {
    cards: {
      getSensitiveData: ({
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

      getPin: ({
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
      }
    }
  }

  return {
    cardProcessorClient: api
  }
}
