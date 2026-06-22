import { href } from 'react-router'
import type { FC } from 'react'
import { Router } from './Router'
import { InterledgerWalletLogo } from './Logo'

export const Header: FC = () => {
  return (
    <header className='top-0 z-50 flex items-center justify-start bg-app p-4 sm:sticky sm:p-8'>
      <div>
        <Router to={href('/')} aria-label='Interledger Wallet logo'>
          <InterledgerWalletLogo className='h-8' />
        </Router>
      </div>
    </header>
  )
}
