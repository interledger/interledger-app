import { NextPage } from 'next'
import { useRouter } from 'next/router'
import { Container, Routes } from 'components'
import { useSession } from 'hooks'
import { LoginForm } from './loginForm'

const LoginPage: NextPage = () => {
  const session = useSession()
  const router = useRouter()

  // existing session
  if (session) {
    router.push(Routes.organisation)
    return null
  }

  // no session so show login form
  return (
    <div className='relative overflow-hidden w-full'>
      <Container className=' overflow-x-hidden'>
        <LoginForm />
      </Container>
    </div>
  )
}

export default LoginPage
