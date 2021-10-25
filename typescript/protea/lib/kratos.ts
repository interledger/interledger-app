import { Configuration, V0alpha2Api } from '@ory/kratos-client'

const KRATOS_URL =
  process.env.NEXT_PUBLIC_ORY_KRATOS_PUBLIC || 'http://127.0.0.1:4433'
export const kratos = new V0alpha2Api(
  new Configuration({
    basePath: KRATOS_URL,
    baseOptions: {
      // Ensure we send credentials over CORSs
      withCredentials: true
    }
  })
)
