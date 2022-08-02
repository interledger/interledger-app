import { Outlet } from '@remix-run/react'
import { route } from 'routes-gen'
import { Logo } from '../Logo'
import { Router } from '../Routes'

export function LandingLayout() {
  return (
    <div className='relative w-full bg-app'>
      <header className='sticky top-0 z-50 flex h-24 w-full items-center border-b border-slate-200 bg-app'>
        <div className='mx-auto flex w-full justify-between px-4 sm:max-w-lg sm:px-0 lg:max-w-3xl xl:max-w-4xl'>
          <div className='flex items-center'>
            <Router to={route('/')} aria-label='Fynbos logo'>
              <Logo className='h-7' />
            </Router>
            <div className='flex space-x-10 pt-3 pb-2 pl-10'>
              <span className='font-sans text-sm font-medium'>
                What is a payment pointer?
              </span>
              <span className='font-sans text-sm font-medium'>About</span>
              <span className='font-sans text-sm font-medium'>Legal</span>
              <span className='font-sans text-sm font-medium'>Contact</span>
            </div>
          </div>
          <div className='flex items-center'>
            <div className='flex space-x-10 pt-3 pb-2'>
              <span className='font-sans text-sm font-medium'>Log in</span>
              <span className='font-sans text-sm font-medium'>Sign up</span>
            </div>
          </div>
        </div>
      </header>
      <Outlet />
      <footer className='w-full bg-[#0B2045]'>
        <section className='mx-auto grid w-full grid-cols-4 content-start gap-4 gap-y-2 overflow-x-visible sm:max-w-lg sm:grid-cols-8 sm:px-0 lg:max-w-3xl lg:grid-cols-12 xl:max-w-4xl'>
          <div className='relative col-span-full h-20'>
            <div className='absolute top-0 right-36 h-20 w-20 rounded-full bg-slate-600' />
            <div className='absolute top-0 -right-4 h-20 w-20 rounded-bl-full bg-slate-600' />
            <div className='absolute top-0 -right-24 h-20 w-20 rounded-tr-full bg-slate-600' />
            <div className='absolute top-0 -right-44 h-20 w-20 rounded-tr-full bg-slate-600' />
          </div>
          <div className='relative col-span-3 mt-10 h-20'>
            <div className='absolute top-0 left-0 h-5 w-5 rounded-full bg-rose-200' />
            <div className='absolute top-0 left-5 h-5 w-5 rounded-tl-full bg-lime-300' />
            <div className='absolute top-0 left-10 h-5 w-5 rounded-full bg-rose-400' />
            <div className='absolute top-0 left-[3.75rem] h-5 w-5 rounded-bl-full bg-green-500' />
            <div className='absolute top-0 left-20 h-5 w-5 rounded-br-full bg-lime-300' />
            <div className='absolute top-0 left-[6.25rem] h-5 w-5 rounded-br-full bg-yellow-100' />
            <Logo className='absolute top-8 h-8 text-white' />
          </div>
          <div className='col-span-3 col-start-4 mt-10 flex flex-col space-y-1'>
            <span className='font-sans text-sm font-medium text-white'>
              Menu
            </span>
            <span className='pt-1.5 font-sans text-sm font-normal text-white'>
              What is a payment pointer?
            </span>
            <span className='font-sans text-sm font-normal text-white'>
              About
            </span>
            <span className='font-sans text-sm font-normal text-white'>
              Legal
            </span>
            <span className='font-sans text-sm font-normal text-white'>
              Contact
            </span>
          </div>
          <div className='col-span-3 col-start-7 mt-10 flex flex-col space-y-1'>
            <span className='font-sans text-sm font-medium text-white'>
              Ecosystem
            </span>
            <span className='pt-1.5 font-sans text-sm font-normal text-white'>
              Interledger Foundation
            </span>
            <span className='font-sans text-sm font-normal text-white'>
              Web monetization
            </span>
            <span className='font-sans text-sm font-normal text-white'>
              Open Payments
            </span>
          </div>
          <div className='col-span-3 col-start-10 mt-10 flex flex-col space-y-1'>
            <span className='font-sans text-sm font-medium text-white'>
              Resources
            </span>
            <span className='pt-1.5 font-sans text-sm font-normal text-white'>
              Blog
            </span>
            <span className='font-sans text-sm font-normal text-white'>
              Privacy policy
            </span>
            <span className='font-sans text-sm font-normal text-white'>
              Terms of use
            </span>
          </div>
          <div className='col-span-6 col-start-4 mt-8 flex flex-col'>
            <span className='font-sans text-sm font-normal text-slate-300'>
              &copy; 2022 Fynbos Technologies Ltd.
            </span>
          </div>
          <div className='col-span-3 col-start-10 mt-8 flex flex-col'>
            <span className='font-sans text-sm font-medium text-white'>
              TODO Social icons
            </span>
          </div>
          <div className='col-span-full col-start-4 mt-3 mb-52 flex flex-col'>
            <span className='font-sans text-sm font-normal text-slate-300'>
              Fynbos is a financial technology company and is not a bank.
              Banking services provided by Piermont Bank; Member FDIC. The
              Fynbos Visa® Debit Card is issued by Piermont Bank pursuant to a
              license from Visa U.S.A. Inc. and may be used everywhere Visa
              debit cards are accepted.
            </span>
          </div>
        </section>
      </footer>
    </div>
  )
}
