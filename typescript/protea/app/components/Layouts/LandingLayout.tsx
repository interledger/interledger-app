import { NavLink, Outlet, useLocation } from '@remix-run/react'
import type { FC } from 'react'
import { route } from 'routes-gen'
import { Logo } from '../Logo'
import { Router } from '../Routes'

type HeaderLinkProps = {
  title: string
  to: string
}
const HeaderLink: FC<HeaderLinkProps> = ({ title, to }) => {
  return (
    <NavLink
      className='relative focus-visible:outline-2 focus-visible:outline-focus'
      to={to}
    >
      {({ isActive }) => (
        <>
          <span
            className={`font-sans text-sm font-medium ${
              isActive ? 'text-rose-600' : 'text-strong'
            }`}
          >
            {title}
          </span>
          {isActive && (
            <div className='absolute -bottom-[34px] h-0.5 w-full bg-rose-600' />
          )}
        </>
      )}
    </NavLink>
  )
}

export function LandingLayout() {
  const location = useLocation()

  const isApp =
    location.pathname.includes('flows') || location.pathname.includes('login')

  return (
    <div className='relative w-full bg-white'>
      {!isApp && (
        <header className='sticky top-0 z-50 flex h-16 w-full items-center border-b border-slate-200 bg-white lg:h-24'>
          <div className='mx-auto flex w-full px-4 sm:max-w-lg sm:px-0 lg:max-w-3xl lg:justify-between xl:max-w-4xl'>
            <div className='flex items-center'>
              <Router to={route('/')} aria-label='Fynbos logo'>
                <Logo className='h-7' />
              </Router>
              <div className='hidden space-x-10 pt-3 pb-2 pl-10 lg:flex'>
                <HeaderLink
                  to='/what-is-a-payment-pointer'
                  title='What is a payment pointer?'
                />
                <HeaderLink to='/about' title='About' />
                <HeaderLink to='/legal' title='Legal' />
                <HeaderLink to='/contact' title='Contact' />
              </div>
            </div>
            <div className='hidden items-center lg:flex'>
              <div className='flex space-x-10 pt-3 pb-2'>
                <Router to='/login'>
                  <span className='font-sans text-sm font-medium'>Log in</span>
                </Router>
                <Router to='/signup'>
                  <span className='font-sans text-sm font-medium'>Sign up</span>
                </Router>
              </div>
            </div>
          </div>
        </header>
      )}
      <Outlet />
      {!isApp && (
        <footer className='w-full overflow-hidden bg-[#0B2045]'>
          <section className='mx-auto grid w-full grid-cols-4 content-start gap-4 gap-y-2 overflow-x-visible px-4 sm:max-w-lg sm:grid-cols-8 sm:px-0 lg:max-w-3xl lg:grid-cols-12 xl:max-w-4xl'>
            <div className='relative col-span-full h-20'>
              <div className='absolute right-64 top-0 h-20 w-20 rounded-full bg-slate-600 lg:right-36' />
              <div className='absolute right-36 top-0 h-20 w-20 rounded-bl-full bg-slate-600 lg:-right-4' />
              <div className='absolute right-16 top-0 h-20 w-20 rounded-tr-full bg-slate-600 lg:-right-24' />
              <div className='absolute -right-4  top-0 h-20 w-20 rounded-tr-full bg-slate-600 lg:-right-44' />
            </div>
            <div className='relative col-span-full mt-10 h-20 lg:col-span-3'>
              <div className='absolute top-0 left-0 h-5 w-5 rounded-full bg-rose-200' />
              <div className='absolute top-0 left-5 h-5 w-5 rounded-tl-full bg-lime-300' />
              <div className='absolute top-0 left-10 h-5 w-5 rounded-full bg-rose-400' />
              <div className='absolute top-0 left-[3.75rem] h-5 w-5 rounded-bl-full bg-green-500' />
              <div className='absolute top-0 left-20 h-5 w-5 rounded-br-full bg-lime-300' />
              <div className='absolute top-0 left-[6.25rem] h-5 w-5 rounded-br-full bg-yellow-100' />
              <Logo className='absolute top-8 h-8 text-white' />
            </div>
            <div className='col-span-full mt-10 flex flex-col space-y-1 lg:col-span-3 lg:col-start-4'>
              <span className='font-sans text-sm font-medium text-white'>
                Menu
              </span>
              <Router to='/what-is-a-payment-pointer'>
                <span className='pt-1.5 font-sans text-sm font-normal text-white'>
                  What is a payment pointer?
                </span>
              </Router>
              <Router to='/about'>
                <span className='font-sans text-sm font-normal text-white'>
                  About
                </span>
              </Router>
              <Router to='/legal'>
                <span className='font-sans text-sm font-normal text-white'>
                  Legal
                </span>
              </Router>
              <Router to='/contact'>
                <span className='font-sans text-sm font-normal text-white'>
                  Contact
                </span>
              </Router>
            </div>
            <div className='col-span-full mt-10 flex flex-col space-y-1 lg:col-span-3 lg:col-start-7'>
              <span className='font-sans text-sm font-medium text-white'>
                Ecosystem
              </span>
              <Router.a to='https://interledger.org/'>
                <span className='pt-1.5 font-sans text-sm font-normal text-white'>
                  Interledger Foundation
                </span>
              </Router.a>
              <Router.a to='https://webmonetization.org/'>
                <span className='font-sans text-sm font-normal text-white'>
                  Web monetization
                </span>
              </Router.a>
              <Router.a to='https://docs.openpayments.guide/'>
                <span className='font-sans text-sm font-normal text-white'>
                  Open Payments
                </span>
              </Router.a>
            </div>
            <div className='col-span-full mt-10 flex flex-col space-y-1 lg:col-span-3 lg:col-start-10'>
              <span className='font-sans text-sm font-medium text-white'>
                Resources
              </span>
              <Router to='/blog'>
                <span className='pt-1.5 font-sans text-sm font-normal text-white'>
                  Blog
                </span>
              </Router>
              <Router to='/privacy-policy'>
                <span className='font-sans text-sm font-normal text-white'>
                  Privacy policy
                </span>
              </Router>
              <Router to='/terms-of-use'>
                <span className='font-sans text-sm font-normal text-white'>
                  Terms of use
                </span>
              </Router>
            </div>
            <div className='col-span-full mt-8 flex flex-col lg:col-span-6 lg:col-start-4'>
              <span className='font-sans text-sm font-normal text-slate-300'>
                &copy; 2022 Fynbos Technologies Ltd.
              </span>
            </div>
            <div className='col-span-full mt-8 flex flex-col lg:col-span-3 lg:col-start-10'>
              <span className='font-sans text-sm font-medium text-white'>
                TODO Social icons
              </span>
            </div>
            <div className='col-span-full mt-3 mb-20 flex flex-col lg:col-span-full lg:col-start-4 lg:mb-52'>
              <span className='font-sans text-sm font-normal text-slate-300'>
                Fynbos is a financial technology company and is not a bank.
                Banking services provided by Piermont Bank; Member FDIC. The
                Fynbos Visa&reg; Debit Card is issued by Piermont Bank pursuant
                to a license from Visa U.S.A. Inc. and may be used everywhere
                Visa debit cards are accepted.
              </span>
            </div>
          </section>
        </footer>
      )}
    </div>
  )
}
