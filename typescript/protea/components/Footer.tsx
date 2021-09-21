import { FC } from 'react'
import Link from 'next/link'
import { Logo } from './Logo'

export const Footer: FC = () => {
  return (
    <footer className='grid grid-cols-2 sm:grid-cols-4 gap-8 justify-start items-start p-4 sm:p-8 mb-12 mt-60 2xl:mt-80'>
      <div className='flex col-span-2 h-20 sm:h-40 sm:justify-between flex-col'>
        <Link href='/'>
          <a aria-label='Fynbos Blog'>
            <Logo className='h-6 sm:h-12 mb-2' />
          </a>
        </Link>
        <span>&copy; Fynbos</span>
      </div>
      <div className='flex sm:h-40 flex-col space-y-2'>
        <span className='font-display font-medium mb-3'>Ecosystem</span>
        <Link href='https://interledger.org'>
          <a aria-label='Interledger'>Interledger</a>
        </Link>
        <Link href='https://openpayments.dev'>
          <a aria-label='Open Payments'>Open Payments</a>
        </Link>
      </div>
      <div className='flex sm:h-40 flex-col space-y-2'>
        <span className='font-display font-medium mb-3'>Resources</span>
        <Link href='mailto:hello@fynbos.dev'>
          <a aria-label='Contact us'>Contact us</a>
        </Link>
        {/*<Link href='/blog'>*/}
        {/*  <a aria-label='Fynbos Blog'>Blog</a>*/}
        {/*</Link>*/}
      </div>
    </footer>
  )
}
