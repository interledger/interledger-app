import type { RouteMatch } from '@remix-run/react'
import { NavLink, Outlet, useMatches, useNavigate } from '@remix-run/react'
import clsx from 'clsx'
import { AnimatePresence, motion } from 'framer-motion'
import type { FC, ReactNode } from 'react'
import { useState } from 'react'
import { StructuredText } from 'react-datocms'
import { route } from 'routes-gen'
import {
  AnchorRouter,
  ButtonRouter,
  FynbosLogo,
  Icon,
  IconButton,
  MarketingRouter,
  Router,
  WalletGrid,
  WalletShapes
} from '~/components'
import type { FooterRecord } from '~/generated/dato-cms-graphql'
import { NavDrawer } from './NavDrawer'

export type ApplicationProps = {
  layout: Layouts | ((match: RouteMatch) => Layouts)
  scaffold?: ScaffoldProps
}

export enum Layouts {
  Focus = 'Focus',
  Docs = 'Docs',
  Wallet = 'Wallet',
  Marketing = 'Marketing'
}

export enum Fab {
  Pay = 'Pay',
  Identity = 'Identity',
  Account = 'Account'
}

export type ScaffoldHeaderActions = {
  type: 'search' | 'chip' | 'shapes'
  content?: (match: RouteMatch) => ReactNode
}

/**
 * ScaffoldProps
 * @property header - Scaffold header props
 * @property header.back - Scaffold Back button route
 * @property header.title - Scaffold header title
 * @property header.actions - Scaffold header actions
 * @property fab - Scaffold floating action button
 * @property footer - Scaffold footer for marketing pages
 * @property isNested - Is the current route nested in a parent route?
 */
export type ScaffoldProps = {
  header: {
    // Back should check the history stack, and if the previous route is the same as the specified route, it should pop the history stack
    back?: string | ((match: RouteMatch) => string)
    title?: string | ((match: RouteMatch) => string)
    actions?: ScaffoldHeaderActions[] // TODO: use a better type here, this is too generic
  }
  footer?: (match: RouteMatch) => FooterRecord
  fab?: Fab
  isNested?: boolean
}

const NavDrawerRoot = ({ children }: { children?: ReactNode }) => {
  return (
    <nav className='z-50 hidden min-w-max select-none px-3 py-4 lg:fixed lg:inset-y-0 lg:z-50 lg:flex lg:w-56 lg:flex-col'>
      {children}
    </nav>
  )
}

export function Scaffold() {
  const [openNavModal, setOpenNavModal] = useState<boolean>(false)
  const matches = useMatches()
  const isUser = matches[0]?.data.isUser
  const isSignupGated = matches[0]?.data.isSignupGated
  // TODO should use second last match for scaffold if current match is nested (Only on desktop)
  const scaffold: ScaffoldProps = matches[matches.length - 1].handle?.scaffold
  const footer = scaffold.footer && scaffold.footer(matches[matches.length - 1])
  const titleHandle = scaffold.header?.title
  const navigate = useNavigate()

  const layoutHandle = matches[matches.length - 1]?.handle?.layout

  let layout: Layouts
  if (typeof layoutHandle === 'function')
    layout = layoutHandle(matches[matches.length - 1])
  else layout = layoutHandle

  let title: string
  if (typeof titleHandle === 'function')
    title = titleHandle(matches[matches.length - 1])
  else title = titleHandle ?? ''

  return (
    <div
      className={clsx(
        'relative inset-0 flex min-h-screen flex-col',
        layout === Layouts.Marketing && 'bg-mk-page'
      )}
    >
      {layout === Layouts.Wallet && (
        <NavDrawerRoot>
          <NavDrawer.List>
            <div className='ml-4'>
              <Router to={route('/')} aria-label='Fynbos logo'>
                <FynbosLogo className='h-8' />
              </Router>
            </div>
            <Router
              to={route('/pay')}
              className='mb-2 mt-10 flex w-full space-x-3 rounded-2xl bg-primary p-4 text-white'
            >
              <Icon>attach_money</Icon>
              <span className='font-display font-medium'>Pay</span>
            </Router>
            <NavDrawer.ListItem to={route('/')}>Home</NavDrawer.ListItem>
            <NavDrawer.ListItem to={route('/transactions')}>
              Transactions
            </NavDrawer.ListItem>
            <NavDrawer.ListItem to={route('/settings')}>
              Settings
            </NavDrawer.ListItem>
            <NavDrawer.ListItem to={route('/support')}>
              Support
            </NavDrawer.ListItem>
          </NavDrawer.List>
        </NavDrawerRoot>
      )}
      <header
        className={clsx(
          'sticky top-0 z-40 flex w-full select-none items-center justify-start space-x-4 p-4',
          layout === Layouts.Marketing &&
            'h-16 border-b border-slate-200 bg-mk-page dark:border-slate-800 lg:h-24',
          layout === Layouts.Focus &&
            'mx-auto h-16 select-none bg-app sm:mt-[5.5rem] sm:max-w-[29rem]',
          layout === Layouts.Wallet &&
            'h-16 bg-app lg:mt-[5.5rem] lg:pl-[16.25rem]',
          layout === Layouts.Docs && 'h-[9.5rem] bg-app lg:pl-[15.75rem]'
        )}
      >
        {layout === Layouts.Marketing && (
          <div className='mx-auto flex w-full max-w-[59rem] items-center'>
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
              <HeaderLink to={route('/wallet')} title='Wallet' />
              <HeaderLink to={route('/about')} title='About' />
              {/*<HeaderLink to={route('/docs')} title='Docs' />*/}
              <HeaderLink to={route('/blog')} title='Blog' />
              <HeaderLink to={route('/contact')} title='Contact' />
            </div>
            {/*<div className='ml-auto flex items-center lg:hidden'>*/}
            {/*  <IconButton>light_mode</IconButton>*/}
            {/*</div>*/}
            <div className='ml-auto hidden items-center lg:flex'>
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
        )}
        {layout !== Layouts.Marketing && (
          <div className='mx-auto flex w-full items-center sm:max-w-lg lg:max-w-3xl xl:max-w-[59rem]'>
            <IconButton
              className={clsx(
                layout === Layouts.Focus && 'hidden',
                'mr-4 lg:hidden'
              )}
              onClick={() => setOpenNavModal(true)}
              aria-label='Open menu'
            >
              menu
            </IconButton>
            {/* TODO Make this smarter. */}
            {scaffold.header.back && (
              <IconButton
                className='mr-4'
                onClick={() => {
                  navigate(-1)
                  console.log('Go back sir')
                }}
                aria-label='Back'
              >
                arrow_back
              </IconButton>
            )}
            {title && <h1 className='text-xl font-medium'>{title}</h1>}
            <Router
              className={clsx(!title && 'lg:hidden', title && 'hidden')}
              to={route('/')}
              aria-label='Fynbos logo'
            >
              <FynbosLogo className='h-8' />
            </Router>
            {scaffold.header?.actions && (
              <div className='ml-auto flex items-center space-x-4'>
                {scaffold.header.actions.map((action, index) => {
                  return (
                    <div key={'header-action' + index} className='ml-auto'>
                      {action.type === 'chip' &&
                        action?.content &&
                        action?.content(matches[matches.length - 1])}
                      {action.type === 'search' && (
                        <IconButton>light_mode</IconButton>
                      )}
                      {action.type === 'shapes' && <WalletShapes />}
                    </div>
                  )
                })}
              </div>
            )}
          </div>
        )}
      </header>
      <main
        className={clsx(
          'relative flex w-full grow flex-col',
          layout === Layouts.Marketing && 'mx-auto xl:max-w-[80rem]',
          layout === Layouts.Focus &&
            'mx-auto w-full gap-y-6 px-4 sm:max-w-[29rem] sm:px-0',
          layout === Layouts.Wallet && 'mb-32 w-full px-4 lg:pl-[16.25rem]',
          layout === Layouts.Docs && 'w-full px-4 lg:pl-[16.25rem]'
        )}
      >
        <Outlet />
      </main>
      <footer
        className={clsx(
          'w-full',
          layout === Layouts.Marketing &&
            'mx-auto mb-8 flex max-w-[80rem] rounded-2xl bg-mk-footer',
          layout === Layouts.Focus &&
            'mx-auto flex w-full items-center gap-x-3 px-4 py-6 sm:max-w-[29rem] sm:px-0',
          layout === Layouts.Wallet &&
            'fixed bottom-0 z-50 hidden w-56 items-center gap-x-3 px-4 lg:flex',
          layout === Layouts.Docs && 'z-50 flex w-56 items-center gap-x-3 px-4'
        )}
      >
        {layout !== Layouts.Marketing && (
          <>
            <span className='text-xs font-medium text-medium'>
              &copy;&nbsp;Fynbos
            </span>
            <Router
              className='text-xs font-medium text-primary'
              to={route('/legal')}
            >
              Privacy &amp; Terms
            </Router>
          </>
        )}
        {layout === Layouts.Marketing && footer && (
          <div className='relative mx-auto flex w-full flex-col px-4 pb-12 pt-52 lg:px-0 lg:pl-40 lg:pt-20 xl:max-w-[59rem]'>
            <img
              alt='Fynbos logo'
              className='absolute left-4 top-10 lg:left-4 lg:top-20'
              loading='lazy'
              src={footer.logo?.url}
            />
            <div className='flex w-full flex-col gap-y-10 lg:flex-row lg:gap-x-40 lg:gap-y-0'>
              <div className='flex flex-col'>
                <h3 className='font-medium text-on-color'>
                  {footer.column1Title}
                </h3>
                {footer.column1.map((link, index) => (
                  <MarketingRouter
                    key={link.id + 'FooterLink'}
                    to={link}
                    className='mt-1 text-disabled first-of-type:mt-4'
                  />
                ))}
              </div>
              <div className='flex flex-col'>
                <h3 className='font-medium text-on-color'>
                  {footer.column2Title}
                </h3>
                {footer.column2.map((link, index) => (
                  <MarketingRouter
                    key={link.id + 'FooterLink'}
                    to={link}
                    className='mt-1 text-disabled first-of-type:mt-4'
                  />
                ))}
              </div>
              <div className='flex flex-col'>
                <h3 className='font-medium text-on-color'>
                  {footer.column3Title}
                </h3>
                {footer.column3.map((link, index) => (
                  <MarketingRouter
                    key={link.id + 'FooterLink'}
                    to={link}
                    className='mt-1 text-disabled first-of-type:mt-4'
                  />
                ))}
              </div>
            </div>
            <div className='mt-10 flex items-center space-x-4'>
              {footer.socialIcons.map((icon, index) => {
                return (
                  <AnchorRouter to={icon.url ?? ''} key={'social-icon' + index}>
                    <img
                      alt='Social logo'
                      className='block'
                      loading='lazy'
                      src={icon.icon?.url}
                    />
                  </AnchorRouter>
                )
              })}
            </div>
            <div className='mt-6 flex w-full items-center justify-between'>
              <div className='prose prose-sm prose-invert prose-a:rounded prose-a:text-primary prose-a:no-underline prose-a:focus-visible:outline prose-a:focus-visible:outline-2 prose-a:focus-visible:outline-focus'>
                {footer.legalText && (
                  <StructuredText data={footer.legalText.value} />
                )}
              </div>
            </div>
          </div>
        )}
      </footer>

      <NavDrawer.Modal open={openNavModal} setOpen={setOpenNavModal}>
        <NavDrawer>
          {layout === Layouts.Wallet && (
            <>
              <NavDrawer.List>
                <div className='relative mb-8 ml-1 flex items-center space-x-4'>
                  <IconButton
                    onClick={() => setOpenNavModal(!openNavModal)}
                    aria-label='Close menu'
                  >
                    menu_open
                  </IconButton>
                  <Router to={route('/')} aria-label='Fynbos logo'>
                    <FynbosLogo className='h-8' />
                  </Router>
                </div>
                <NavDrawer.ListItem to={route('/')}>Home</NavDrawer.ListItem>
                <NavDrawer.ListItem to={route('/transactions')}>
                  Transactions
                </NavDrawer.ListItem>
                <NavDrawer.ListItem to={route('/settings')}>
                  Settings
                </NavDrawer.ListItem>
                <NavDrawer.ListItem to={route('/support')}>
                  Support
                </NavDrawer.ListItem>
              </NavDrawer.List>
              <footer className='flex w-full space-x-3 pb-2 pl-4'>
                <span className='text-xs font-medium text-medium'>
                  &copy;&nbsp;Fynbos
                </span>
                <Router
                  className='text-xs font-medium text-primary'
                  to={route('/legal')}
                >
                  Privacy &amp; Terms
                </Router>
              </footer>
            </>
          )}
          {layout === Layouts.Marketing && (
            <>
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
                <NavDrawer.ListItem to={route('/wallet')}>
                  Wallet
                </NavDrawer.ListItem>
                <NavDrawer.ListItem to={route('/about')}>
                  About
                </NavDrawer.ListItem>
                {/*<NavDrawer.ListItem to={route('/docs')}>*/}
                {/*  Docs*/}
                {/*</NavDrawer.ListItem>*/}
                <NavDrawer.ListItem to={route('/blog')}>
                  Blog
                </NavDrawer.ListItem>
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
            </>
          )}
        </NavDrawer>
      </NavDrawer.Modal>
      <AnimatePresence>
        {scaffold.fab && layout !== Layouts.Marketing && (
          <FAB to={route('/pay')} />
        )}
      </AnimatePresence>
    </div>
  )
}

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
export function Scaffold2() {
  // TODO: try global snackbar
  const [openNavModal, setOpenNavModal] = useState<boolean>(false)
  const matches = useMatches()
  const title = matches[matches.length - 1].handle?.title
  // TODO adjust the content stage based on layout
  // TODO should use second last match for scaffold if current match is nested (Only on desktop)
  const scaffold = matches[matches.length - 1].handle?.scaffold

  const layout = matches[matches.length - 1]?.handle?.layout

  return (
    <div className='relative inset-0 flex min-h-screen flex-col lg:flex-row'>
      <div className='fixed top-0 hidden lg:flex'>
        {scaffold.hasNavDrawer && (
          <NavDrawer>
            <NavDrawer.List>
              <div className='ml-4'>
                <Router to={route('/')} aria-label='Fynbos logo'>
                  <FynbosLogo className='h-8' />
                </Router>
              </div>
              <Router
                to={route('/pay')}
                className='mb-2 mt-10 flex w-full space-x-3 rounded-2xl bg-primary p-4 text-white'
              >
                <Icon>attach_money</Icon>
                <span className='font-display font-medium'>Pay</span>
              </Router>
              <NavDrawer.ListItem to={route('/')}>Home</NavDrawer.ListItem>
              <NavDrawer.ListItem to={route('/transactions')}>
                Transactions
              </NavDrawer.ListItem>
              <NavDrawer.ListItem to={route('/settings')}>
                Settings
              </NavDrawer.ListItem>
              <NavDrawer.ListItem to={route('/support')}>
                Support
              </NavDrawer.ListItem>
            </NavDrawer.List>
            <footer className='flex w-full space-x-3 pb-2 pl-4'>
              <span className='text-xs font-medium text-medium'>
                &copy;&nbsp;Fynbos
              </span>
              <Router
                className='text-xs font-medium text-primary'
                to={route('/legal')}
              >
                Privacy &amp; Terms
              </Router>
            </footer>
          </NavDrawer>
        )}
      </div>
      <div className='w-full'>
        <header className='fixed top-0 z-50 flex h-16 w-full select-none items-center justify-start space-x-4 bg-page p-4 sm:min-w-full lg:hidden'>
          <IconButton
            className='lg:hidden'
            onClick={() => setOpenNavModal(true)}
            aria-label='Open menu'
          >
            menu
          </IconButton>
          {title && <h1 className='text-xl font-medium'>{title}</h1>}
          {!title && <FynbosLogo className='h-8' />}
        </header>
        {/* TODO adjust the conntent stage based on layout:*/}
        {layout === Layouts.Marketing && (
          <div className='relative mx-auto w-full sm:max-w-lg lg:max-w-3xl xl:max-w-[59rem]'>
            <Outlet />
          </div>
        )}
        {layout === Layouts.Wallet && (
          <div className='mb-32 mt-16 overflow-hidden lg:my-[5.5rem] lg:ml-64'>
            <div className='relative mx-auto w-full sm:max-w-lg lg:max-w-3xl xl:max-w-[59rem]'>
              {title && (
                <WalletGrid>
                  <div className='col-span-full hidden justify-between px-4 pb-6 sm:col-span-6 sm:col-start-2 lg:col-start-4 lg:flex'>
                    <h1 className='text-2xl font-medium'>{title}</h1>
                    <WalletShapes />
                  </div>
                </WalletGrid>
              )}

              <Outlet />
            </div>
          </div>
        )}
        {layout === Layouts.Focus && (
          <div className='col-span-full mt-16 grid grid-cols-1 gap-y-6 px-4 sm:px-0 lg:col-span-6 lg:col-start-4 lg:mt-36'>
            <Outlet />
          </div>
        )}
        {layout === Layouts.Docs && (
          <div className='mb-32 mt-16 overflow-hidden lg:my-[5.5rem] lg:ml-64'>
            <div className='relative mx-auto w-full sm:max-w-lg lg:max-w-3xl xl:max-w-[59rem]'>
              {title && (
                <WalletGrid>
                  <div className='col-span-full hidden justify-between px-4 pb-6 sm:col-span-6 sm:col-start-2 lg:col-start-4 lg:flex'>
                    <h1 className='text-2xl font-medium'>{title}</h1>
                    <WalletShapes />
                  </div>
                </WalletGrid>
              )}

              <Outlet />
            </div>
          </div>
        )}
      </div>
      {scaffold.hasNavDrawer && (
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
                <Router to={route('/')} aria-label='Fynbos logo'>
                  <FynbosLogo className='h-8' />
                </Router>
              </div>
              <NavDrawer.ListItem to={route('/')}>Home</NavDrawer.ListItem>
              <NavDrawer.ListItem to={route('/transactions')}>
                Transactions
              </NavDrawer.ListItem>
              <NavDrawer.ListItem to={route('/settings')}>
                Settings
              </NavDrawer.ListItem>
              <NavDrawer.ListItem to={route('/support')}>
                Support
              </NavDrawer.ListItem>
            </NavDrawer.List>
            <footer className='flex w-full space-x-3 pb-2 pl-4'>
              <span className='text-xs font-medium text-medium'>
                &copy;&nbsp;Fynbos
              </span>
              <Router
                className='text-xs font-medium text-primary'
                to={route('/legal')}
              >
                Privacy &amp; Terms
              </Router>
            </footer>
          </NavDrawer>
        </NavDrawer.Modal>
      )}
      <AnimatePresence>
        {scaffold.hasFab && <FAB to={route('/pay')} />}
      </AnimatePresence>
    </div>
  )
}

function FAB({ to }: { to: string }) {
  return (
    <Router to={to}>
      <motion.div
        key={to}
        animate={{ opacity: 1, scale: 1 }}
        initial={{ opacity: 0, scale: 0.5 }}
        exit={{
          opacity: 0,
          scale: 0.5,
          transition: {
            duration: 0.2
          }
        }}
        transition={{
          type: 'spring',
          stiffness: 400,
          damping: 20,
          duration: 0.3
        }}
        className='fixed bottom-4 right-4 flex h-[6rem] w-[6rem] items-center justify-center rounded-[1.75rem] bg-primary shadow-lg lg:hidden'
      >
        <svg
          width='36'
          height='36'
          viewBox='0 0 36 36'
          fill='none'
          xmlns='http://www.w3.org/2000/svg'
        >
          <mask
            id='mask0_1852_4585'
            style={{ maskType: 'alpha' }}
            maskUnits='userSpaceOnUse'
            x='0'
            y='0'
            width='36'
            height='36'
          >
            <rect width='36' height='36' fill='#D9D9D9' />
          </mask>
          <g mask='url(#mask0_1852_4585)'>
            <path
              d='M16.5375 31.5V28.275C15.2125 27.975 14.069 27.4 13.107 26.55C12.144 25.7 11.4375 24.5 10.9875 22.95L13.7625 21.825C14.1375 23.025 14.694 23.9375 15.432 24.5625C16.169 25.1875 17.1375 25.5 18.3375 25.5C19.3625 25.5 20.2315 25.269 20.9445 24.807C21.6565 24.344 22.0125 23.625 22.0125 22.65C22.0125 21.775 21.7375 21.081 21.1875 20.568C20.6375 20.056 19.3625 19.475 17.3625 18.825C15.2125 18.15 13.7375 17.344 12.9375 16.407C12.1375 15.469 11.7375 14.325 11.7375 12.975C11.7375 11.35 12.2625 10.0875 13.3125 9.1875C14.3625 8.2875 15.4375 7.775 16.5375 7.65V4.5H19.5375V7.65C20.7875 7.85 21.819 8.306 22.632 9.018C23.444 9.731 24.0375 10.6 24.4125 11.625L21.6375 12.825C21.3375 12.025 20.9125 11.425 20.3625 11.025C19.8125 10.625 19.0625 10.425 18.1125 10.425C17.0125 10.425 16.175 10.669 15.6 11.157C15.025 11.644 14.7375 12.25 14.7375 12.975C14.7375 13.8 15.1125 14.45 15.8625 14.925C16.6125 15.4 17.9125 15.9 19.7625 16.425C21.4875 16.925 22.794 17.7185 23.682 18.8055C24.569 19.8935 25.0125 21.15 25.0125 22.575C25.0125 24.35 24.4875 25.7 23.4375 26.625C22.3875 27.55 21.0875 28.125 19.5375 28.35V31.5H16.5375Z'
              fill='white'
            />
          </g>
        </svg>
      </motion.div>
    </Router>
  )
}
