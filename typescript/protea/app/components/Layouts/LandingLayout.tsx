import { NavLink, Outlet, useMatches } from '@remix-run/react'
import { DateTime } from 'luxon'
import type { FC } from 'react'
import { useState } from 'react'
import { route } from 'routes-gen'
import { IconButton } from '~/components'
import { FynbosLogo } from '../Logos'
import { AnchorRouter, ButtonRouter, Router } from '../Router'
import { NavDrawer } from './NavDrawer'

type HeaderLinkProps = {
  title: string
  to: string
}
const HeaderLink: FC<HeaderLinkProps> = ({ title, to }) => {
  return (
    <NavLink
      className='relative rounded focus-visible:outline focus-visible:outline-2 focus-visible:outline-focus'
      to={to}
    >
      {({ isActive }) => (
        <>
          <span
            className={`text-sm font-medium ${isActive && 'text-rose-600'}`}
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
  const [openNavModal, setOpenNavModal] = useState<boolean>(false)
  const matches = useMatches()
  const isUser = matches[0]?.data.isUser
  const isSignupGated = matches[0]?.data.isSignupGated
  return (
    <div className='relative flex min-h-screen w-full flex-col bg-mk-page'>
      <header className='fixed top-0 z-10 mb-16 flex h-16 w-full items-center border-b border-slate-200 bg-mk-page lg:h-24'>
        <div className='mx-auto flex w-full justify-between px-4 sm:max-w-lg sm:px-0 lg:max-w-3xl xl:max-w-[59rem]'>
          <div className='flex items-center'>
            <IconButton
              className='lg:hidden'
              onClick={() => setOpenNavModal(true)}
              aria-label='Open menu'
            >
              menu
            </IconButton>
            <div className='ml-4 lg:ml-0'>
              <Router to={route('/')} aria-label='Fynbos logo'>
                <FynbosLogo className='h-8' />
              </Router>
            </div>
            <div className='hidden space-x-10 pb-2 pl-10 pt-3 lg:flex'>
              <HeaderLink
                to={route('/what-is-a-payment-pointer')}
                title='What is a payment pointer?'
              />
              {/*<HeaderLink to={route('/about')} title='About' />*/}
              <HeaderLink to={route('/blog')} title='Blog' />
              <HeaderLink to={route('/contact')} title='Contact' />
            </div>
          </div>
          <div className='hidden items-center lg:flex'>
            {!isUser && (
              <div className='flex space-x-10 pb-2 pt-3'>
                <Router to={route('/login')}>
                  <span className='text-sm font-medium'>Log in</span>
                </Router>
                {isSignupGated && (
                  <Router to={route('/waitlist')}>
                    <span className='text-sm font-medium'>
                      Join the waitlist
                    </span>
                  </Router>
                )}
                {!isSignupGated && (
                  <Router to={route('/signup')}>
                    <span className='text-sm font-medium'>Sign up</span>
                  </Router>
                )}
              </div>
            )}
            {isUser && (
              <div className='flex items-center '>
                <ButtonRouter to={route('/')}>
                  <span className='text-sm font-medium'>Go to app</span>
                </ButtonRouter>
              </div>
            )}
          </div>
        </div>
      </header>
      {/* Spacer for fixed header */}
      <div className='mb-16 lg:mb-24' />
      <div className='grow'>
        <Outlet />
      </div>

      <footer className='z-10 w-full flex-shrink-0 overflow-hidden bg-mk-footer'>
        <section className='mx-auto grid w-full grid-cols-4 content-start gap-4 gap-y-2 overflow-x-visible px-8 sm:max-w-lg sm:grid-cols-8 sm:px-0 lg:max-w-3xl lg:grid-cols-12 xl:max-w-[59rem]'>
          <div className='relative col-span-full h-20'>
            <div className='absolute right-64 top-0 h-20 w-20 rounded-full bg-[#182C4F] lg:right-36' />
            <div className='absolute right-32 top-0 h-20 w-20 rounded-bl-full bg-[#182C4F] lg:-right-4' />
            <div className='absolute right-12 top-0 h-20 w-20 rounded-tr-full bg-[#182C4F] lg:-right-24' />
            <div className='absolute -right-8  top-0 h-20 w-20 rounded-tr-full bg-[#182C4F] lg:-right-44' />
          </div>
          <div className='relative col-span-full mt-10 h-20 lg:col-span-3'>
            <div className='absolute left-0 top-0 h-5 w-5 rounded-full bg-rose-500' />
            <div className='absolute left-5 top-0 h-5 w-5 rounded-tl-full bg-yellow-400' />
            <div className='absolute left-10 top-0 h-5 w-5 rounded-full bg-slate-500' />
            <div className='absolute left-[3.75rem] top-0 h-5 w-5 rounded-bl-full bg-lime-500' />
            <div className='absolute left-20 top-0 h-5 w-5 rounded-br-full bg-slate-300' />
            <div className='absolute left-[6.25rem] top-0 h-5 w-5 rounded-tr-full bg-slate-700' />
            <FynbosLogo className='absolute top-8 h-8 text-white' />
          </div>
          <div className='col-span-full mt-10 flex flex-col space-y-1 lg:col-span-3 lg:col-start-4'>
            <span className='pb-2 text-sm font-medium text-white'>Menu</span>
            <Router to={route('/what-is-a-payment-pointer')}>
              <span className='text-sm text-white'>
                What is a payment pointer?
              </span>
            </Router>
            {/*<Router to={route('/about')}>*/}
            {/*  <span className='text-sm text-white'>About</span>*/}
            {/*</Router>*/}
            <Router to={route('/contact')}>
              <span className='text-sm text-white'>Contact</span>
            </Router>
          </div>
          <div className='col-span-full mt-10 flex flex-col space-y-1 lg:col-span-3 lg:col-start-7'>
            <span className='pb-2 text-sm font-medium text-white'>
              Ecosystem
            </span>
            <AnchorRouter to='https://paymentpointers.org/'>
              <span className='text-sm text-white'>Payment Pointers</span>
            </AnchorRouter>
            <AnchorRouter to='https://interledger.org/'>
              <span className='text-sm text-white'>Interledger Foundation</span>
            </AnchorRouter>
            <AnchorRouter to='https://docs.openpayments.guide/'>
              <span className='text-sm text-white'>Open Payments</span>
            </AnchorRouter>
          </div>
          <div className='col-span-full mt-10 flex flex-col space-y-1 lg:col-span-3 lg:col-start-10'>
            <span className='pb-2 text-sm font-medium text-white'>
              Resources
            </span>
            <Router to={route('/blog')}>
              <span className='text-sm text-white'>Blog</span>
            </Router>
            <Router to={route('/legal')}>
              <span className='text-sm text-white'>Legal Agreements</span>
            </Router>
          </div>
          <div className='col-span-full mt-8 flex space-x-4 lg:col-span-3 lg:col-start-4'>
            <AnchorRouter to='https://mobile.twitter.com/fynbosdev'>
              <svg
                width='22'
                height='19'
                viewBox='0 0 22 19'
                fill='none'
                xmlns='http://www.w3.org/2000/svg'
              >
                <path
                  d='M21.8824 2.29828C21.0629 2.66571 20.1938 2.90688 19.3038 3.01379C20.2308 2.45142 20.9426 1.56087 21.2778 0.499801C20.3965 1.02911 19.4324 1.40205 18.427 1.60249C17.608 0.719387 16.4413 0.16748 15.15 0.16748C12.6707 0.16748 10.6605 2.20207 10.6605 4.71148C10.6605 5.06768 10.7002 5.41445 10.7767 5.74711C7.04561 5.55755 3.73761 3.74852 1.5233 0.999277C1.13694 1.67041 0.915554 2.45107 0.915554 3.28373C0.915554 4.86029 1.70819 6.25109 2.91275 7.06601C2.19982 7.04341 1.50257 6.84851 0.879226 6.49758C0.87897 6.51662 0.87897 6.53565 0.87897 6.55477C0.87897 8.75643 2.42646 10.5931 4.48016 11.0105C3.81906 11.1925 3.12561 11.2191 2.45279 11.0884C3.02404 12.8937 4.68205 14.2074 6.64651 14.2442C5.11004 15.4629 3.17422 16.1894 1.07095 16.1894C0.708527 16.1894 0.351229 16.1678 0 16.1259C1.98676 17.4152 4.34655 18.1675 6.88183 18.1675C15.1396 18.1675 19.6551 11.2433 19.6551 5.23847C19.6551 5.04137 19.6509 4.84541 19.6421 4.65057C20.5211 4.00745 21.2797 3.2109 21.8824 2.29828Z'
                  fill='white'
                />
              </svg>
            </AnchorRouter>

            <AnchorRouter to='https://www.linkedin.com/company/fynbos'>
              <svg
                width='20'
                height='18'
                viewBox='0 0 20 18'
                fill='none'
                xmlns='http://www.w3.org/2000/svg'
              >
                <g clipPath='url(#clip0_2720_8710)'>
                  <path
                    d='M5.08994 17.9986V5.85397H1.00822V17.9986H5.08994ZM3.04961 4.19484C4.47298 4.19484 5.35896 3.26228 5.35896 2.09688C5.33244 0.905213 4.47303 -0.00146484 3.07662 -0.00146484C1.68042 -0.00146484 0.767395 0.905231 0.767395 2.09688C0.767395 3.26233 1.65316 4.19484 3.02296 4.19484H3.04948H3.04961ZM7.34917 17.9986H11.4309V11.2164C11.4309 10.8535 11.4574 10.4909 11.5652 10.2314C11.8603 9.50619 12.5319 8.75508 13.6594 8.75508C15.1364 8.75508 15.7273 9.86878 15.7273 11.5014V17.9985H19.8088V11.0349C19.8088 7.30454 17.7951 5.56885 15.1096 5.56885C12.9076 5.56885 11.9408 6.78605 11.4037 7.61507H11.431V5.85372H7.34926C7.40283 6.9933 7.34926 17.9983 7.34926 17.9983L7.34917 17.9986Z'
                    fill='white'
                  />
                </g>
                <defs>
                  <clipPath id='clip0_2720_8710'>
                    <rect
                      width='19.0385'
                      height='18'
                      fill='white'
                      transform='translate(0.766998)'
                    />
                  </clipPath>
                </defs>
              </svg>
            </AnchorRouter>
          </div>
          <div className='col-span-full mt-8 flex flex-col lg:col-span-6 lg:col-start-4'>
            <span className='text-sm text-slate-300'>
              &copy;&nbsp;{DateTime.now().year} Fynbos Technologies Ltd.
            </span>
          </div>
          <div className='col-span-full mb-20 mt-8 flex flex-col lg:col-start-4'>
            <span className='text-sm text-slate-300'>
              Fynbos Technologies LLC is registered as an agent of Golden Money
              Transfer Inc. a licensed money services business with FINCEN MSB
              Registration Number: 31000231163732. All payments originated
              within the USA are processed under the appropriate state licenses.
            </span>
          </div>
        </section>
      </footer>
      <NavDrawer.Modal open={openNavModal} setOpen={setOpenNavModal}>
        <NavDrawer>
          <NavDrawer.List>
            <div className='relative mb-8 ml-1 flex items-center space-x-4'>
              <IconButton
                onClick={() => setOpenNavModal(!openNavModal)}
                aria-label='Close menu'
              >
                menu_open
              </IconButton>
              <div className='ml-4 lg:ml-0'>
                <Router to={route('/')} aria-label='Fynbos logo'>
                  <FynbosLogo className='h-8' />
                </Router>
              </div>
            </div>
            <NavDrawer.ListItem to={route('/')}>Home</NavDrawer.ListItem>
            <NavDrawer.ListItem to={route('/what-is-a-payment-pointer')}>
              What is a payment pointer?
            </NavDrawer.ListItem>
            {/*<NavDrawer.ListItem to={route('/about')}>About</NavDrawer.ListItem> */}
            <NavDrawer.ListItem to={route('/blog')}>Blog</NavDrawer.ListItem>
            <NavDrawer.ListItem to={route('/contact')}>
              Contact
            </NavDrawer.ListItem>
          </NavDrawer.List>
          <NavDrawer.List>
            {!isUser && (
              <div className='flex flex-col space-y-2'>
                <Router
                  className='flex h-11 w-full items-center justify-center'
                  to={route('/login')}
                >
                  <span className='font-display font-medium text-medium'>
                    Log in
                  </span>
                </Router>
                {isSignupGated && (
                  <ButtonRouter className='h-11' to={route('/waitlist')}>
                    Join the waitlist
                  </ButtonRouter>
                )}
                {!isSignupGated && (
                  <ButtonRouter className='h-11' to={route('/signup')}>
                    Sign up
                  </ButtonRouter>
                )}
              </div>
            )}
            {isUser && (
              <div className='flex flex-col space-y-2'>
                <ButtonRouter className='h-11' to={route('/')}>
                  Go to app
                </ButtonRouter>
              </div>
            )}
          </NavDrawer.List>
        </NavDrawer>
      </NavDrawer.Modal>
    </div>
  )
}
