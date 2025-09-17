import type { SerializeFrom } from '@remix-run/node'
import type { UIMatch } from '@remix-run/react'
import {
  NavLink,
  Outlet,
  useMatches,
  useNavigate,
  useRouteLoaderData,
  useSearchParams
} from '@remix-run/react'
import clsx from 'clsx'
import type { MotionProps } from 'framer-motion'
import { AnimatePresence, motion } from 'framer-motion'
import type { FC, ReactNode } from 'react'
import { forwardRef, useEffect, useState } from 'react'
import { StructuredText } from 'react-datocms'
import { route } from 'routes-gen'
import {
  AnchorRouter,
  ButtonRouter,
  Icon,
  IconButton,
  InterledgerWalletLogo,
  Router,
  SnackbarStage
} from '~/components'
import { ContentRouter, Prose } from '~/components/Content'
import { CommandActions } from '~/components/Scaffold/CommandActions'
import { CommandPalette } from '~/components/Scaffold/CommandPalette'
import type { FooterRecord } from '~/generated/dato-cms-graphql'
import {
  ConnectDomainStep,
  useConnectDomainStore
} from '~/lib/useConnectDomainStore'
import { useFormStore } from '~/lib/useFormStore'
import { PayStep, usePayStore } from '~/lib/usePayStore'
import { useScaffoldStore } from '~/lib/useScaffoldStore'
import { SignupStep, useSignupStore } from '~/lib/useSignupStore'
import type { loader as rootLoader } from '~/root'
import { NavDrawer } from './NavDrawer'

export type ApplicationProps = {
  layout: Layouts | ((match: UIMatch<any, ApplicationProps>) => Layouts)
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
    back?: string | ((match: UIMatch<any, ApplicationProps>) => string)
    title?: string | ((match: UIMatch<any, ApplicationProps>) => string)
    actions?: (match: UIMatch<any, ApplicationProps>) => {
      key: string
      nodes: ReactNode | ReactNode[]
    } | null
  }
  footer?: (
    match: UIMatch<any, ApplicationProps>
  ) => SerializeFrom<FooterRecord> | null | undefined
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
  const navigate = useNavigate()
  const [search] = useSearchParams()
  const { isUser, snackbar, features } = useRouteLoaderData(
    'root'
  ) as SerializeFrom<typeof rootLoader>

  const [payStep, payStepBack] = usePayStore((state) => [
    state.step,
    state.stepBack
  ])

  const [signupStep, signupStepBack] = useSignupStore((state) => [
    state.step,
    state.stepBack
  ])
  const [formStep, formStepBack] = useFormStore((state) => [
    state.step,
    state.stepBack
  ])
  const [connectDomainStep, connectDomainStepBack] = useConnectDomainStore(
    (state) => [state.step, state.stepBack]
  )

  // TODO should use second last match for scaffold if current match is nested (Only on desktop)
  let currentMatch = matches[matches.length - 1] as UIMatch<
    any,
    ApplicationProps
  >

  const scaffold: ScaffoldProps | undefined = currentMatch.handle.scaffold

  const [pushSnackbar, setCommandPaletteOpen] = useScaffoldStore((state) => [
    state.pushSnackbar,
    state.setCommandPalletOpen
  ])

  useEffect(() => {
    if (snackbar) pushSnackbar(snackbar)
  }, [pushSnackbar, snackbar])

  const footer = scaffold?.footer && scaffold.footer(currentMatch)

  const layoutHandle = currentMatch?.handle?.layout

  let layout: Layouts
  if (typeof layoutHandle === 'function') layout = layoutHandle(currentMatch)
  else layout = layoutHandle

  const actionHandle = scaffold?.header?.actions
  let actions = null
  if (typeof actionHandle !== 'undefined') {
    actions = actionHandle(currentMatch) ?? null
  }

  const titleHandle = scaffold?.header?.title

  let title: string
  if (typeof titleHandle === 'function') title = titleHandle(currentMatch)
  else title = titleHandle ?? ''

  const parentMatch = matches[matches.length - 2] as UIMatch<
    any,
    ApplicationProps
  >

  const parentTitleHandle = parentMatch?.handle?.scaffold?.header?.title

  let parentTitle = ''
  if (scaffold?.isNested) {
    if (typeof parentTitleHandle === 'function')
      parentTitle = parentTitleHandle(parentMatch)
    else parentTitle = parentTitleHandle ?? ''

    // if (!actions && !isNested) {
    //   if (typeof actionHandle === 'function')
    //     actions = actionHandle(matches[matches.length - 2]) ?? null
    //   else actions = actionHandle ?? null
    // }
  }

  return (
    <div
      className={clsx(
        'relative inset-0 flex min-h-screen flex-col',
        (layout === Layouts.Marketing || layout === Layouts.Docs) &&
          'bg-mk-page'
      )}
    >
      {layout === Layouts.Wallet && (
        <NavDrawerRoot>
          <NavDrawer.List>
            <div className='ml-4'>
              <Router to={route('/')} aria-label='Interledger Wallet logo'>
                <InterledgerWalletLogo className='h-8' />
              </Router>
            </div>
            {features && features.accountEnabled && (
              <button
                onClick={() => setCommandPaletteOpen(true)}
                className='mb-2 mt-10 flex w-full space-x-3 rounded-2xl bg-primary p-4 text-white'
              >
                <Icon>attach_money</Icon>
                <span className='font-medium'>Pay</span>
              </button>
            )}
            <NavDrawer.ListItem to={route('/')}>Home</NavDrawer.ListItem>
            <NavDrawer.ListItem to={route('/accounts')}>
              Accounts
            </NavDrawer.ListItem>
            <NavDrawer.ListItem to={route('/payments')}>
              Payments
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
            'mx-auto h-16 select-none bg-page sm:mt-[5.5rem] sm:max-w-[29rem]',
          layout === Layouts.Wallet &&
            'min-h-16 flex-col justify-end bg-page lg:mt-[5.5rem] lg:pl-[16.25rem]',
          layout === Layouts.Docs &&
            'h-16 bg-mk-page lg:mt-[5.5rem] lg:pl-[16.25rem]'
        )}
      >
        {layout === Layouts.Marketing && (
          <div className='mx-auto flex w-full max-w-[59rem] items-center'>
            <div className='lg:hidden'>
              <IconButton
                onClick={() => setOpenNavModal(true)}
                aria-label='Open menu'
              >
                menu
              </IconButton>
            </div>
            <div className='ml-4 lg:ml-0'>
              <Router to={route('/')} aria-label='Interledger Wallet logo'>
                <InterledgerWalletLogo className='h-8' />
              </Router>
            </div>
            <div className='hidden space-x-10 pb-2 pl-10 pt-3 lg:flex'>
              {/*<HeaderLink to='/about' title='About' />*/}
              {/*<HeaderLink to='/wallet' title='Wallet' />*/}
              {/*<HeaderPopover />*/}
              {/*<HeaderLink to={route('/docs')} title='Docs' />*/}
              {/*<HeaderLink to={route('/blog')} title='Blog' />*/}
              <HeaderLink to={route('/contact')} title='Contact' />
            </div>
            <div className='ml-auto hidden items-center lg:flex'>
              {!isUser && (
                <div className='flex space-x-10 pb-2 pt-3'>
                  <Router to={route('/login')}>
                    <span className='text-sm font-medium'>Log in</span>
                  </Router>
                  <Router to={route('/signup')}>
                    <span className='text-sm font-medium'>Sign up</span>
                  </Router>
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
          <div
            className={clsx(
              'mx-auto flex w-full items-center sm:max-w-lg lg:max-w-3xl xl:max-w-[59rem]',
              layout != Layouts.Focus && 'lg:px-4'
            )}
          >
            {!scaffold?.header.back && (
              <div className='lg:hidden'>
                <IconButton
                  className={clsx(layout === Layouts.Focus && 'hidden', 'mr-4')}
                  onClick={() => setOpenNavModal(true)}
                  aria-label='Open menu'
                >
                  menu
                </IconButton>
              </div>
            )}
            {scaffold?.header.back && (
              <IconButton
                className={clsx('mr-4', scaffold.isNested && 'lg:hidden')}
                onClick={() => {
                  if (search.has('back')) {
                    const searchBack = search.get('back')
                    if (searchBack) navigate(-parseInt(searchBack))
                  }
                  if (scaffold.header.back === 'pay') {
                    if (payStep == PayStep.AMOUNT) {
                      payStepBack()
                      navigate(-1)
                    } else payStepBack()
                  } else if (scaffold.header.back === 'signup') {
                    if (signupStep == SignupStep.LANDING) {
                      signupStepBack()
                      navigate(-1)
                    } else signupStepBack()
                  } else if (scaffold.header.back === 'form') {
                    if (formStep == 0) {
                      formStepBack()
                      navigate(-1)
                    } else formStepBack()
                  } else if (scaffold.header.back === 'connect-domain') {
                    if (connectDomainStep == ConnectDomainStep.LANDING) {
                      connectDomainStepBack()
                      navigate(-1)
                    } else connectDomainStepBack()
                  } else {
                    if (typeof scaffold.header.back === 'string') {
                      navigate(scaffold.header.back)
                    } else {
                      navigate(-1)
                    }
                  }
                }}
                aria-label='Back'
              >
                arrow_back
              </IconButton>
            )}
            {title && (
              <h1 className='text-xl font-medium lg:hidden'>{title}</h1>
            )}
            {(title || parentTitle) && (
              <h1 className='hidden text-xl font-medium lg:block'>
                {parentTitle || title}
              </h1>
            )}
            <Router
              className={clsx(
                !title && layout !== Layouts.Focus && 'lg:hidden',
                title && 'hidden'
              )}
              to={route('/')}
              aria-label='Interledger Wallet logo'
            >
              <InterledgerWalletLogo className='h-8' />
            </Router>
            <div className='ml-auto flex items-center space-x-4'>
              <AnimatePresence mode='wait'>
                {actions == null && ''}
                {actions != null && (
                  <motion.div
                    key={'header-action' + actions.key}
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
                      damping: 20
                    }}
                    className='flex items-center space-x-2'
                  >
                    {Array.isArray(actions.nodes) &&
                      actions.nodes.map((action, index) => {
                        return (
                          <div
                            key={'header-action' + index}
                            className='ml-auto'
                          >
                            {action}
                          </div>
                        )
                      })}
                    {!Array.isArray(actions.nodes) && actions.nodes}
                  </motion.div>
                )}
              </AnimatePresence>
            </div>
          </div>
        )}
      </header>
      <main
        className={clsx(
          'relative flex w-full grow flex-col',
          layout === Layouts.Marketing && 'mx-auto xl:max-w-[80rem]',
          layout === Layouts.Focus &&
            'mx-auto w-full gap-y-4 px-4 sm:max-w-[29rem] sm:px-0',
          layout === Layouts.Wallet && 'mb-32 w-full px-4 lg:pl-[16.25rem]',
          layout === Layouts.Docs && 'mb-32 w-full px-4 lg:pl-[16.25rem]'
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
          (layout === Layouts.Wallet || layout === Layouts.Docs) &&
            'fixed bottom-0 z-50 hidden w-56 items-center gap-x-3 px-4 py-6 lg:flex'
        )}
      >
        {layout !== Layouts.Marketing && (
          <>
            <Router
              className='text-xs font-medium text-primary'
              to='https://interledger.org'
            >
              The Interledger Foundation
            </Router>
            <span className='text-xs font-medium text-medium'>|</span>
            <Router className='text-xs font-medium text-primary' to='/legal'>
              Privacy &amp; Terms
            </Router>
          </>
        )}
        {layout === Layouts.Marketing && footer && (
          <div className='relative mx-auto flex w-full flex-col px-4 pb-12 pt-52 lg:px-0 lg:pl-40 lg:pt-20 xl:max-w-[59rem]'>
            <img
              alt='Interledger logo'
              className='absolute left-4 top-10 max-h-20 lg:left-3 lg:top-20'
              loading='lazy'
              src={footer.logo?.url}
            />
            <div className='flex w-full flex-col gap-y-10 lg:flex-row lg:gap-x-40 lg:gap-y-0'>
              <div className='flex flex-col'>
                <h3 className='font-medium text-on-color'>
                  {footer.column1Title}
                </h3>
                {footer.column1.map((link, index) => (
                  <ContentRouter
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
                  <ContentRouter
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
                  <ContentRouter
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
              <Prose className='prose-sm prose-p:text-on-color'>
                {footer.legalText && (
                  <StructuredText data={footer.legalText.value} />
                )}
              </Prose>
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
                  <Router to={route('/')} aria-label='Interledger Wallet logo'>
                    <InterledgerWalletLogo className='h-8' />
                  </Router>
                </div>
                <NavDrawer.ListItem to={route('/')}>Home</NavDrawer.ListItem>
                <NavDrawer.ListItem to={route('/accounts')}>
                  Accounts
                </NavDrawer.ListItem>
                <NavDrawer.ListItem to={route('/payments')}>
                  Payments
                </NavDrawer.ListItem>
                <NavDrawer.ListItem to={route('/settings')}>
                  Settings
                </NavDrawer.ListItem>
                <NavDrawer.ListItem to={route('/support')}>
                  Support
                </NavDrawer.ListItem>
              </NavDrawer.List>
              <footer className='flex w-full space-x-3 pb-2 pl-4'>
                <Router
                  className='text-xs font-medium text-primary'
                  to='/legal'
                >
                  Privacy &amp; Terms
                </Router>
              </footer>
            </>
          )}

          {layout === Layouts.Docs && (
            <>
              <NavDrawer.List>
                {!isUser && (
                  <div className='flex flex-col space-y-2'>
                    <Router
                      className='flex h-11 w-full items-center justify-center'
                      to={route('/login')}
                    >
                      <span className='font-medium text-medium'>Log in</span>
                    </Router>
                    <ButtonRouter className='h-11' to={route('/signup')}>
                      Sign up
                    </ButtonRouter>
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
                    <Router to={route('/')} aria-label='Interledger App logo'>
                      <InterledgerWalletLogo className='h-8' />
                    </Router>
                  </div>
                </div>
                <NavDrawer.ListItem to={route('/')}>Home</NavDrawer.ListItem>
                {/*<NavDrawer.ListItem to='/about'>About</NavDrawer.ListItem>*/}
                {/* <NavDrawer.ListItem to='/wallet'>Wallet</NavDrawer.ListItem> */}
                {/* <NavDrawer.ListItem to='/wealth'>Wealth</NavDrawer.ListItem> */}
                {/*<NavDrawer.ListItem to={route('/docs')}>*/}
                {/*  Docs*/}
                {/*</NavDrawer.ListItem>*/}
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
                      <span className='font-medium text-medium'>Log in</span>
                    </Router>
                    <ButtonRouter className='h-11' to={route('/signup')}>
                      Sign up
                    </ButtonRouter>
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
      <div
        className={clsx(
          'fixed bottom-4 left-0 z-50 mx-auto flex w-full flex-col items-end justify-center gap-y-4 overflow-y-visible px-4 text-center lg:bottom-auto lg:top-4 lg:items-center',
          layout === Layouts.Wallet && 'lg:pl-64 lg:pr-0'
        )}
      >
        <AnimatePresence mode='popLayout'>
          {features &&
            features.accountEnabled &&
            scaffold?.fab &&
            layout !== Layouts.Marketing && (
              <FAB key='fab' onTap={() => setCommandPaletteOpen(true)} />
            )}
          <SnackbarStage />
        </AnimatePresence>
        {features && features.accountEnabled && layout === Layouts.Wallet && (
          <CommandPalette>
            <CommandActions />
          </CommandPalette>
        )}
      </div>
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

// const HeaderPopover: FC = () => {
//   return (
//     <Popover className='relative'>
//       {({ open }) => (
//         <>
//           <Popover.Button className='inline-flex items-center gap-x-2 rounded text-sm font-medium focus-visible:outline focus-visible:outline-2 focus-visible:outline-focus'>
//             <span>Products</span>
//             <Icon>expand_more</Icon>
//           </Popover.Button>
//           <Transition
//             as={Fragment}
//             enter='transition ease-out duration-200'
//             enterFrom='opacity-0 translate-y-1'
//             enterTo='opacity-100 translate-y-0'
//             leave='transition ease-in duration-150'
//             leaveFrom='opacity-100 translate-y-0'
//             leaveTo='opacity-0 translate-y-1'
//           >
//             <Popover.Panel
//               static
//               className='absolute left-1/2 z-10 mt-3 w-screen max-w-sm -translate-x-1/2 transform px-4 sm:px-0'
//             >
//               <div className='relative flex flex-col gap-y-1 rounded-[1.25rem] bg-mk-section p-2 shadow-lg'>
//                 <Popover.Button
//                   as={CardLink}
//                   to='/wallet'
//                   className='flex gap-2 rounded-xl p-3 hover:bg-nav focus-visible:outline-2 focus-visible:outline-focus'
//                 >
//                   <svg
//                     width='24'
//                     height='24'
//                     viewBox='0 0 24 24'
//                     fill='none'
//                     xmlns='http://www.w3.org/2000/svg'
//                   >
//                     <path
//                       d='M12 18C12 14.6863 14.6863 12 18 12V12C21.3137 12 24 14.6863 24 18V18C24 21.3137 21.3137 24 18 24V24C14.6863 24 12 21.3137 12 18V18Z'
//                       fill='#FED7AA'
//                     />
//                     <path
//                       d='M12 0H24V6C24 9.31371 21.3137 12 18 12V12C14.6863 12 12 9.31371 12 6V0Z'
//                       fill='#FB923C'
//                     />
//                     <path
//                       fillRule='evenodd'
//                       clipRule='evenodd'
//                       d='M12 0C5.37258 0 0 5.37258 0 12V18C0 14.6863 2.68629 12 6 12C9.31371 12 12 14.6863 12 18V0Z'
//                       fill='#F97316'
//                     />
//                   </svg>

//                   <div className='flex flex-col gap-2'>
//                     <p className='text-sm font-medium text-strong'>Wallet</p>
//                     <p className='text-sm text-weak'>
//                       Send money as easily as email.
//                     </p>
//                   </div>
//                 </Popover.Button>
//                 <Popover.Button
//                   as={CardLink}
//                   to='https://wealth.fynbos.app'
//                   className='flex gap-2 rounded-xl p-3 hover:bg-nav focus-visible:outline-2 focus-visible:outline-focus'
//                 >
//                   <svg
//                     width='24'
//                     height='24'
//                     viewBox='0 0 24 24'
//                     fill='none'
//                     xmlns='http://www.w3.org/2000/svg'
//                   >
//                     <path
//                       d='M12 12C12 5.37258 17.3726 0 24 0V12C24 18.6274 18.6274 24 12 24V12Z'
//                       fill='#84CC16'
//                     />
//                     <path
//                       d='M0 6C0 2.68629 2.68629 0 6 0C9.31371 0 12 2.68629 12 6V12H0V6Z'
//                       fill='#A3E635'
//                     />
//                     <path
//                       d='M0 12C0 8.68629 2.68629 6 6 6C9.31371 6 12 8.68629 12 12C12 15.3137 9.31371 18 6 18C2.68629 18 0 15.3137 0 12Z'
//                       fill='#F1F5F9'
//                     />
//                     <path
//                       d='M0 12C6.62742 12 12 17.3726 12 24C5.37258 24 0 18.6274 0 12Z'
//                       fill='#A3E635'
//                     />
//                   </svg>

//                   <div className='flex flex-col gap-2'>
//                     <p className='text-sm font-medium text-strong'>Wealth</p>
//                     <p className='text-sm text-weak'>
//                       Grow your wealth easily.
//                     </p>
//                   </div>
//                 </Popover.Button>
//               </div>
//             </Popover.Panel>
//           </Transition>
//         </>
//       )}
//     </Popover>
//   )
// }

interface ButtonProps extends MotionProps {
  shrink?: boolean // sm:max-w-fit
}

export const FAB = forwardRef<never, ButtonProps>(
  ({ children, shrink, ...buttonProps }, ref) => {
    return (
      <motion.button
        key='Mobile-FAB'
        ref={ref}
        {...buttonProps}
        animate={{ opacity: 1, scale: 1, y: 0 }}
        initial={{ opacity: 0, scale: 0.5, y: 40 }}
        exit={{
          opacity: 0,
          scale: 0.5,
          y: 8,
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
        className='z-50 flex h-[6rem] w-[6rem] items-center justify-center self-end rounded-[1.75rem] bg-primary shadow-lg lg:hidden'
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
      </motion.button>
    )
  }
)

FAB.displayName = 'FAB'
