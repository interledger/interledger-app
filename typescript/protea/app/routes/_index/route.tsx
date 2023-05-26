import type { LoaderArgs } from '@remix-run/node'
import { defer } from '@remix-run/node'
import { useLoaderData } from '@remix-run/react'
import { Fragment, useState } from 'react'
import { route } from 'routes-gen'
import type { ApplicationProps } from '~/components'
import {
  AnimatedSchedule,
  ButtonRouter,
  Card,
  Fab,
  FynbosLogo,
  HomeShapes,
  Icon,
  IconButton,
  Layouts,
  Router,
  Shape,
  Snackbar,
  WalletGrid
} from '~/components'
import { getUserSession, hasUserSession } from '~/lib/kratos.server'
import { getPusherArgs } from '~/lib/pusher.server'
import { IS_SIGNUP_GATED } from '~/lib/signupCheck.server'
import type { SnackbarType } from '~/lib/snackbar.server'
import { getSnackbar } from '~/lib/snackbar.server'
import type { PusherArgs } from '~/lib/usePusher'
import { usePusher } from '~/lib/usePusher'
import type { Transaction } from '~/lib/wallet.server'
import {
  getKycStatus,
  getLinkedAccounts,
  getTransactionsWithPending,
  getWalletPaymentPointer
} from '~/lib/wallet.server'

export enum KycStatus {
  Unknown = 0,
  InProgress = 1,
  DocumentsRequired = 2,
  Verified = 3,
  Suspended = 4
}

export async function loader({ request }: LoaderArgs) {
  const isUser = hasUserSession(request)

  let data = {
    isUser: isUser,
    pusherArgs: {} as PusherArgs,
    isSignupGated: IS_SIGNUP_GATED,
    firstName: '',
    paymentPointer: {
      url: '',
      asset: 'USD',
      assetScale: 2,
      alias: 'default',
      walletID: '',
      formatted: ''
    },
    balance: '' as unknown as Promise<string>,
    transactions: [] as Transaction[],
    kycStatus: KycStatus.Unknown,
    canTopUp: false,
    canWithdraw: false,
    nextStep: {
      title: '',
      icon: '',
      action: { to: '', text: '' },
      show: false
    },
    snackbar: {
      message: ''
    } as SnackbarType
  }

  if (isUser) {
    const [
      session,
      paymentPointer,
      transactions,
      kycStatus,
      linkedAccounts,
      snackbar,
      pusherArgs
    ] = await Promise.all([
      getUserSession(request),
      getWalletPaymentPointer(request),
      getTransactionsWithPending(request, { pageSize: 3 }),
      getKycStatus(request),
      getLinkedAccounts(request),
      getSnackbar(request),
      getPusherArgs(request)
    ])

    data = {
      ...data,
      firstName: session.identity.traits.firstName,
      paymentPointer,
      transactions: transactions.transactions,
      kycStatus: kycStatus.kycStatus,
      canTopUp: linkedAccounts.canTopUp,
      canWithdraw: linkedAccounts.canWithdraw,
      snackbar,
      pusherArgs
    }

    /**
     * Next Step state machine
     * 1. Activate PP - KYCStatus.Unknown
     * 2. Add debit - KYCStatus.Verified + !hasTransactions + !canTopUp
     * 3. Add bank - KYCStatus.Verified + hasTransactions + !canWithdraw
     */
    if (data.kycStatus == KycStatus.Unknown) {
      data.nextStep = {
        title:
          'Your payment pointer is reserved, we just need a few more details to activate it.',
        icon: 'attach_money',
        action: {
          to: route('/personal-details'),
          text: 'Activate payment pointer'
        },
        show: true
      }
    } else if (
      data.kycStatus == KycStatus.Verified &&
      transactions.transactions.length == 0 &&
      !data.canTopUp
    ) {
      data.nextStep = {
        title:
          'Add a debit card to easily send payments or top up your cash balance.',
        icon: 'add_card',
        action: {
          to: route('/link-account/card'),
          text: 'Add a debit card'
        },
        show: true
      }
    } else if (
      data.kycStatus == KycStatus.Verified &&
      transactions.transactions.length > 0 &&
      !data.canWithdraw
    ) {
      data.nextStep = {
        title:
          'Add a bank account to securely withdraw from your cash balance at any time.',
        icon: 'account_balance',
        action: {
          to: route('/link-account/bank'),
          text: 'Add bank account'
        },
        show: true
      }
    }
  }
  return defer(data)
}

export const handle: ApplicationProps = {
  layout: (match) => (match.data.isUser ? Layouts.Wallet : Layouts.Marketing),
  scaffold: {
    header: { title: 'transparent' },
    fab: Fab.Pay
  }
}

export default function Page() {
  const { isUser } = useLoaderData<typeof loader>()

  if (isUser) return <AppPage />
  else return <MarketingPage />
}

function MarketingPage() {
  return (
    <main className='w-full'>
      <header className='fixed top-0 z-10 mb-16 flex h-16 w-full items-center border-b border-slate-200 bg-mk-page lg:h-24'>
        <div className='mx-auto flex w-full justify-between px-4 sm:max-w-lg sm:px-0 lg:max-w-3xl xl:max-w-[59rem]'>
          <div className='flex items-center'>
            <IconButton
              className='lg:hidden'
              // onClick={() => setOpenNavModal(true)}
              aria-label='Open menu'
            >
              menu
            </IconButton>
            <div className='ml-4 lg:ml-0'>
              <Router to={route('/')} aria-label='Fynbos logo'>
                <FynbosLogo className='h-8' />
              </Router>
            </div>
            {/*<div className='hidden space-x-10 pt-3 pb-2 pl-10 lg:flex'>*/}
            {/*  <HeaderLink*/}
            {/*    to={route('/what-is-a-payment-pointer')}*/}
            {/*    title='What is a payment pointer?'*/}
            {/*  />*/}
            {/*  /!*<HeaderLink to={route('/about')} title='About' />*!/*/}
            {/*  <HeaderLink to={route('/blog')} title='Blog' />*/}
            {/*  <HeaderLink to={route('/contact')} title='Contact' />*/}
            {/*</div>*/}
          </div>
          <div className='hidden items-center lg:flex'>
            {/*{!isUser && (*/}
            {/*  <div className='flex space-x-10 pt-3 pb-2'>*/}
            {/*    <Router to={route('/login')}>*/}
            {/*      <span className='text-sm font-medium'>Log in</span>*/}
            {/*    </Router>*/}
            {/*    {isSignupGated && (*/}
            {/*      <Router to={route('/waitlist')}>*/}
            {/*        <span className='text-sm font-medium'>*/}
            {/*          Join the waitlist*/}
            {/*        </span>*/}
            {/*      </Router>*/}
            {/*    )}*/}
            {/*    {!isSignupGated && (*/}
            {/*      <Router to={route('/signup')}>*/}
            {/*        <span className='text-sm font-medium'>Sign up</span>*/}
            {/*      </Router>*/}
            {/*    )}*/}
            {/*  </div>*/}
            {/*)}*/}
            {/*{isUser && (*/}
            <div className='flex items-center '>
              <ButtonRouter to={route('/')}>
                <span className='text-sm font-medium'>Go to app</span>
              </ButtonRouter>
            </div>
            {/*)}*/}
          </div>
        </div>
      </header>
      <section className='relative mx-auto grid w-full grid-cols-4 content-start gap-4 gap-y-2 overflow-x-visible px-4 sm:max-w-lg sm:grid-cols-8 sm:px-0 lg:max-w-3xl lg:grid-cols-12 xl:max-w-[59rem]'>
        <div className='relative col-span-full h-20 lg:h-48'>
          <div className='absolute -right-20 top-0 hidden h-20 w-20 rounded-br-full bg-rose-300 dark:bg-rose-400 lg:block' />
          <div className='absolute right-0 top-20 hidden h-20 w-20 rounded-br-full bg-slate-300 dark:bg-slate-600 lg:block' />
          <div className='absolute left-0 top-0 hidden h-20 w-20 rounded-bl-full bg-lime-300 dark:bg-lime-500 lg:block' />
          <div className='absolute -left-20 top-20 hidden h-20 w-20 rounded-br-full bg-slate-100 dark:bg-slate-800 lg:block' />
          {/* Mobile */}
          <div className='absolute -left-4 top-0 block h-20 w-20 rounded-br-full bg-slate-100 dark:bg-slate-800 lg:hidden' />
        </div>
        <div className='col-span-full'>
          <span className='flex justify-center text-center font-display text-3xl font-medium lg:text-6xl'>
            Building the better way to pay
          </span>
        </div>
        <div className='col-span-full lg:col-span-10 lg:col-start-2 lg:mt-7'>
          <span className='flex text-center lg:text-2xl'>
            A payment pointer from Fynbos is a secure, programmable wallet,
            which connects to all your accounts.
          </span>
        </div>
        <div className='col-span-full mt-10 flex flex-col items-center justify-center space-y-4 lg:mt-5 lg:flex-row lg:space-x-4 lg:space-y-0'>
          <ButtonRouter shrink to={route('/waitlist')}>
            Join the waitlist
          </ButtonRouter>
        </div>
        <div className='relative col-span-full h-48 lg:h-56'>
          <div className='absolute -left-[calc(100vw)] bottom-20 hidden h-20 w-screen bg-slate-700 dark:bg-slate-800 lg:block' />
          <div className='absolute -left-20 bottom-0 hidden h-20 w-20 rounded-tl-full bg-slate-700 dark:bg-slate-800 lg:block' />
          <div className='absolute bottom-0 left-0 hidden h-20 w-20 rounded-br-full bg-lime-500 dark:bg-lime-600 lg:block' />
          <div className='absolute bottom-20 right-20 hidden h-20 w-20 rounded-tl-full bg-lime-300 lg:block' />
          <div className='absolute bottom-20 right-0 hidden h-20 w-20 rounded-tr-full bg-lime-500 lg:block' />
          <div className='absolute bottom-0 right-20 hidden h-20 w-20 rounded-full bg-rose-500 lg:block' />
          <div className='absolute -right-[calc(100vw-5rem)] bottom-0 hidden h-20 w-screen rounded-bl-full bg-yellow-200 lg:block' />
          {/* Mobile */}
          <div className='absolute -right-4 bottom-0 block h-20 w-20 rounded-br-full bg-lime-300 lg:hidden' />
          <div className='absolute bottom-0 right-16 block h-20 w-20 rounded-bl-full bg-lime-500 lg:hidden' />
          <div className='absolute -right-4 bottom-20 block h-20 w-20 rounded-full bg-rose-500 lg:hidden' />
          <div className='absolute bottom-20 right-16 block h-20 w-20 rounded-tr-full bg-yellow-200 lg:hidden' />
        </div>
      </section>
      <div className='w-full bg-mk-section'>
        <section className='mx-auto grid w-full grid-cols-4 content-start gap-4 gap-y-2 overflow-x-visible px-4 sm:max-w-lg sm:grid-cols-8 sm:px-0 lg:max-w-3xl lg:grid-cols-12 xl:max-w-[59rem]'>
          <div className='order-1 col-span-full pt-20 lg:col-span-10 lg:col-start-2'>
            <span className='font-display text-2xl lg:text-4xl lg:leading-[2.75rem]'>
              Simplicity, security and programmability built on web-native
              technology.
            </span>
          </div>
          <div className='relative order-2 col-span-full mt-16 h-28 lg:col-span-4 lg:col-start-2 lg:mt-36'>
            <div className='absolute left-0 top-14 h-14 w-14 rounded-tl-full bg-orange-300' />
            <div className='absolute left-14 top-14 h-14 w-14 rounded-tl-full bg-orange-200' />
            <div className='absolute left-28 top-0 h-14 w-14 rounded-bl-full bg-orange-700' />
            <div className='absolute left-[10.5rem] top-14 h-14 w-14 rounded-full bg-orange-500' />
          </div>
          <div className='order-3 col-span-full mt-8 flex flex-col space-y-4 lg:col-span-6 lg:col-start-6 lg:mt-36 lg:space-y-7'>
            <span className='font-display text-xl font-medium'>Simple</span>
            <span>
              Everything you need and nothing you don’t. A clean and simple user
              interface for managing your connected accounts. Make deposits and
              withdrawals, send and receive payments using payment pointers, and
              configure connected applications.
            </span>
          </div>
          <div className='order-5 col-span-full mt-8 flex flex-col space-y-4 lg:order-4 lg:col-span-6 lg:col-start-2 lg:mt-36 lg:space-y-7'>
            <span className='font-display text-xl font-medium'>Secure</span>
            <span>
              Never share your card number or bank account details again.
              Securely load them into your wallet and share your payment pointer
              instead.
            </span>
          </div>
          <div className='relative order-4 col-span-full mt-16 h-28 lg:order-5 lg:col-span-4 lg:col-start-9 lg:mt-36'>
            <div className='absolute left-0 top-14 h-14 w-14 rounded-full bg-green-800' />
            <div className='absolute left-14 top-0 h-28 w-28 rounded-full bg-green-400' />
            <div className='absolute left-[10.5rem] top-0 h-14 w-14 rounded-br-full bg-green-200' />
          </div>
          <div className='relative order-6 col-span-full mt-16 h-28 lg:col-span-4 lg:col-start-2 lg:mt-36'>
            <div className='absolute left-0 top-14 h-14 w-14 rounded-full bg-purple-300' />
            <div className='absolute left-14 top-14 h-14 w-14 rounded-full bg-purple-900' />
            <div className='absolute left-14 top-0 h-14 w-14 rounded-tl-full bg-purple-600' />
            <div className='absolute left-[10.5rem] top-14 h-14 w-14 rounded-bl-full bg-purple-200' />
          </div>
          <div className='order-7 col-span-full mt-8 flex flex-col space-y-4 lg:col-span-6 lg:col-start-6 lg:mt-36 lg:space-y-7'>
            <span className='font-display text-xl font-medium'>
              Programmable Money
            </span>
            <span>
              Open Payments APIs allow any third-party to build applications
              that tightly integrate with your wallet for both sending and
              receiving payments.
            </span>
          </div>
          <div className='order-9 col-span-full mb-40 mt-8 flex flex-col space-y-4 lg:order-8 lg:col-span-6 lg:col-start-2 lg:mt-36 lg:space-y-7'>
            <span className='font-display text-xl font-medium'>
              Complete Control
            </span>
            <span>
              Every connected application has a unique connection to your
              account and you have complete control over how they can use it.
              Single payments, recurring payments, daily, weekly or monthly
              limits, you decide. You can change your mind and change or revoke
              their access any time.
            </span>
          </div>
          <div className='relative order-8 col-span-full mt-16 h-28 lg:order-9 lg:col-span-4 lg:col-start-9 lg:mt-36'>
            <div className='absolute left-0 top-14 h-14 w-14 rounded-tl-full bg-blue-500' />
            <div className='absolute left-14 top-14 h-14 w-14 rounded-tl-full bg-blue-100' />
            <div className='absolute left-28 top-0 h-14 w-14 rounded-tl-full bg-blue-900' />
            <div className='absolute left-28 top-14 h-14 w-14 rounded-full bg-blue-200' />
            <div className='absolute left-[10.5rem] top-14 h-14 w-14 rounded-full bg-blue-500' />
          </div>
        </section>
      </div>
      <section className='relative mx-auto  grid w-full grid-cols-4 content-start gap-4 gap-y-2 overflow-x-visible px-4 sm:max-w-lg sm:grid-cols-8 sm:px-0 lg:max-w-3xl lg:grid-cols-12 xl:max-w-[59rem]'>
        <div className='relative col-span-full h-20'>
          <div className='absolute -left-4 top-0 h-20 w-20 rounded-full bg-mk-section lg:-left-40' />
          <div className='absolute left-16 top-0 h-20 w-20 rounded-br-full bg-mk-section lg:-left-20' />
        </div>
        <div className='col-span-full mt-14'>
          <span className='font-display text-2xl lg:text-4xl'>
            The future of digital payments
          </span>
        </div>
        <div className='relative col-span-full mt-16 flex flex-col space-y-6 lg:col-span-3'>
          <div className='flex'>
            <Shape radius='rounded-bl-full' color='bg-slate-600' width='w-10' />
            <Shape radius='rounded-full' color='bg-rose-600' width='w-10' />
          </div>
          <span className='font-display font-medium'>Payment pointers</span>
          <span className='text-sm'>
            Step into the future with a payment pointer from Fynbos. Better than
            a credit/debit card or a private key, a payment pointer is a URL, a
            native building block of the Web.
          </span>
        </div>
        <div className='relative col-span-full mt-16 flex flex-col space-y-6 lg:col-span-3 lg:col-start-4'>
          <div className='flex'>
            <Shape radius='rounded-full' color='bg-slate-600' width='w-10' />
            <Shape radius='rounded-full' color='bg-rose-600' width='w-10' />
          </div>
          <span className='font-display font-medium'>Connect Securely</span>
          <span className='text-sm'>
            Your payment pointer connects third-party applications to your
            wallet. We use GNAP, the successor to OAuth 2.0 and OIDC being
            developed at IETF, to manage delegating access to your accounts to
            3rd party applications. This gives you fine-grained control over the
            connections to your wallet.
          </span>
        </div>
        <div className='relative col-span-full mt-16 flex flex-col space-y-6 lg:col-span-3 lg:col-start-7'>
          <div className='flex'>
            <Shape radius='rounded-none' color='bg-slate-600' width='w-10' />
            <Shape radius='rounded-tl-full' color='bg-rose-600' width='w-10' />
          </div>
          <span className='font-display font-medium'>Link accounts</span>
          <span className='text-sm'>
            Link your bank account or debit card to your Fynbos wallet, with
            support for more account types coming soon. Connect your wallet to
            applications that add new features and services to your accounts.
          </span>
        </div>
        <div className='relative col-span-full mt-16 flex flex-col space-y-6 lg:col-span-3 lg:col-start-10'>
          <div className='flex'>
            <Shape radius='rounded-full' color='bg-rose-600' width='w-10' />
            <Shape radius='rounded-br-full' color='bg-slate-600' width='w-10' />
          </div>
          <span className='font-display font-medium'>
            Open wallet ecosystem
          </span>
          <span className='text-sm'>
            Use your wallet to send and receive payments online.
          </span>
        </div>
        <div className='relative order-9 col-span-full mt-20 h-20'>
          <div className='absolute right-12 top-0 h-20 w-20 rounded-bl-full bg-mk-section lg:right-20' />
          <div className='absolute -right-8 top-0 h-20 w-20 rounded-tr-full bg-mk-section lg:right-0' />
          <div className='absolute -right-20 top-0 hidden h-20 w-20 rounded-br-full bg-mk-section lg:block' />
          <div className='absolute -right-40 top-0 hidden h-20 w-20 rounded-tr-full bg-mk-section lg:block' />
        </div>
      </section>
      <div className='w-full bg-mk-section'>
        {/*  <section className='mx-auto grid w-full grid-cols-4 content-start  gap-4 gap-y-2 overflow-x-visible px-4 sm:max-w-lg sm:grid-cols-8 sm:px-0 lg:max-w-3xl lg:grid-cols-12 xl:max-w-[59rem]'>*/}
        {/*    <div className='col-span-full mt-20'>*/}
        {/*      <span className='font-display text-2xl lg:text-4xl'>*/}
        {/*        Frequently asked questions*/}
        {/*      </span>*/}
        {/*    </div>*/}
        {/*    <div className='col-span-full mt-6 mb-20 rounded-xl border border-slate-200 bg-white px-4'>*/}
        {/*      <Disclosure>*/}
        {/*        {({ open }) => (*/}
        {/*          <>*/}
        {/*            <Disclosure.Button className='flex w-full items-center justify-between border-b border-slate-200 py-6 text-sm last:border-0 focus:outline-none focus-visible:ring focus-visible:ring-focus'>*/}
        {/*              <span className={`${open && 'text-primary'}`}>*/}
        {/*                What is your refund policy?*/}
        {/*              </span>*/}
        {/*              <Icon*/}
        {/*                className={`${*/}
        {/*                  open && 'rotate-180 transform text-primary'*/}
        {/*                }`}*/}
        {/*              >*/}
        {/*                expand_more*/}
        {/*              </Icon>*/}
        {/*            </Disclosure.Button>*/}
        {/*            <Disclosure.Panel className='border-b border-slate-200 py-4 text-sm last:border-0 last:pb-2'>*/}
        {/*              If you're unhappy with your purchase for any reason, email*/}
        {/*              us within 90 days and we'll refund you in full, no questions*/}
        {/*              asked.*/}
        {/*            </Disclosure.Panel>*/}
        {/*          </>*/}
        {/*        )}*/}
        {/*      </Disclosure>*/}
        {/*      <Disclosure as='div'>*/}
        {/*        {({ open }) => (*/}
        {/*          <>*/}
        {/*            <Disclosure.Button className='flex w-full justify-between border-b border-slate-200 py-6 text-sm last:border-0 focus:outline-none focus-visible:ring focus-visible:ring-focus'>*/}
        {/*              <span className={`${open && 'text-primary'}`}>*/}
        {/*                Do you offer technical support?*/}
        {/*              </span>*/}
        {/*              <Icon*/}
        {/*                className={`${*/}
        {/*                  open && 'rotate-180 transform text-primary'*/}
        {/*                }`}*/}
        {/*              >*/}
        {/*                expand_more*/}
        {/*              </Icon>*/}
        {/*            </Disclosure.Button>*/}
        {/*            <Disclosure.Panel className='border-b border-slate-200 py-4 text-sm last:border-0 last:pb-2'>*/}
        {/*              This is the first item's accordion body. It is shown by*/}
        {/*              default, until the collapse plugin adds the appropriate*/}
        {/*              classes that we use to style each element. These classes*/}
        {/*              control the overall appearance, as well as the showing and*/}
        {/*              hiding via CSS transitions. You can modify any of this with*/}
        {/*              custom CSS or overriding our default variables. It's also*/}
        {/*              worth noting that just about any HTML can go within the*/}
        {/*              .accordion-body, though the transition does limit overflow.*/}
        {/*            </Disclosure.Panel>*/}
        {/*          </>*/}
        {/*        )}*/}
        {/*      </Disclosure>*/}
        {/*    </div>*/}
        {/*    <div className='relative col-span-full h-20'>*/}
        {/*      <div className='absolute top-0 -left-8 h-20 w-20 rounded-tr-full bg-app lg:-left-40' />*/}
        {/*      <div className='absolute top-0 left-12 h-20 w-20 rounded-br-full bg-app lg:-left-20' />*/}
        {/*    </div>*/}
        {/*  </section>*/}
        {/*</div>*/}
        <section className='mx-auto grid w-full grid-cols-4 content-start gap-4 gap-y-2 overflow-x-visible px-4 sm:max-w-lg sm:grid-cols-8 sm:px-0 lg:max-w-3xl lg:grid-cols-12 xl:max-w-[59rem]'>
          <div className='col-span-full mt-12 flex items-center justify-center lg:col-span-6 lg:col-start-2 lg:my-20 lg:justify-start'>
            <span className='font-display text-3xl font-medium'>
              Ready to get started?
            </span>
          </div>
          <div className='col-span-full mb-12 mt-6 lg:col-span-4 lg:col-start-8 lg:my-20'>
            <ButtonRouter shrink to={route('/waitlist')}>
              Join the waitlist
            </ButtonRouter>
          </div>
          {/*<div className='col-span-full mb-12 mt-2 lg:col-span-2 lg:col-start-10 lg:my-20'>*/}
          {/*<Router*/}
          {/*  to={route('/contact')}*/}
          {/*  className='flex h-[50px] w-full items-center justify-center rounded-full border border-focus'*/}
          {/*>*/}
          {/*  <span className='font-display font-medium text-primary'>*/}
          {/*    Contact us*/}
          {/*  </span>*/}
          {/*</Router>*/}
          {/*</div>*/}
        </section>
      </div>
    </main>
  )
}

function AppPage() {
  const {
    firstName,
    paymentPointer,
    snackbar,
    transactions,
    kycStatus,
    canTopUp,
    nextStep,
    pusherArgs
  } = useLoaderData<typeof loader>()

  const [snackbarState, setSnackbar] = useState<any>(snackbar)
  const [showSnackbar, setShowSnackbar] = useState<boolean>(
    snackbar.show ?? false
  )

  usePusher(pusherArgs, ['transaction', 'kyc'])

  return (
    <WalletGrid>
      <Card className='col-span-full sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <div className='mt-2'>
          <HomeShapes />
        </div>
        <h1 className='mt-6 text-center font-display text-2xl font-medium'>
          Welcome {firstName}
        </h1>
        {kycStatus != KycStatus.Verified && (
          <p className='mb-2 mt-4 text-center'>
            Thank you for signing up to Fynbos.
          </p>
        )}
        {kycStatus == KycStatus.Verified && !canTopUp && (
          <p className='mb-2 mt-4 text-center'>
            Your payment pointer is activated. You are now able to send and
            receive payments.
          </p>
        )}
        {kycStatus == KycStatus.Verified && canTopUp && (
          <p className='mb-2 mt-4 text-center'>
            Send and receive payments with your payment pointer.
          </p>
        )}
        {kycStatus == KycStatus.Verified && (
          <button
            type='button'
            onClick={async () => {
              if (typeof navigator.clipboard == 'undefined') {
                setSnackbar({
                  message: "Couldn't copy to clipboard.",
                  icon: 'close',
                  show: true
                })
                setShowSnackbar(true)
              } else
                navigator.clipboard.writeText(paymentPointer.formatted).then(
                  () => {
                    setSnackbar({
                      message: 'Payment pointer copied to clipboard.',
                      icon: 'close',
                      show: true
                    })
                    setShowSnackbar(true)
                  },
                  () => {
                    setSnackbar({
                      message: "Couldn't copy to clipboard.",
                      icon: 'close',
                      show: true
                    })
                    setShowSnackbar(true)
                  }
                )
            }}
            className='mt-4 flex flex items-center justify-between rounded-xl bg-nav p-4 hover:bg-nav-hover'
          >
            <span className='font-medium text-medium'>
              {paymentPointer.formatted}
            </span>
            <Icon className='text-medium'>content_copy</Icon>
          </button>
        )}
      </Card>

      {kycStatus == KycStatus.InProgress && (
        <Card className='col-span-full sm:col-span-6 sm:col-start-2 lg:col-start-4'>
          <h2 className='font-display text-lg font-medium'>
            Activation pending
          </h2>
          <p className='mt-4'>Just a moment, we are verifying your details.</p>
        </Card>
      )}
      {/*{kycStatus == KycStatus.Retry && (*/}
      {/*  <Card className='col-span-full sm:col-span-6 sm:col-start-2 lg:col-start-4'>*/}
      {/*    <h2 className='font-display text-lg font-medium'>*/}
      {/*      Activation failed*/}
      {/*    </h2>*/}
      {/*    <p className='mt-4'>*/}
      {/*      Some of the details you provided were not correct. Please fix them*/}
      {/*      and submit again.*/}
      {/*    </p>*/}
      {/*    <Router*/}
      {/*      className='mt-4 text-sm font-medium text-primary'*/}
      {/*      to={route('/personal-details')}*/}
      {/*    >*/}
      {/*      Fix personal details*/}
      {/*    </Router>*/}
      {/*  </Card>*/}
      {/*)}*/}
      {kycStatus == KycStatus.Suspended && (
        <Card className='col-span-full sm:col-span-6 sm:col-start-2 lg:col-start-4'>
          <h2 className='font-display text-lg font-medium'>
            Activation failed
          </h2>
          <p className='mt-4'>
            We could not verify your identity, please contact support to
            continue.
          </p>
          <Router
            className='mt-4 text-sm font-medium text-primary'
            to={route('/support')}
          >
            Contact support
          </Router>
        </Card>
      )}
      {/*{kycStatus == KycStatus.ReviewPending && (*/}
      {/*  <Card className='col-span-full sm:col-span-6 sm:col-start-2 lg:col-start-4'>*/}
      {/*    <h2 className='font-display text-lg font-medium'>*/}
      {/*      Reviewing activation*/}
      {/*    </h2>*/}
      {/*    <p className='mt-4'>*/}
      {/*      We need to manually review your verification details. We will notify*/}
      {/*      you when this process completes.*/}
      {/*    </p>*/}
      {/*  </Card>*/}
      {/*)}*/}

      {nextStep.show && (
        <Card className='col-span-full space-y-6 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
          <h1 className='font-display text-lg font-medium'>Next step</h1>
          <div className='flex items-start space-x-4'>
            <div className='flex items-center justify-between rounded-full bg-nav p-5 text-medium'>
              <Icon>{nextStep.icon}</Icon>
            </div>
            <div className='flex flex-col space-y-4'>
              <p className='text-sm text-medium'>{nextStep.title}</p>
              <Router
                className='text-sm font-medium text-primary'
                to={nextStep.action.to}
              >
                {nextStep.action.text}
              </Router>
            </div>
          </div>
        </Card>
      )}

      {kycStatus == KycStatus.Verified && (
        <Card className='col-span-full sm:col-span-6 sm:col-start-2 lg:col-start-4'>
          <div className='flex items-center justify-between'>
            <h1 className='font-display text-lg font-medium'>
              Latest transactions
            </h1>
            <Router className='flex max-h-fit' to={route('/transactions')}>
              <Icon className='text-medium'>read_more</Icon>
            </Router>
          </div>
          {transactions.length == 0 && (
            <div className='mt-4 flex flex-col space-y-4'>
              <span className='text-sm text-medium'>
                Your payment activity will appear here once you start using your
                payment pointer.
              </span>
              <Router
                to={route('/pay')}
                className='text-sm font-medium text-primary'
              >
                Send or receive payments now
              </Router>
            </div>
          )}
          {transactions.map((transaction, index) => (
            <Fragment key={transaction.id}>
              {(index == 0 ||
                transaction.date != transactions[index - 1].date) && (
                <span className='mt-6 text-xs text-medium'>
                  {transaction.date}
                </span>
              )}
              <Router
                to={`/transaction/${transaction.type}/${transaction.id}`}
                className='mt-2 flex w-full justify-between'
              >
                <div className='flex space-x-1'>
                  {transaction.icon == 'schedule' && (
                    <div className='mt-0.5'>
                      <AnimatedSchedule />
                    </div>
                  )}
                  {transaction.icon != 'schedule' && (
                    <Icon className='mt-0.5 text-medium'>
                      {transaction.icon}
                    </Icon>
                  )}
                  <div className='flex flex-col space-y-2'>
                    <span className='text-medium'>{transaction.title}</span>
                    <span className='text-xs text-medium'>
                      {transaction.time}
                    </span>
                  </div>
                </div>
                <span className='font-medium'>{transaction.total}</span>
              </Router>
            </Fragment>
          ))}
        </Card>
      )}
      <Snackbar
        message={snackbarState.message}
        action={snackbarState.action}
        icon={snackbarState.icon}
        show={showSnackbar}
        id='cookie-snackbar'
        dismissAfter={3000}
        offset
        onClose={() => setShowSnackbar(false)}
      />
    </WalletGrid>
  )
}
