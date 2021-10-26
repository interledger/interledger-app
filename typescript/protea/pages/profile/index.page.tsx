import { useRouter } from 'next/router'
import { Container, Routes } from 'components'
import { useSession } from 'hooks'
import { ProfileForm } from './profileForm'

const SettingsPage = () => {
  const session = useSession()
  const router = useRouter()

  if (!session) {
    router.push(Routes.login)
    return null
  }

  return (
    <div className='relative overflow-hidden w-full'>
      <Container className=' overflow-x-hidden'>
        <ProfileForm />
      </Container>
    </div>
  )
}

export default SettingsPage
