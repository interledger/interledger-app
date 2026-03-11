import { Link } from '@remix-run/react'
import type { FC } from 'react'

type ErrorProps = {
  status?: number
  statusText?: string
  data: {
    title?: string
    action?: {
      route: string
      text: string
    }
  }
}

export const QuickPayError: FC<ErrorProps> = ({ status, statusText, data }) => {
  return (
    <main className='w-full overflow-hidden'>
      <section className='relative mx-auto grid w-full grid-cols-4 content-start gap-4 gap-y-2 overflow-x-visible px-8 sm:max-w-lg sm:grid-cols-8 sm:px-0 lg:max-w-3xl lg:grid-cols-12 xl:max-w-[59rem]'>
        <div className='relative col-span-full -mx-8 h-44 sm:col-span-6 sm:col-start-2 lg:col-start-4 lg:mx-0'>
          <div className='absolute right-[10.5rem] top-0 h-14 w-14 rounded-bl-full bg-slate-700' />
          <div className='absolute right-28 top-0 h-14 w-14 rounded-tr-full bg-slate-400' />
          <div className='absolute right-14 top-0 h-14 w-14 rounded-full bg-slate-300' />
          <div className='absolute right-0 top-0 h-14 w-14 bg-slate-100' />
          <div className='absolute right-56 top-14 h-14 w-14 rounded-full bg-slate-300' />
          <div className='absolute right-28 top-14 h-14 w-14 rounded-b-full bg-slate-100' />
          <div className='absolute right-14 top-14 h-14 w-14 rounded-full bg-rose-500' />
          <div className='absolute right-0 top-14 h-14 w-14 rounded-tl-full bg-slate-500' />
          <div className='absolute right-0 top-28 h-14 w-14 rounded-full bg-slate-300' />
          <div className='absolute right-14 top-[10.5rem] h-14 w-14 rounded-full bg-slate-600' />
          <div className='absolute right-0 top-[10.5rem] h-14 w-14 rounded-b-full bg-slate-100' />
          {/* Desktop only */}
          <div className='absolute -right-14 top-0 hidden h-14 w-14 rounded-full bg-slate-600 lg:block' />
          <div className='absolute -right-28 top-0 hidden h-14 w-14 rounded-t-full bg-slate-100 lg:block' />
          <div className='absolute -right-14 top-14 hidden h-14 w-14 rounded-full bg-slate-300 lg:block' />
          <div className='absolute -right-14 top-28 hidden h-14 w-14 rounded-full bg-slate-200 lg:block' />
          <div className='absolute -right-28 top-28 hidden h-14 w-14 rounded-br-full bg-slate-300 lg:block' />
          <div className='absolute -right-14 top-[10.5rem] hidden h-14 w-14 rounded-full bg-slate-300 lg:block' />
          <div className='absolute -right-28 top-[10.5rem] hidden h-14 w-14 bg-slate-100 lg:block' />
        </div>
        <div className='col-span-full flex flex-grow flex-col items-start justify-center sm:col-span-6 sm:col-start-2 lg:col-start-4'>
          {status && (
            <p className='text-9xl font-medium text-medium'>{status}</p>
          )}
          {!status && <div className='h-32' />}
          <div className='sm:mt-12'>
            <div>
              <h1 className='text-4xl font-medium text-medium'>
                {statusText || 'An error occurred'}
              </h1>
              {data?.title && <p className='mt-2 text-weak'>{data?.title}</p>}
            </div>
            <div className='mt-10'>
              <Link to={data?.action?.route || '/quick-pay'}>
                <span className='text-primary'>
                  {data?.action?.text || 'Start over'}
                </span>
              </Link>
            </div>
          </div>
        </div>
      </section>
    </main>
  )
}
