import { FC } from 'react'
import { Logo } from './Logo'
import { Router, Routes } from './Routes'

export const Header: FC = () => {
  return (
    <header className='sm:sticky top-0 flex justify-start items-center p-4 sm:p-8 bg-white z-50'>
      <div>
        <Router href={Routes.home} aria-label='Fynbos logo'>
          <Logo className='h-8 sm:h-12' />
        </Router>
      </div>
    </header>
  )
}
