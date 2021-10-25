import { Container } from 'components'
import { NextPage } from 'next'

const ErrorPage: NextPage = () => {
  return (
    <div className='relative overflow-hidden w-full'>
      <Container className=' overflow-x-hidden'>
        An error occured...
      </Container>
    </div>
  )
}

export default ErrorPage
