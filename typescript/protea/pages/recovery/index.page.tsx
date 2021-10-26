import { useRouter } from 'next/router'
import { NextPage } from 'next'
import { useSession } from 'hooks'
import { RecoveryForm } from './recoveryForm'
import { Container, Routes } from 'components'

const RecoveryPage: NextPage = () => {
  const session = useSession()
  const router = useRouter()

  if (session) {
    // TODO: user has valid session. what should we do?
    // Kratos will return error nodes if there is a valid session
    // and user is trying to recover.
    router.push(Routes.organisation)
    return null
  }

  return (
    <div className='relative overflow-hidden w-full'>
      <Container className=' overflow-x-hidden'>
        <RecoveryForm />
      </Container>
    </div>
  )
}

export default RecoveryPage
