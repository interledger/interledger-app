import { FC } from 'react'
import { Logo } from './Logo'
import { Router, Routes } from './Routes'

export const Footer: FC = () => {
  return (
    <footer className='grid grid-cols-2 sm:grid-cols-4 gap-8 justify-start items-start p-4 sm:p-8 mb-12 mt-60 2xl:mt-80'>
      <div className='flex col-span-2 h-20 sm:h-40 sm:justify-between flex-col'>
        <Router href={Routes.home} aria-label='Fynbos logo'>
          <Logo className='h-6 sm:h-8 mb-2' />
        </Router>
        <span>&copy; Fynbos</span>
      </div>
      <div className='flex sm:h-40 flex-col space-y-2'>
        <span className='font-display font-medium mb-3'>Ecosystem</span>
        <Router href={Routes.interledger} aria-label='Interledger'>
          Interledger
        </Router>
        <Router href={Routes.openPayments} aria-label='Open Payments'>
          Open Payments
        </Router>
      </div>
      <div className='flex sm:h-40 flex-col space-y-2'>
        <span className='font-display font-medium mb-3'>Resources</span>
        <Router href={Routes.email} aria-label='Contact us'>
          Contact us
        </Router>
      </div>
    </footer>
  )
}
