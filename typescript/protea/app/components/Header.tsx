import { FC } from 'react'
import { Logo } from './Logo'
import { Router, Routes } from './Routes'

export const Header: FC = () => {
  return (
    <header className='top-0 z-50 flex items-center justify-start bg-white p-4 sm:sticky sm:p-8'>
      <div>
        <Router href={Routes.home} aria-label='Fynbos logo'>
          <Logo className='h-8 sm:h-12' />
        </Router>
      </div>
    </header>
  )
}
