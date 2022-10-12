import { AnchorRouter, ButtonRouter, Shape } from '~/components'
import { route } from 'routes-gen'

export default function Page() {
  return (
    <main className='w-full overflow-hidden'>
      <section className='relative mx-auto grid w-full grid-cols-4 content-start gap-4 gap-y-2 overflow-x-visible px-4 sm:max-w-lg sm:grid-cols-8 sm:px-0 lg:max-w-3xl lg:grid-cols-12 xl:max-w-[59rem]'>
        <div className='relative col-span-full h-20'>
          <div className='absolute top-20 -left-40 hidden h-20 w-20 rounded-full bg-slate-50 lg:block' />
          <div className='absolute top-0 -left-20 hidden h-20 w-20 rounded-tl-full bg-slate-100 lg:block' />
          <div className='absolute -right-20 top-40 hidden h-20 w-20 rounded-br-full bg-slate-100 lg:block' />
          <div className='absolute right-0 top-0 h-20 w-20 rounded-bl-full bg-yellow-100 lg:-right-40' />
        </div>
        <div className='col-span-full'>
          <h1 className='flex justify-center font-display text-2xl font-medium lg:text-4xl'>
            What is a payment pointer?
          </h1>
        </div>
        <div className='col-span-full lg:col-span-10 lg:col-start-2 lg:mt-7'>
          <span className='flex justify-center text-center lg:text-2xl'>
            A payment pointer is an email address for your digital wallet.
          </span>
        </div>
        <div className='relative col-span-full h-28'>
          <div className='absolute right-16 bottom-0 h-20 w-20 rounded-tl-full bg-slate-50 lg:-right-20' />
          <div className='absolute -right-4 bottom-0 h-20 w-20 rounded-tr-full bg-slate-100 lg:-right-40' />
        </div>
      </section>
      <div className='w-full bg-slate-50'>
        <section className='mx-auto grid w-full grid-cols-4 content-start gap-4 gap-y-2 overflow-x-visible p-12 sm:max-w-lg sm:grid-cols-8 sm:px-0 sm:py-20 lg:max-w-3xl lg:grid-cols-12 xl:max-w-[59rem]'>
          <div className='col-span-full flex flex-col items-center lg:col-span-2 lg:col-start-2 lg:items-start'>
            <div className='flex'>
              <Shape
                radius='rounded-tl-full'
                color='bg-slate-500'
                width='w-12'
              />
              <Shape radius='rounded-full' color='bg-rose-500' width='w-12' />
            </div>
            <div className='flex'>
              <Shape
                radius='rounded-bl-full'
                color='bg-lime-500'
                width='w-12'
              />
              <Shape
                radius='rounded-tl-full'
                color='bg-lime-300'
                width='w-12'
              />
            </div>
          </div>
          <div className='col-span-full mt-6 lg:col-span-8 lg:col-start-4 lg:mt-0'>
            <p className='text-center text-xl lg:text-left lg:text-2xl'>
              Digital wallets are revolutionising the way we pay. They are easy
              to set up, secure and make payments a breeze.
            </p>
          </div>
        </section>
      </div>
      <section className='relative mx-auto  grid w-full grid-cols-4 content-start gap-4 gap-y-12 overflow-x-visible px-4 pb-12 sm:max-w-lg sm:grid-cols-8 sm:px-0 lg:max-w-3xl lg:grid-cols-12 lg:pb-14 xl:max-w-[59rem]'>
        <div className='relative col-span-full h-40 lg:h-20'>
          <div className='absolute top-0 -left-4 h-20 w-20 rounded-tr-full bg-slate-50 lg:-left-40' />
          <div className='absolute top-0 left-16 h-20 w-20 rounded-bl-full bg-slate-50 lg:-left-20' />
          <div className='absolute top-20 -left-4 h-20 w-20 rounded-br-full bg-slate-50 lg:-left-40' />
        </div>
        <div className='col-span-full lg:col-span-10 lg:col-start-2'>
          <p className='mb-6 text-xl'>
            Most digital wallets exist in a walled garden. You are unable to
            make a payment from an Apple Pay wallet to a Google Pay merchant,
            for instance, or send money from a PayPal wallet to someone unless
            they have a PayPal wallet too.
          </p>
        </div>
        <div className='col-span-full flex flex-col lg:col-span-4 lg:col-start-2'>
          <div className='flex flex-col'>
            <div className='flex'>
              <Shape
                radius='rounded-tr-full'
                color='bg-slate-500'
                width='w-10'
              />
              <Shape
                radius='rounded-tl-full'
                color='bg-orange-600'
                width='w-10'
              />
            </div>
            <div className='flex'>
              <Shape
                radius='rounded-bl-full'
                color='bg-orange-400'
                width='w-10'
              />
              <Shape
                radius='rounded-br-full'
                color='bg-orange-900'
                width='w-10'
              />
            </div>
            <div className='flex'>
              <Shape
                radius='rounded-bl-full'
                color='bg-orange-300'
                width='w-10'
              />
              <Shape
                radius='rounded-br-full'
                color='bg-slate-400'
                width='w-10'
              />
            </div>
          </div>
          <h3 className='mt-8 text-xl font-medium'>Digital wallets</h3>
          <p className='mt-4 text-medium'>
            With a payment pointer, digital wallets become part of an open
            ecosystem as ubiquitous as the web. A wallet with a payment pointer
            is accessible and interoperable with almost any other digital
            wallet, making them easier to use than email.
          </p>
        </div>
        <div className='col-span-full flex flex-col lg:col-span-4 lg:col-start-8'>
          <div className='flex flex-col'>
            <div className='flex'>
              <Shape radius='rounded-full' color='bg-rose-600' width='w-10' />
              <Shape
                radius='rounded-br-full'
                color='bg-rose-300'
                width='w-10'
              />
            </div>
            <div className='flex'>
              <Shape
                radius='rounded-tl-full'
                color='bg-slate-900'
                width='w-10'
              />
              <Shape
                radius='rounded-br-full'
                color='bg-slate-500'
                width='w-10'
              />
            </div>
            <div className='flex'>
              <Shape
                radius='rounded-tl-full'
                color='bg-slate-400'
                width='w-10'
              />
              <Shape
                radius='rounded-br-full'
                color='bg-rose-400'
                width='w-10'
              />
            </div>
          </div>
          <h3 className='mt-8 text-xl font-medium'>Send and receive</h3>
          <p className='mt-4 text-medium'>
            Payment pointers are URLs that connect securely to a digital wallet.
            You can use a payment pointer to send and receive payments to and
            from your wallet.
          </p>
        </div>
        <div className='col-span-full flex flex-col lg:col-span-4 lg:col-start-2'>
          <div className='flex flex-col'>
            <div className='flex'>
              <Shape
                radius='rounded-tl-full'
                color='bg-indigo-600'
                width='w-10'
              />
              <Shape
                radius='rounded-tr-full'
                color='bg-slate-500'
                width='w-10'
              />
            </div>
            <div className='flex'>
              <Shape
                radius='rounded-bl-full'
                color='bg-indigo-400'
                width='w-10'
              />
              <Shape
                radius='rounded-br-full'
                color='bg-indigo-900'
                width='w-10'
              />
            </div>
            <div className='flex'>
              <Shape
                radius='rounded-br-full'
                color='bg-slate-400'
                width='w-10'
              />
              <Shape
                radius='rounded-bl-full'
                color='bg-indigo-300'
                width='w-10'
              />
            </div>
          </div>
          <h3 className='mt-8 text-xl font-medium'>Control access</h3>
          <p className='mt-4 text-medium'>
            Share a payment pointer with others and control the access they have
            to deliver additional features and services with your wallet.
          </p>
        </div>
        <div className='col-span-full flex flex-col lg:col-span-4 lg:col-start-8'>
          <div className='flex flex-col'>
            <div className='flex'>
              <Shape
                radius='rounded-tr-full'
                color='bg-lime-300'
                width='w-10'
              />
              <Shape
                radius='rounded-tl-full'
                color='bg-slate-400'
                width='w-10'
              />
            </div>
            <div className='flex'>
              <Shape
                radius='rounded-br-full'
                color='bg-lime-400'
                width='w-10'
              />
              <Shape
                radius='rounded-bl-full'
                color='bg-slate-900'
                width='w-10'
              />
            </div>
            <div className='flex'>
              <Shape
                radius='rounded-br-full'
                color='bg-slate-500'
                width='w-10'
              />
              <Shape radius='rounded-full' color='bg-lime-600' width='w-10' />
            </div>
          </div>
          <h3 className='mt-8 text-xl font-medium'>Registration is free</h3>
          <p className='mt-4 text-medium'>
            Based on open source and open API standards, payment pointers are
            safe to share, free to register and easy to remember.
          </p>
        </div>
      </section>
      <div className='w-full bg-slate-50'>
        <section className='mx-auto grid w-full grid-cols-4 content-start gap-4 gap-y-2 overflow-x-visible px-4 sm:max-w-lg sm:grid-cols-8 sm:px-0 lg:max-w-3xl lg:grid-cols-12 xl:max-w-[59rem]'>
          <div className='relative col-span-full h-20'>
            <div className='absolute bottom-0 -right-20 hidden h-20 w-20 rounded-full bg-white lg:block' />
            <div className='absolute bottom-0 -left-4 h-20 w-20 rounded-full bg-white lg:-left-40' />
            <div className='absolute bottom-0 left-16 h-20 w-20 rounded-br-full bg-white lg:-left-20' />
          </div>
          <div className='col-span-full mt-12 flex justify-center lg:mt-6'>
            <ButtonRouter shrink to={route('/waitlist')}>
              Join the waitlist
            </ButtonRouter>
          </div>
          <div className='col-span-full mt-4 flex items-center justify-center text-center'>
            <span className='text-medium'>
              For a more detailed technical outline of how payment pointers
              work,
              <br /> please visit{' '}
              <AnchorRouter className='text-primary' to='paymentpointers.org'>
                paymentpointers.org
              </AnchorRouter>
            </span>
          </div>
          <div className='relative col-span-full h-16 lg:h-20'>
            <div className='absolute bottom-0 -right-20 hidden h-20 w-20 rounded-tl-full bg-white lg:block' />
            <div className='absolute bottom-0 -right-40 hidden h-20 w-20 rounded-tr-full bg-white lg:block' />
          </div>
        </section>
      </div>
      <section className='relative mx-auto grid w-full grid-cols-4 content-start gap-4 gap-y-12 overflow-x-visible px-4 text-center sm:max-w-lg sm:grid-cols-8 sm:px-0 lg:max-w-3xl lg:grid-cols-12 xl:max-w-[59rem]'>
        <div className='relative col-span-full h-20'>
          <div className='absolute top-0 -left-4 h-20 w-20 rounded-tr-full bg-slate-50 lg:-left-40' />
          <div className='absolute top-0 left-16 h-20 w-20 rounded-bl-full bg-slate-50 lg:-left-20' />
          <div className='absolute top-0 left-36 h-20 w-20 rounded-br-full bg-slate-50 lg:left-0' />
        </div>
        <div className='col-span-full lg:col-span-10 lg:col-start-2'>
          <h1 className='mb-6 font-display text-4xl'>More information</h1>
        </div>
        <AnchorRouter
          to='https://docs.openpayments.guide'
          className='col-span-full flex flex-col items-center lg:col-span-4 lg:col-start-2'
        >
          <svg
            width='62'
            height='63'
            viewBox='0 0 62 63'
            fill='none'
            xmlns='http://www.w3.org/2000/svg'
          >
            <path
              fillRule='evenodd'
              clipRule='evenodd'
              d='M61.8252 28.2854C60.6497 17.233 53.666 7.91522 43.9848 3.44336H34.9478V10.8452C43.6149 12.5204 50.3883 19.5085 51.744 28.2854H61.8252Z'
              fill='#F59297'
            />
            <path
              fillRule='evenodd'
              clipRule='evenodd'
              d='M16.7368 59.133V47.0982C15.7957 46.2401 14.9338 45.299 14.1616 44.2861H6.73682V50.8989C9.45592 54.3144 12.863 57.1198 16.7368 59.133Z'
              fill='#8FD1C1'
            />
            <path
              fillRule='evenodd'
              clipRule='evenodd'
              d='M61.8464 28.4971H51.775C51.9178 29.487 51.9917 30.499 51.9917 31.5285C51.9917 34.6407 51.3168 37.5954 50.1053 40.2541V44.2866H59.2945C61.0331 40.4148 62 36.1216 62 31.6024C62 30.5545 61.9479 29.5184 61.8464 28.4971Z'
              fill='#FABD84'
            />
            <path
              fillRule='evenodd'
              clipRule='evenodd'
              d='M56.4739 23.9707C54.0032 23.9707 52.0002 25.9737 52.0002 28.4443V29.076C52.0002 31.5466 54.0032 33.5496 56.4739 33.5496C58.9448 33.5496 60.9475 31.5466 60.9475 29.076V28.4443C60.9475 25.9737 58.9448 23.9707 56.4739 23.9707Z'
              fill='#FCC9B3'
            />
            <path
              fillRule='evenodd'
              clipRule='evenodd'
              d='M60.5759 40.917C56.6208 53.4861 44.8749 62.6014 31 62.6014C29.7352 62.6014 28.4881 62.5257 27.2631 62.3784V52.2757C28.4726 52.4873 29.6982 52.5935 30.9261 52.593C39.1863 52.593 46.3363 47.8387 49.7883 40.917H60.5759Z'
              fill='#7FC78C'
            />
            <path
              fillRule='evenodd'
              clipRule='evenodd'
              d='M11.0184 39.6543C11.3018 39.914 11.5767 40.1828 11.8427 40.4603C11.7179 40.1942 11.5988 39.9254 11.4853 39.6543H11.0184ZM25.3686 52.2332C25.3686 55.5732 24.7248 58.7665 23.5518 61.7011C20.7069 60.9992 17.9767 59.895 15.4437 58.4219C16.2571 56.3812 16.6603 54.2002 16.6302 52.0036C16.6001 49.807 16.1373 47.6379 15.2684 45.6202C17.9301 48.5757 21.4167 50.7738 25.3658 51.8516C25.3675 51.9787 25.3686 52.1057 25.3686 52.2332Z'
              fill='#459789'
            />
            <path
              fillRule='evenodd'
              clipRule='evenodd'
              d='M2.13028 42.9163H9.1578V24.916H0.722497C0.240938 27.1115 -0.00130584 29.3527 5.29372e-06 31.6004C5.29372e-06 35.5937 0.755142 39.4106 2.13028 42.9163Z'
              fill='#9EC7D0'
            />
            <path
              fillRule='evenodd'
              clipRule='evenodd'
              d='M1.8462 42.1636C5.95972 39.1575 8.6318 34.2967 8.6318 28.8115C8.6318 24.178 6.72472 19.9898 3.65306 16.989C5.51409 13.5165 8.0173 10.4287 11.0298 7.88965C13.481 10.1075 15.525 12.7374 17.0691 15.6602C12.6512 19.5219 9.86041 25.1986 9.86041 31.5272C9.86041 36.7964 11.7948 41.6137 14.9925 45.3076C13.0379 48.0146 10.6206 50.3551 7.85211 52.2215C5.22869 49.2813 3.1902 45.8677 1.84595 42.1636H1.8462Z'
              fill='#51797D'
            />
            <path
              fillRule='evenodd'
              clipRule='evenodd'
              d='M32.0003 0.617505C31.6683 0.606624 31.3348 0.601562 31.0002 0.601562C19.4935 0.601562 9.45076 6.87117 4.10205 16.1806H16.4963C20.2661 12.6347 25.3425 10.4621 30.9263 10.4621C31.2864 10.4621 31.6445 10.471 32.0003 10.4889V0.617505Z'
              fill='#978AA4'
            />
            <path
              fillRule='evenodd'
              clipRule='evenodd'
              d='M28.6164 62.5107C28.6493 58.9157 29.5871 55.5366 31.2133 52.5907C28.9194 52.6241 26.6354 52.2825 24.4517 51.5792C23.0883 54.655 22.3045 57.9561 22.1398 61.3165C24.2499 61.9442 26.4211 62.3445 28.6164 62.5107ZM61.3636 37.8812C58.3135 36.4309 55.0209 35.559 51.6528 35.3096C51.2438 37.566 50.4667 39.7399 49.3525 41.7442C49.4279 41.7434 49.5033 41.7429 49.579 41.7429C53.1085 41.7429 56.4342 42.6152 59.3525 44.1559C60.2455 42.1423 60.9197 40.0387 61.3636 37.8812Z'
              fill='#6D995C'
            />
            <path
              fillRule='evenodd'
              clipRule='evenodd'
              d='M61.5883 26.5365C60.7514 24.4755 58.7297 23.0225 56.3684 23.0225C54.1452 23.0225 52.223 24.3105 51.3074 26.1814C51.8189 28.1402 52.0459 30.1624 51.9815 32.1858C53.0138 33.4663 54.5954 34.2857 56.3684 34.2857C59.1801 34.2857 61.5106 32.2253 61.9319 29.532C61.8654 28.5192 61.75 27.5201 61.5883 26.5365ZM59.8948 28.6541C59.8948 30.6017 58.3159 32.1805 56.3684 32.1805C54.4208 32.1805 52.8422 30.6017 52.8422 28.6541C52.8422 26.7065 54.4208 25.1279 56.3684 25.1279C58.3159 25.1279 59.8948 26.7065 59.8948 28.6541Z'
              fill='#F47F5F'
            />
            <path
              fillRule='evenodd'
              clipRule='evenodd'
              d='M58.7413 17.7496C58.53 17.7555 58.3181 17.7585 58.1053 17.7585C47.7576 17.7585 39.0624 10.6821 36.5956 1.10449C46.3162 2.87618 54.4549 9.18172 58.7413 17.7496Z'
              fill='#CE6564'
            />
            <path
              fillRule='evenodd'
              clipRule='evenodd'
              d='M4.32263 15.802C14.9206 14.1237 24.2234 8.55429 30.7218 0.602828C30.8144 0.602069 30.907 0.601562 30.9999 0.601562C35.2827 0.601562 39.3626 1.47007 43.073 3.04057C41.1692 5.97097 38.9877 8.71128 36.5587 11.2236C34.7241 10.7166 32.8293 10.4604 30.926 10.4619C21.1323 10.4619 12.8994 17.1455 10.5399 26.2003C7.16366 27.0278 3.71565 27.5284 0.24353 27.6951C0.771068 23.4924 2.15978 19.4437 4.32289 15.802H4.32263Z'
              fill='#845578'
            />
          </svg>
          <h3 className='mt-8 text-xl font-medium'>Open payments</h3>
          <p className='mt-4 text-medium'>
            An API standard that can be implemented by any account servicing
            entity such as digital wallets, banks, crypto-wallets and mobile
            money providers.
          </p>
        </AnchorRouter>
        <AnchorRouter
          to='https://paymentpointers.org'
          className='col-span-full flex flex-col items-center lg:col-span-4 lg:col-start-8'
        >
          <svg
            width='62'
            height='63'
            viewBox='0 0 62 63'
            fill='none'
            xmlns='http://www.w3.org/2000/svg'
          >
            <path
              d='M0 16.1016C0 7.54115 6.93959 0.601562 15.5 0.601562L31 0.601562V31.6016H15.5C6.93959 31.6016 0 24.662 0 16.1016Z'
              fill='#84CC16'
            />
            <path
              d='M0 31.6016L31 31.6016L31 47.1016C31 55.662 24.0604 62.6016 15.5 62.6016C6.93959 62.6016 0 55.662 0 47.1016L0 31.6016Z'
              fill='#64748B'
            />
            <path
              d='M31 31.6016C31 14.4807 44.8792 0.601562 62 0.601562C62 17.7224 48.1208 31.6016 31 31.6016Z'
              fill='#CBD5E1'
            />
            <circle cx='46.5' cy='47.1016' r='15.5' fill='#F43F5E' />
          </svg>
          <h3 className='mt-8 text-xl font-medium'>Payment Pointers</h3>
          <p className='mt-4 text-medium'>
            A standard for URLs used to access Open payments APIs.
          </p>
        </AnchorRouter>
        <AnchorRouter
          to='https://interledger.org'
          className='col-span-full flex flex-col items-center lg:col-span-4 lg:col-start-2'
        >
          <svg
            width='62'
            height='63'
            viewBox='0 0 62 63'
            fill='none'
            xmlns='http://www.w3.org/2000/svg'
          >
            <path
              d='M22.405 49.5187L22.3889 49.4983C18.7015 47.7223 15.6486 44.8566 13.6432 41.2887C10.3156 38.9433 7.82075 35.5998 6.51968 31.7422C5.42236 35.1398 5.2972 38.7766 6.15834 42.2417C7.01949 45.7067 8.8326 48.8618 11.3928 51.3505C13.953 53.8391 17.1582 55.5621 20.6463 56.3247C24.1343 57.0873 27.7661 56.8592 31.1313 55.666C27.6993 54.4843 24.6732 52.3526 22.405 49.5187Z'
              fill='#8C528C'
            />
            <path
              d='M39.7941 49.4204L39.807 49.4043L39.7812 49.4172C37.0824 50.742 34.1188 51.4389 31.1124 51.4559C28.1061 51.4729 25.1348 50.8094 22.4212 49.5152L22.3889 49.499L22.4051 49.5195C24.6733 52.3533 27.6993 54.485 31.1313 55.6667C34.5523 54.4472 37.5565 52.281 39.7941 49.4204Z'
              fill='#8775BD'
            />
            <path
              d='M22.4211 49.5148C25.1347 50.809 28.106 51.4724 31.1124 51.4555C34.1187 51.4385 37.0823 50.7416 39.7811 49.4168C36.7502 48.732 33.8264 47.638 31.0904 46.165C28.3663 47.6704 25.4497 48.7974 22.4211 49.5148Z'
              fill='#ADC8E6'
            />
            <path
              d='M48.4644 41.0907C46.4989 44.6823 43.4775 47.5834 39.8091 49.4015L39.7962 49.4176C37.5581 52.277 34.554 54.442 31.1334 55.6607C34.5116 56.8166 38.1458 57.0048 41.6253 56.2039C45.1048 55.403 48.291 53.645 50.8239 51.1285C53.3567 48.6119 55.1354 45.4372 55.9588 41.9629C56.7822 38.4887 56.6176 34.8534 55.4835 31.4678C54.2243 35.3391 51.766 38.7093 48.4644 41.0907Z'
              fill='#56C1BF'
            />
            <path
              d='M48.4235 21.9678C50.0336 24.8953 50.8813 28.1809 50.8885 31.522C50.8956 34.863 50.062 38.1522 48.4644 41.0866C51.766 38.7052 54.2244 35.335 55.4835 31.4637C54.19 27.6333 51.7189 24.3097 48.4235 21.9678Z'
              fill='#2EA690'
            />
            <path
              d='M48.4644 41.0886C50.062 38.1542 50.8957 34.865 50.8885 31.5239C50.8814 28.1828 50.0337 24.8972 48.4235 21.9697C47.8337 25.3037 46.7535 28.532 45.218 31.5496C46.7639 34.5521 47.8578 37.7664 48.4644 41.0886Z'
              fill='#81DEC8'
            />
            <path
              d='M45.2181 31.5527C44.7157 32.5366 44.1707 33.492 43.583 34.4192C41.0736 38.3652 37.7928 41.7635 33.9375 44.4102C33.019 45.0395 32.07 45.6251 31.0905 46.1669C33.8265 47.6399 36.7502 48.7338 39.7812 49.4186L39.8071 49.4057C43.4755 47.5876 46.4968 44.6865 48.4623 41.0949C47.8567 37.7718 46.7635 34.5563 45.2181 31.5527Z'
              fill='#A2EBDA'
            />
            <path
              d='M13.4743 22.1572L13.4883 22.1314C15.468 18.4696 18.5422 15.5177 22.2813 13.6882H22.2899C24.5198 10.8764 27.4939 8.74659 30.8741 7.54093C27.4953 6.38293 23.86 6.19324 20.3791 6.99331C16.8982 7.79338 13.7106 9.5513 11.1765 12.0683C8.64248 14.5853 6.86312 17.7611 6.0396 21.2365C5.21609 24.712 5.38128 28.3485 6.51648 31.7349C7.76783 27.8886 10.2031 24.5363 13.4743 22.1572Z'
              fill='#FF7A7F'
            />
            <path
              d='M13.64 41.2812C12.0099 38.3639 11.1402 35.0833 11.111 31.7415C11.0819 28.3998 11.8942 25.1045 13.4732 22.1592C10.2017 24.5388 7.76633 27.8919 6.51538 31.739C7.81753 35.5952 10.3127 38.9371 13.64 41.2812Z'
              fill='#F2666D'
            />
            <path
              d='M13.4915 22.1458V22.1318L13.4775 22.1577C11.8986 25.103 11.0862 28.3983 11.1154 31.74C11.1445 35.0818 12.0142 38.3624 13.6444 41.2797C14.2144 37.9447 15.275 34.7124 16.7917 31.688C15.2266 28.6876 14.1145 25.472 13.4915 22.1458Z'
              fill='#FFB0BC'
            />
            <path
              d='M28.2197 44.4363C24.3291 41.8282 21.0067 38.4596 18.4525 34.5335C17.8562 33.6135 17.3026 32.6649 16.7917 31.6875C15.2735 34.7117 14.2114 37.9439 13.64 41.2792C15.6446 44.8512 18.6989 47.7205 22.3889 49.4985L22.4212 49.5147C25.4498 48.7973 28.3664 47.6703 31.0904 46.1649C30.1037 45.6346 29.1468 45.0584 28.2197 44.4363Z'
              fill='#FFDBDB'
            />
            <path
              d='M39.5185 13.5907H39.5282C43.2943 15.3844 46.4045 18.312 48.4224 21.9629C51.7186 24.306 54.1897 27.6311 55.4824 31.4631C56.5806 28.0651 56.7062 24.4277 55.8453 20.962C54.9843 17.4963 53.171 14.3406 50.6104 11.8515C48.0498 9.36246 44.8439 7.63933 41.3552 6.87691C37.8666 6.11448 34.2342 6.34317 30.8687 7.53713C34.2618 8.70536 37.2589 10.8028 39.5185 13.5907Z'
              fill='#68CB88'
            />
            <path
              d='M22.2812 13.6844C24.9639 12.3805 27.9049 11.6952 30.8876 11.679C33.8703 11.6628 36.8186 12.3161 39.5153 13.5907C37.2557 10.8028 34.2586 8.70534 30.8655 7.53711C27.4853 8.74276 24.5112 10.8726 22.2812 13.6844Z'
              fill='#FFA252'
            />
            <path
              d='M30.9204 16.955C33.6232 15.4524 36.517 14.3224 39.5229 13.5956H39.5282H39.5186C36.8219 12.321 33.8736 11.6677 30.8909 11.6839C27.9082 11.7001 24.9671 12.3854 22.2845 13.6893H22.2759C25.2914 14.3832 28.1993 15.4817 30.9204 16.955Z'
              fill='#FFC38F'
            />
            <path
              d='M30.9203 16.9556C36.0457 19.7364 40.3885 23.7628 43.5485 28.6635C44.1505 29.5963 44.707 30.5597 45.218 31.5536C46.7534 28.536 47.8336 25.3077 48.4234 21.9737C46.4055 18.3228 43.2954 15.3952 39.5292 13.6016H39.5239C36.5179 14.3265 33.6238 15.4548 30.9203 16.9556Z'
              fill='#CFEDAB'
            />
            <path
              d='M16.7917 31.6877C17.2897 30.6975 17.8326 29.7359 18.4203 28.803C20.9048 24.8539 24.1601 21.4463 27.9915 18.784C28.9345 18.1281 29.9107 17.5182 30.9204 16.9541C28.2009 15.4817 25.2949 14.3835 22.2813 13.6895C18.5427 15.5181 15.4685 18.4688 13.4883 22.1294V22.1434C14.1121 25.4705 15.2253 28.6869 16.7917 31.6877Z'
              fill='#F9D4B2'
            />
            <path
              d='M33.8696 18.7505C32.9171 18.1031 31.9328 17.5039 30.9203 16.9551C29.9114 17.5155 28.9351 18.1255 27.9915 18.7849C24.1601 21.447 20.9048 24.8542 18.4202 28.8029C17.8368 29.7358 17.2939 30.6973 16.7916 31.6876C17.3011 32.6635 17.8544 33.6111 18.4514 34.5304C21.0056 38.4565 24.328 41.8251 28.2186 44.4331C29.145 45.0539 30.1019 45.6301 31.0893 46.1618C32.0688 45.6215 33.0178 45.0359 33.9363 44.4052C37.7914 41.7591 41.0722 38.3615 43.5819 34.4163C44.171 33.4906 44.716 32.5351 45.2169 31.5498C44.7088 30.5596 44.1527 29.5973 43.5485 28.663C41.0213 24.7418 37.7294 21.3704 33.8696 18.7505Z'
              fill='white'
            />
          </svg>
          <h3 className='mt-8 text-xl font-medium'>Interledger</h3>
          <p className='mt-4 text-medium'>
            An open and inclusive payments network that puts humanity first.
          </p>
        </AnchorRouter>
        <AnchorRouter
          to='https://webmonetization.org'
          className='col-span-full flex flex-col items-center lg:col-span-4 lg:col-start-8'
        >
          <svg
            width='48'
            height='63'
            viewBox='0 0 48 63'
            fill='none'
            xmlns='http://www.w3.org/2000/svg'
          >
            <g clipPath='url(#clip0_440_5646)'>
              <path
                d='M16.4351 40.0936C16.4351 38.7807 16.0327 37.6995 15.228 36.85C14.4232 35.9851 13.082 35.2126 11.2042 34.533C9.32649 33.838 7.86604 33.1969 6.82284 32.6101C3.35053 30.6793 1.61437 27.8217 1.61437 24.0373C1.61437 21.4732 2.36695 19.3649 3.87212 17.7122C5.3773 16.0594 6.97188 15.1249 9.55003 14.816V9.71875H13.5737V14.816C16.1668 15.2021 18.1712 16.3452 19.5869 18.245C21.0028 20.1295 21.7107 22.5854 21.7107 25.6128H16.3009C16.3009 23.6667 15.8762 22.1375 15.0268 21.0254C14.1922 19.8979 13.0522 19.334 11.6066 19.334C10.1759 19.334 9.05824 19.7356 8.25351 20.5388C7.44876 21.3419 7.0464 22.4928 7.0464 23.991C7.0464 25.3349 7.44131 26.4161 8.23114 27.2347C9.03589 28.0379 10.392 28.8025 12.2996 29.5285C14.2071 30.2544 15.7048 30.9264 16.7927 31.5442C17.8806 32.162 18.7971 32.8726 19.5422 33.6757C20.2873 34.4635 20.8612 35.3749 21.2636 36.4097C21.6659 37.4446 21.867 38.6572 21.867 40.0473C21.867 42.6577 21.0921 44.7739 19.5422 46.3958C18.0073 48.0176 16.1445 48.9985 13.3875 49.2919V52.8521H9.55003V49.2919C6.5993 48.952 4.32665 47.8553 2.73206 46.0019C1.15239 44.1483 0.362549 41.6923 0.362549 38.6341H5.79457C5.79457 40.5802 6.264 42.0862 7.20286 43.1521C8.15663 44.2177 9.49788 44.7508 11.2266 44.7508C12.9255 44.7508 14.2146 44.3259 15.0938 43.4765C15.988 42.6269 16.4351 41.4992 16.4351 40.0936Z'
                fill='#6ADAAB'
              />
              <path
                d='M32.6195 27.6914H29.0354V38.4748H32.6195V27.6914Z'
                fill='#6ADAAB'
              />
              <path
                d='M39.7872 24.0967H36.2031V42.069H39.7872V24.0967Z'
                fill='#6ADAAB'
              />
              <path
                d='M46.9557 20.5029H43.3716V45.6642H46.9557V20.5029Z'
                fill='#6ADAAB'
              />
            </g>
            <defs>
              <clipPath id='clip0_440_5646'>
                <rect
                  width='47.275'
                  height='62'
                  fill='white'
                  transform='translate(0.362549 0.601562)'
                />
              </clipPath>
            </defs>
          </svg>
          <h3 className='mt-8 text-xl font-medium'>Web monetization</h3>
          <p className='mt-4 text-medium'>
            A JavaScript browser API that allows the creation of a payment
            stream from the user agent to the website
          </p>
        </AnchorRouter>
        <div className='relative col-span-full h-40 lg:h-20'>
          <div className='absolute bottom-0 -right-4 h-20 w-20 rounded-bl-full bg-slate-50 lg:-right-40' />
          <div className='absolute bottom-20 -right-4 h-20 w-20 rounded-bl-full bg-slate-50 lg:-right-40' />
          <div className='absolute bottom-0 -left-4 h-20 w-20 rounded-tr-full bg-rose-300 lg:-left-20' />
        </div>
      </section>
    </main>
  )
}
