import { FC } from 'react'
import Link from 'next/link'
import { Logo } from './Logo'

export const Header: FC = () => {
  return (
    <header className='sm:sticky top-0 flex justify-start items-center p-4 sm:p-8 bg-white dark:bg-gray-900 z-50'>
      <div>
        <Link href='/'>
          <a aria-label='Fynbos logo'>
            <Logo className='h-8 sm:h-12' />
          </a>
        </Link>
      </div>
    </header>
  )
}
