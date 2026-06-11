import { href } from 'react-router'
import type { FC } from 'react'
import { Logo } from './Logo'
import { Router } from './Router'

export const Header: FC = () => {
  return (
    <header className='top-0 z-50 flex items-center justify-start bg-app p-4 sm:sticky sm:p-8'>
      <div>
        <Router to={href('/')} aria-label='Interledger logo'>
          <Logo className='h-8 sm:h-12' />
        </Router>
      </div>
    </header>
  )
}
