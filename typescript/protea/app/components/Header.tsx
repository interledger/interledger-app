import { FC } from 'react'
import { route } from 'routes-gen'
import { Logo } from './Logo'
import { Router } from './Routes'

export const Header: FC = () => {
  return (
    <header className='top-0 z-50 flex items-center justify-start bg-white p-4 sm:sticky sm:p-8'>
      <div>
        <Router to={route('/')} aria-label='Fynbos logo'>
          <Logo className='h-8 sm:h-12' />
        </Router>
      </div>
    </header>
  )
}
