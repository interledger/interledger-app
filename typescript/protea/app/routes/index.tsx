import type { LoaderArgs } from '@remix-run/node'
import { json } from '@remix-run/node'
import { useLoaderData } from '@remix-run/react'
import { route } from 'routes-gen'
import {
  ButtonRouter,
  HomeShapes,
  Icon,
  Layouts,
  Router,
  Shape,
  WalletGrid
} from '~/components'
import { hasUserSession, requireUserSession } from '~/lib/kratos.server'
import type { Transaction } from '~/lib/wallet.server'
import {
  getPendingTransactions,
  getTransactions,
  getWalletBalance,
  getWalletPaymentPointer
} from '~/lib/wallet.server'
import { Fragment } from 'react'

export async function loader({ request }: LoaderArgs) {
  const isUser = await hasUserSession(request)

  let data = {
    isUser: isUser,
    firstName: '',
    paymentPointer: {
      url: '',
      asset: 'USD',
      assetScale: 2,
      alias: 'default',
      walletID: '',
      formatted: ''
    },
    balance: '',
    transactions: [] as Transaction[]
  }

  if (isUser) {
    const [
      session,
      paymentPointer,
      balance,
      pendingTransactions,
      transactions
    ] = await Promise.all([
      requireUserSession(request),
      getWalletPaymentPointer(request),
      getWalletBalance(request),
      getPendingTransactions(request),
      getTransactions(request, { page: 1, pageSize: 3 })
    ])

    /** TODO whatNext state machine
     * Verify user data (Get cash balance)
     * Set up receiving
     * Set up sending
     * Make a payment / share payment pointer
     * Set up payouts ?
     */

    data = {
      ...data,
      firstName: session.identity.traits.firstName,
      paymentPointer,
      balance,
      transactions: [...pendingTransactions, ...transactions]
    }
  }
  return json(data)
}

export const handle = {
  layout: (isUser: boolean) =>
    isUser ? Layouts.WalletLayout : Layouts.LandingLayout
}

export default function Page() {
  const { isUser } = useLoaderData<typeof loader>()

  if (isUser) return <AppPage />
  else return <MarketingPage />
}

function MarketingPage() {
  return (
    <main className='w-full overflow-hidden'>
      <section className='relative mx-auto grid w-full grid-cols-4 content-start gap-4 gap-y-2 overflow-x-visible px-4 sm:max-w-lg sm:grid-cols-8 sm:px-0 lg:max-w-3xl lg:grid-cols-12 xl:max-w-[59rem]'>
        <div className='relative col-span-full h-20 lg:h-48'>
          <div className='absolute -right-20 top-0 hidden h-20 w-20 rounded-br-full bg-rose-300 lg:block' />
          <div className='absolute right-0 top-20 hidden h-20 w-20 rounded-br-full bg-slate-300 lg:block' />
          <div className='absolute top-0 left-0 hidden h-20 w-20 rounded-bl-full bg-lime-300 lg:block' />
          <div className='absolute top-20 -left-20 hidden h-20 w-20 rounded-br-full bg-slate-100 lg:block' />
          {/* Mobile */}
          <div className='absolute top-0 -left-4 block h-20 w-20 rounded-br-full bg-slate-100 lg:hidden' />
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
        <div className='col-span-full mt-10 flex flex-col items-center justify-center space-y-4 lg:mt-5 lg:flex-row lg:space-y-0 lg:space-x-4'>
          <ButtonRouter shrink to={route('/waitlist')}>
            Join the waitlist
          </ButtonRouter>
        </div>
        <div className='relative col-span-full h-48 lg:h-56'>
          <div className='absolute -left-[calc(100vw)] bottom-20 hidden h-20 w-screen bg-slate-700 lg:block' />
          <div className='absolute -left-20 bottom-0 hidden h-20 w-20 rounded-tl-full bg-slate-700 lg:block' />
          <div className='absolute left-0 bottom-0 hidden h-20 w-20 rounded-br-full bg-lime-500 lg:block' />
          <div className='absolute right-20 bottom-20 hidden h-20 w-20 rounded-tl-full bg-lime-300 lg:block' />
          <div className='absolute right-0 bottom-20 hidden h-20 w-20 rounded-tr-full bg-lime-500 lg:block' />
          <div className='absolute right-20 bottom-0 hidden h-20 w-20 rounded-full bg-rose-500 lg:block' />
          <div className='absolute -right-[calc(100vw-5rem)] bottom-0 hidden h-20 w-screen rounded-bl-full bg-yellow-200 lg:block' />
          {/* Mobile */}
          <div className='absolute -right-4 bottom-0 block h-20 w-20 rounded-br-full bg-lime-300 lg:hidden' />
          <div className='absolute right-16 bottom-0 block h-20 w-20 rounded-bl-full bg-lime-500 lg:hidden' />
          <div className='absolute -right-4 bottom-20 block h-20 w-20 rounded-full bg-rose-500 lg:hidden' />
          <div className='absolute right-16 bottom-20 block h-20 w-20 rounded-tr-full bg-yellow-200 lg:hidden' />
        </div>
      </section>
      <div className='w-full bg-slate-50'>
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
          <div className='order-9 col-span-full mt-8 mb-40 flex flex-col space-y-4 lg:order-8 lg:col-span-6 lg:col-start-2 lg:mt-36 lg:space-y-7'>
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
          <div className='absolute top-0 -left-4 h-20 w-20 rounded-full bg-slate-50 lg:-left-40' />
          <div className='absolute top-0 left-16 h-20 w-20 rounded-br-full bg-slate-50 lg:-left-20' />
        </div>
        <div className='col-span-full mt-14'>
          <span className='font-display text-2xl lg:text-4xl'>
            The future of digital payments
          </span>
        </div>
        <div className='relative col-span-full mt-16 flex flex-col space-y-6 lg:col-span-3'>
          <div className='flex'>
            <Shape radius='rounded-bl-full' color='bg-slate-300' width='w-10' />
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
            <Shape radius='rounded-full' color='bg-slate-300' width='w-10' />
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
            <Shape radius='rounded-none' color='bg-slate-300' width='w-10' />
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
            <Shape radius='rounded-br-full' color='bg-slate-300' width='w-10' />
          </div>
          <span className='font-display font-medium'>
            Open wallet ecosystem
          </span>
          <span className='text-sm'>
            Use your wallet to send and receive payments online.
          </span>
        </div>
        <div className='relative order-9 col-span-full mt-20 h-20'>
          <div className='absolute top-0 right-12 h-20 w-20 rounded-bl-full bg-slate-50 lg:right-20' />
          <div className='absolute top-0 -right-8 h-20 w-20 rounded-tr-full bg-slate-50 lg:right-0' />
          <div className='absolute top-0 -right-20 hidden h-20 w-20 rounded-br-full bg-slate-50 lg:block' />
          <div className='absolute top-0 -right-40 hidden h-20 w-20 rounded-tr-full bg-slate-50 lg:block' />
        </div>
      </section>
      <div className='w-full bg-slate-50'>
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
  const { firstName, paymentPointer, balance, transactions } =
    useLoaderData<typeof loader>()
  return (
    <WalletGrid>
      <div className='col-span-full flex flex-col rounded-2xl bg-page p-4 pb-8 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <div className='mt-2'>
          <HomeShapes />
        </div>
        <h1 className='mt-6 font-display text-2xl'>Welcome {firstName}</h1>
        <p className='mt-6'>
          Your payment pointer has been set up successfully.
        </p>
        <button
          type='button'
          onClick={async () => {
            navigator.clipboard.writeText(paymentPointer.formatted).then(
              () => {
                // TODO show success snackbar.
                console.log('Successfully set payment pointer to clipboard.')
              },
              () => {
                console.log("Couldn't set payment pointer to clipboard.")
              }
            )
          }}
          className='mt-4 flex justify-between rounded-xl bg-container p-4'
        >
          <span className='font-medium text-medium'>
            {paymentPointer.formatted}
          </span>
          <Icon className='text-medium'>content_copy</Icon>
        </button>
      </div>

      {balance && (
        <div className='col-span-full flex justify-between rounded-2xl bg-page p-4 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
          <h1 className='font-display text-lg font-medium'>Cash balance</h1>
          <h1 className='font-display text-lg font-medium'>{balance}</h1>
        </div>
      )}

      <div className='col-span-full flex flex-col space-y-6 rounded-2xl bg-page p-4 pb-6 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <h1 className='font-display text-lg font-medium'>What to do next</h1>
        <div className='flex items-start space-x-4'>
          <div className='flex items-center justify-between rounded-full bg-container p-5 text-medium'>
            <Icon>account_balance</Icon>
          </div>
          <div className='flex flex-col space-y-2'>
            <h1 className='font-medium text-medium'>Receive money</h1>
            <p className='text-sm text-medium'>
              Receive money into your cash balance.
            </p>
            <Router
              className='text-sm font-medium text-primary'
              to={route('/linked-account/:type', { type: 'bank' })}
            >
              Enable receiving
            </Router>
          </div>
        </div>
        <div className='flex items-start space-x-4'>
          <div className='flex items-center justify-between rounded-full bg-container p-5 text-medium'>
            <Icon>credit_card</Icon>
          </div>
          <div className='flex flex-col space-y-2'>
            <h1 className='font-medium text-medium'>Send money</h1>
            <p className='text-sm text-medium'>
              Easily send money from a debit card.
            </p>
            <Router
              className='text-sm font-medium text-primary'
              to={route('/linked-account/:type', { type: 'card' })}
            >
              Enable sending
            </Router>
          </div>
        </div>
      </div>
      <div className='col-span-full flex flex-col rounded-2xl bg-page p-4 pb-6 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
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
              to={route('/transaction/:type/:transactionId', {
                type: transaction.type,
                transactionId: transaction.id
              })}
              className='mt-2 flex w-full justify-between'
            >
              <div className='flex space-x-1'>
                <Icon className='mt-0.5 text-medium'>{transaction.icon}</Icon>
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
      </div>
    </WalletGrid>
  )
}
