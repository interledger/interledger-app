import { NextPage } from 'next'
import { useRouter } from 'next/router'
import { Container, Routes } from 'components'
import { useSession } from 'hooks'
import { SignupForm } from './signupForm'

const SignupPage: NextPage = () => {
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
        <SignupForm />
      </Container>
    </div>
  )
}

export default SignupPage
