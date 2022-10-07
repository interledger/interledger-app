import type { FC } from 'react'
import { route } from 'routes-gen'
import { Logo } from './Logo'
import { Router } from './Router'

export const Footer: FC = () => {
  return (
    <footer className='mx-auto mb-12 mt-60 grid grid-cols-2 items-start justify-start gap-8 p-4 sm:max-w-lg sm:grid-cols-4 sm:p-8 lg:max-w-3xl xl:max-w-4xl 2xl:mt-80'>
      <div className='col-span-2 flex h-20 flex-col sm:h-40 sm:justify-between'>
        <Router to={route('/')} aria-label='Fynbos logo'>
          <Logo className='mb-2 h-6 sm:h-8' />
        </Router>
        <span>&copy; Fynbos</span>
      </div>
      <div className='flex flex-col space-y-2 sm:h-40'>
        <span className='mb-3 font-display font-medium'>Ecosystem</span>
        <Router to={'https://interledger.org'} aria-label='Interledger'>
          Interledger
        </Router>
        <Router
          to={'https://docs.openpayments.guide'}
          aria-label='Open Payments'
        >
          Open Payments
        </Router>
      </div>
      <div className='flex flex-col space-y-2 sm:h-40'>
        <span className='mb-3 font-display font-medium'>Resources</span>
        <Router to={route('/blog')} aria-label='Blog'>
          Blog
        </Router>
        <Router to={route('/contact')} aria-label='Contact us'>
          Contact us
        </Router>
      </div>
    </footer>
  )
}
