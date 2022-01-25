import { FC } from 'react'
import { Logo } from './Logo'
import { Router, Routes } from './Routes'

export const Footer: FC = () => {
  return (
    <footer className='mb-12 mt-60 grid grid-cols-2 items-start justify-start gap-8 p-4 sm:grid-cols-4 sm:p-8 2xl:mt-80'>
      <div className='col-span-2 flex h-20 flex-col sm:h-40 sm:justify-between'>
        <Router href={Routes.home} aria-label='Fynbos logo'>
          <Logo className='mb-2 h-6 sm:h-8' />
        </Router>
        <span>&copy; Fynbos</span>
      </div>
      <div className='flex flex-col space-y-2 sm:h-40'>
        <span className='mb-3 font-display font-medium'>Ecosystem</span>
        <Router href={Routes.interledger} aria-label='Interledger'>
          Interledger
        </Router>
        <Router href={Routes.openPayments} aria-label='Open Payments'>
          Open Payments
        </Router>
      </div>
      <div className='flex flex-col space-y-2 sm:h-40'>
        <span className='mb-3 font-display font-medium'>Resources</span>
        <Router href={Routes.blog} aria-label='Contact us'>
          Blog
        </Router>
        <Router href={Routes.email} aria-label='Contact us'>
          Contact us
        </Router>
      </div>
    </footer>
  )
}
