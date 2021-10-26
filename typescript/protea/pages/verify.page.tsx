import type { NextPage } from 'next'
import Head from 'next/head'
import { Container } from 'components'
import { useVerifyFlow } from 'hooks'

const VerifyPage: NextPage = () => {
  useVerifyFlow()

  // TODO: Kratos will return validation errors such as password has been pwned in the flow data.
  // our frontend will need to display these errors as returned from Kratos

  return (
    <>
      <Head>
        <title>Verify your account</title>
      </Head>
      {/* TODO: Extract this into the Container component to handle these situation. */}
      <div className='relative overflow-hidden w-full'>
        <Container className=' overflow-x-hidden'>
          Verifying account...
        </Container>
      </div>
    </>
  )
}

export default VerifyPage
