import { Disclosure } from '@headlessui/react'
import type { LoaderArgs } from '@remix-run/node'
import { json } from '@remix-run/node'
import { Link, useLoaderData } from '@remix-run/react'
import { DateTime } from 'luxon'
import type { FC } from 'react'
import { Fragment } from 'react'
import { route } from 'routes-gen'
import { Icon, Router } from '~/components'
import type { HomeQuery, HomeQueryVariables } from '~/generated/types'
import { HomeDocument, TransactionType } from '~/generated/types'
import { apolloClient } from '~/lib/apollo.server'
import { hasUserSession, requireUserSession } from '~/lib/kratos.server'

type Activity = {
  id: string
  amount: string
  transactionType: TransactionType
  title: string
  description: string
  status: string
}

type Activities = {
  date: string
  activities: Activity[]
}

export async function loader({ request }: LoaderArgs) {
  const isUser = await hasUserSession(request)

  let data = {
    isUser: isUser,
    balance: '0',
    recentActivities: [] as Activities[],
    pendingTransactions: [] as Activity[]
  }

  if (isUser) {
    await requireUserSession(request)
    const cookie = request.headers.get('cookie')

    const account = await apolloClient
      .query<HomeQuery, HomeQueryVariables>({
        query: HomeDocument,
        context: {
          headers: {
            cookie: cookie
          }
        }
      })
      .then((val) => val.data.account)

    data.balance = account?.balance as string

    if (typeof account?.recentTransactions !== 'undefined')
      for (let trx of account?.recentTransactions) {
        const activity = {
          id: trx.id,
          amount: trx.amount,
          transactionType: trx.type,
          title: activityTitle(trx.type),
          description: trx.description,
          status: trx.status
        }
        if (activity.status == 'pending') {
          data.pendingTransactions.push(activity)
          continue
        }
        const date = DateTime.fromISO(trx.timestamp).toFormat('dd LLLL yyyy')
        const indexToPush = data.recentActivities.findIndex(
          (val) => val.date == date
        )
        if (indexToPush >= 0)
          data.recentActivities[indexToPush].activities.push(activity)
        else
          data.recentActivities.push({
            date: date,
            activities: [activity]
          })
      }
  }
  return json(data)
}

export default function Page() {
  const { isUser } = useLoaderData<typeof loader>()

  if (isUser) return <AppPage />
  else return <MarketingPage />
}

function MarketingPage() {
  return (
    <main className='w-full overflow-hidden'>
      <section className='relative mx-auto grid w-full grid-cols-4 content-start gap-4 gap-y-2 overflow-x-visible px-8 sm:max-w-lg sm:grid-cols-8 sm:px-0 lg:max-w-3xl lg:grid-cols-12 xl:max-w-4xl'>
        <div className='relative col-span-full h-20 lg:h-48'>
          <div className='absolute -right-40 top-0 hidden h-20 w-20 rounded-br-full bg-rose-300 lg:block' />
          <div className='absolute right-0 top-20 hidden h-20 w-20 rounded-br-full bg-slate-300 lg:block' />
          <div className='absolute top-0 left-0 hidden h-20 w-20 rounded-bl-full bg-lime-300 lg:block' />
          <div className='absolute top-20 -left-20 hidden h-20 w-20 rounded-br-full bg-slate-100 lg:block' />
          {/* Mobile */}
          <div className='absolute top-0 -left-8 block h-20 w-20 rounded-br-full bg-slate-100 lg:hidden' />
        </div>
        <div className='col-span-full'>
          <span className='flex justify-center font-display text-3xl font-medium lg:text-6xl'>
            The better way to pay
          </span>
        </div>
        <div className='col-span-full lg:col-span-10 lg:col-start-2 lg:mt-7'>
          <span className='flex text-center font-sans text-base font-normal lg:text-2xl'>
            With a payment pointer from Fynbos you get a connected account that
            is simple, secure and programmable.
          </span>
        </div>
        <div className='col-span-full mt-10 flex flex-col justify-center space-y-4 lg:mt-5 lg:flex-row lg:space-y-0 lg:space-x-4'>
          <Router
            to={route('/signup')}
            className='flex h-[50px] w-full items-center justify-center rounded-full bg-primary px-10 sm:max-w-fit'
          >
            <span className='font-display text-base font-medium text-white'>
              Get a payment pointer
            </span>
          </Router>
          <Router
            to={route('/signup')}
            className='flex h-[50px] w-full items-center justify-center rounded-full border border-focus px-10 sm:max-w-fit'
          >
            <span className='font-display text-base font-medium text-primary'>
              Learn more
            </span>
          </Router>
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
          <div className='absolute right-12 bottom-20 block h-20 w-20 rounded-tl-full bg-lime-300 lg:hidden' />
          <div className='absolute -right-8 bottom-20 block h-20 w-20 rounded-tr-full bg-lime-500 lg:hidden' />
          <div className='absolute right-12 bottom-0 block h-20 w-20 rounded-full bg-rose-500 lg:hidden' />
          <div className='absolute -right-8 bottom-0 block h-20 w-20 rounded-bl-full bg-yellow-200 lg:hidden' />
        </div>
      </section>
      <div className='w-full bg-slate-50'>
        <section className='mx-auto grid w-full grid-cols-4 content-start gap-4 gap-y-2 overflow-x-visible px-8 sm:max-w-lg sm:grid-cols-8 sm:px-0 lg:max-w-3xl lg:grid-cols-12 xl:max-w-4xl'>
          <div className='order-1 col-span-full pt-20 lg:col-span-10 lg:col-start-2'>
            <span className='font-display text-2xl font-normal lg:text-4xl lg:leading-[2.75rem]'>
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
            <span className='font-sans text-base font-normal'>
              Everything you need and nothing you don't. A clean and simple user
              interface for managing your connected account. Make deposits and
              withdrawals, send and receive payments using payment pointers, and
              configure connected applications.
            </span>
          </div>
          <div className='order-5 col-span-full mt-8 flex flex-col space-y-4 lg:order-4 lg:col-span-6 lg:col-start-2 lg:mt-36 lg:space-y-7'>
            <span className='font-display text-xl font-medium'>Secure</span>
            <span className='font-sans text-base font-normal'>
              A FDIC-insured direct deposit account backed by Piermont Bank,
              giving you the freedom to build in the knowledge that your money
              is safe.
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
            <span className='font-sans text-base font-normal'>
              Open Payments APIs allow any third-party to build applications
              that tightly integrate your account for both sending and receiving
              payments via the Interledger protocol.
            </span>
          </div>
          <div className='order-9 col-span-full mt-8 mb-40 flex flex-col space-y-4 lg:order-8 lg:col-span-6 lg:col-start-2 lg:mt-36 lg:space-y-7'>
            <span className='font-display text-xl font-medium'>
              Complete Control
            </span>
            <span className='font-sans text-base font-normal'>
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
      <section className='relative mx-auto  grid w-full grid-cols-4 content-start gap-4 gap-y-2 overflow-x-visible px-8 sm:max-w-lg sm:grid-cols-8 sm:px-0 lg:max-w-3xl lg:grid-cols-12 xl:max-w-4xl'>
        <div className='relative col-span-full h-20'>
          <div className='absolute top-0 -left-8 h-20 w-20 rounded-full bg-slate-50 lg:-left-40' />
          <div className='absolute top-0 left-12 h-20 w-20 rounded-br-full bg-slate-50 lg:-left-20' />
        </div>
        <div className='col-span-full mt-14'>
          <span className='font-display text-2xl font-normal lg:text-4xl'>
            The future of digital payments
          </span>
        </div>
        <div className='relative order-1 col-span-full h-28 lg:col-span-3'>
          <div className='absolute left-0 bottom-0 h-10 w-10 rounded-bl-full bg-slate-300' />
          <div className='absolute left-10 bottom-0 h-10 w-10 rounded-br-full bg-slate-300' />
          <div className='absolute left-8 bottom-4 h-12 w-12 rounded-full bg-rose-600' />
        </div>
        <div className='relative order-3 col-span-full h-28 lg:order-2 lg:col-span-3 lg:col-start-4'>
          <div className='absolute left-8 bottom-0 h-12 w-12 rounded-full bg-rose-600' />
          <div className='absolute left-0 bottom-0 h-12 w-12 rounded-full bg-slate-300' />
        </div>
        <div className='relative order-5 col-span-full h-28 lg:order-3 lg:col-span-3 lg:col-start-7'>
          <div className='absolute left-6 bottom-2 h-10 w-10 rotate-45 bg-rose-600' />
          <div className='absolute left-0 bottom-2 h-10 w-10 rotate-45 bg-slate-300' />
        </div>
        <div className='relative order-7 col-span-full h-28 lg:order-4 lg:col-span-3 lg:col-start-10'>
          <div className='absolute left-12 bottom-0 h-12 w-12 rounded-bl-full bg-rose-600' />
          <div className='absolute left-0 bottom-0 h-12 w-12 rounded-bl-full bg-slate-300' />
        </div>
        <div className='relative order-2 col-span-full mt-8 flex flex-col lg:order-5 lg:col-span-3'>
          <span className='font-display text-base font-medium'>
            Payment pointers
          </span>
          <span className='mt-6 font-sans text-sm font-normal'>
            Step into the future with a payment pointer from Fynbos. Better than
            a credit/debit card or a private key, a payment pointer is a URL, a
            native building block of the Web. Use it to send and receive
            payments online or connect your account to applications that add new
            features or services to your account.
          </span>
        </div>
        <div className='relative order-4 col-span-full mt-8 flex flex-col lg:order-6 lg:col-span-3 lg:col-start-4'>
          <span className='font-display text-base font-medium'>
            Connect Securely
          </span>
          <span className='mt-6 font-sans text-sm font-normal'>
            Your payment pointer connects third-party applications to your
            account. We use GNAP, the successor to OAuth 2.0 and OIDC being
            developed at IETF, to manage delegating access to your account to
            3rd party applications. This gives you fine-grained control over the
            connection to your account.
          </span>
        </div>
        <div className='relative order-6 col-span-full mt-8 flex flex-col lg:order-7 lg:col-span-3 lg:col-start-7'>
          <span className='font-display text-base font-medium'>
            Withdraw and Spend
          </span>
          <span className='mt-6 font-sans text-sm font-normal'>
            Make deposits into your account using a linked debit card or bank
            account with the funds available immediately to send and spend.
          </span>
        </div>
        <div className='relative order-8 col-span-full mt-8 flex flex-col items-start lg:order-8 lg:col-span-3 lg:col-start-10'>
          <span className='font-display text-base font-medium'>
            Access Deposits Instantly
          </span>
          <span className='mt-6 font-sans text-sm font-normal'>
            Use ACH to withdraw funds from your account or use your virtual
            credit card to spend where payment pointers aren't yet accepted.
          </span>
          <div className='mt-8 flex rounded-lg bg-container px-3 py-1.5'>
            <span className='font-sans text-sm font-normal text-medium'>
              Coming soon
            </span>
          </div>
        </div>
        <div className='relative order-9 col-span-full h-20'>
          <div className='absolute top-0 right-12 h-20 w-20 rounded-bl-full bg-slate-50 lg:right-20' />
          <div className='absolute top-0 -right-8 h-20 w-20 rounded-tr-full bg-slate-50 lg:right-0' />
          <div className='absolute top-0 -right-20 hidden h-20 w-20 rounded-br-full bg-slate-50 lg:block' />
          <div className='absolute top-0 -right-40 hidden h-20 w-20 rounded-tr-full bg-slate-50 lg:block' />
        </div>
      </section>
      <div className='w-full bg-slate-50'>
        <section className='mx-auto grid w-full grid-cols-4 content-start  gap-4 gap-y-2 overflow-x-visible px-8 sm:max-w-lg sm:grid-cols-8 sm:px-0 lg:max-w-3xl lg:grid-cols-12 xl:max-w-4xl'>
          <div className='col-span-full mt-20'>
            <span className='font-display text-2xl font-normal lg:text-4xl'>
              Frequently asked questions
            </span>
          </div>
          <div className='col-span-full mt-6 mb-20 rounded-xl border border-slate-200 bg-white px-8'>
            <Disclosure>
              {({ open }) => (
                <>
                  <Disclosure.Button className='flex w-full items-center justify-between border-b border-slate-200 py-6 font-sans text-sm font-normal last:border-0 focus:outline-none focus-visible:ring focus-visible:ring-primary'>
                    <span className={`${open && 'text-primary'}`}>
                      What is your refund policy?
                    </span>
                    <Icon
                      className={`${
                        open && 'rotate-180 transform text-primary'
                      }`}
                    >
                      expand_more
                    </Icon>
                  </Disclosure.Button>
                  <Disclosure.Panel className='border-b border-slate-200 py-4 text-sm last:border-0 last:pb-2'>
                    If you're unhappy with your purchase for any reason, email
                    us within 90 days and we'll refund you in full, no questions
                    asked.
                  </Disclosure.Panel>
                </>
              )}
            </Disclosure>
            <Disclosure as='div'>
              {({ open }) => (
                <>
                  <Disclosure.Button className='flex w-full justify-between border-b border-slate-200 py-6 font-sans text-sm font-normal last:border-0 focus:outline-none focus-visible:ring focus-visible:ring-primary'>
                    <span className={`${open && 'text-primary'}`}>
                      Do you offer technical support?
                    </span>
                    <Icon
                      className={`${
                        open && 'rotate-180 transform text-primary'
                      }`}
                    >
                      expand_more
                    </Icon>
                  </Disclosure.Button>
                  <Disclosure.Panel className='border-b border-slate-200 py-4 text-sm last:border-0 last:pb-2'>
                    This is the first item's accordion body. It is shown by
                    default, until the collapse plugin adds the appropriate
                    classes that we use to style each element. These classes
                    control the overall appearance, as well as the showing and
                    hiding via CSS transitions. You can modify any of this with
                    custom CSS or overriding our default variables. It's also
                    worth noting that just about any HTML can go within the
                    .accordion-body, though the transition does limit overflow.
                  </Disclosure.Panel>
                </>
              )}
            </Disclosure>
          </div>
          <div className='relative col-span-full h-20'>
            <div className='absolute top-0 -left-8 h-20 w-20 rounded-tr-full bg-app lg:-left-40' />
            <div className='absolute top-0 left-12 h-20 w-20 rounded-br-full bg-app lg:-left-20' />
          </div>
        </section>
      </div>
      <section className='mx-auto grid w-full grid-cols-4 content-start gap-4 gap-y-2 overflow-x-visible px-8 sm:max-w-lg sm:grid-cols-8 sm:px-0 lg:max-w-3xl lg:grid-cols-12 xl:max-w-4xl'>
        <div className='col-span-full mt-12 flex items-center justify-center lg:col-span-6 lg:col-start-2 lg:my-20 lg:justify-start'>
          <span className='font-display text-3xl font-medium'>
            Ready to get started?
          </span>
        </div>
        <div className='col-span-full mt-6 lg:col-span-2 lg:col-start-8 lg:my-20'>
          <Router
            to={route('/signup')}
            className='flex h-[50px] w-full items-center justify-center rounded-full bg-primary'
          >
            <span className='font-display text-base font-medium text-white'>
              Sign up
            </span>
          </Router>
        </div>
        <div className='col-span-full mb-12 mt-2 lg:col-span-2 lg:col-start-10 lg:my-20'>
          <Router
            to={route('/contact')}
            className='flex h-[50px] w-full items-center justify-center rounded-full border border-focus'
          >
            <span className='font-display text-base font-medium text-primary'>
              Contact us
            </span>
          </Router>
        </div>
      </section>
    </main>
  )
}

function AppPage() {
  const { balance, recentActivities, pendingTransactions } =
    useLoaderData<typeof loader>()
  return (
    <div className='w-full'>
      {/* Header */}
      <header className='sticky top-0 flex h-16 min-w-full select-none items-center justify-between bg-app p-4 text-medium'>
        <div className='flex items-center justify-start font-display text-2xl font-medium'>
          Home
        </div>
        <Link className='sm:hidden' to={route('/settings')}>
          <div className='-mr-3 p-3 text-medium'>
            <Icon>settings</Icon>
          </div>
        </Link>
      </header>
      {/* Body */}
      <div className='mx-auto grid min-h-[calc(100vh-9rem)] w-full grid-cols-4 content-start gap-4 gap-y-2 overflow-y-auto p-4 pb-24 sm:max-w-lg sm:grid-cols-8 sm:px-0 lg:max-w-3xl lg:grid-cols-12 xl:max-w-4xl'>
        {/* HOME */}
        <div className='col-span-full flex flex-col items-center px-3 pt-4 pb-2 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
          <span className='font-sans text-base font-normal'>Balance</span>
          <span className='font-display text-4xl font-normal'>{balance}</span>
        </div>
        <div className='col-span-full flex justify-center space-x-3 py-4 px-3 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
          <Router
            className='rounded-full'
            to={route('/flows/:flowId/deposit/payment-method', {
              flowId: 'init'
            })}
          >
            <div className='flex h-10 w-36 cursor-pointer items-center justify-center rounded-full bg-container-primary font-display text-sm font-medium text-medium hover:bg-container-primary-hover active:bg-container-primary-active'>
              Deposit
            </div>
          </Router>
          <Router
            className='rounded-full'
            to={route('/flows/:flowId/withdraw/payment-method', {
              flowId: 'init'
            })}
          >
            <div className='flex h-10 w-36 cursor-pointer items-center justify-center rounded-full bg-container-primary font-display text-sm font-medium text-medium hover:bg-container-primary-hover active:bg-container-primary-active'>
              Withdraw
            </div>
          </Router>
        </div>
        {recentActivities.length > 0 && (
          <div className='col-span-full flex justify-between pt-2 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
            <span className='font-display text-lg font-medium'>
              Recent activity
            </span>
            <Router to={route('/activity')}>
              <div className='flex items-center space-x-1 font-display text-sm font-medium text-primary'>
                <span>See all</span>
                <Icon>read_more</Icon>
              </div>
            </Router>
          </div>
        )}
        {/* Activity items */}
        {recentActivities.map((activities) => (
          <Fragment key={activities.date}>
            <span className='col-span-full ml-4 mt-2 font-display text-xs font-normal sm:col-span-6 sm:col-start-2 lg:col-start-4'>
              {activities.date}
            </span>
            {activities.activities.map((activity) => (
              <ActivityCard key={activity.id} activity={activity} />
            ))}
          </Fragment>
        ))}
        {pendingTransactions.length > 0 && (
          <div className='col-span-full flex justify-start pt-4 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
            <span className='font-display text-lg font-medium'>Pending</span>
          </div>
        )}
        {pendingTransactions.map((activity) => (
          <ActivityCard key={activity.id} activity={activity} />
        ))}
      </div>
    </div>
  )
}

// TODO: Replace with implementation in the backend.
const activityIcon = (type: TransactionType, status: string) => {
  switch (status) {
    case 'pending':
      return 'hourglass_empty'
    default:
      break
  }
  switch (type) {
    // case TransactionType.Received:
    //   return <ReceivedIcon />
    case TransactionType.Outgoingpayment:
      return 'north_east'
    case TransactionType.Deposit:
      return 'account_balance'
    case TransactionType.Withdrawal:
      return 'account_balance'
    default:
      return null
  }
}

const activityTitle = (type: TransactionType): string => {
  switch (type) {
    // case TransactionType.Received:
    //   return <ReceivedIcon />
    case TransactionType.Outgoingpayment:
      return 'Sent'
    case TransactionType.Deposit:
      return 'Deposit'
    case TransactionType.Withdrawal:
      return 'Withdrawal'
    default:
      return ''
  }
}

const ActivityCard: FC<{ activity: Activity }> = ({ activity }) => {
  return (
    <Router
      to={route('/activity/transaction/:id', {
        id: activity.id
      })}
      className='col-span-full flex items-center justify-between rounded-xl bg-container py-2 px-3 sm:col-span-6 sm:col-start-2 lg:col-start-4'
    >
      <div className='flex items-center justify-between space-x-3'>
        <Icon>{activityIcon(activity.transactionType, activity.status)}</Icon>
        <div className='flex flex-col'>
          <span className='font-display text-base font-medium'>
            {activity.title}
          </span>
          <span className='font-sans text-xs font-normal'>
            {activity.description}
          </span>
        </div>
      </div>
      <span className='font-sans text-lg font-normal'>{activity.amount}</span>
    </Router>
  )
}
