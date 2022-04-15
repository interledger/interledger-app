import type { FC } from 'react'
import React from 'react'
import { Container } from './Container'
import { LeavesDecor } from './Decor'
import { Link } from '@remix-run/react'

type ErrorProps = {
  status?: number
  data: {
    title?: string
    body?: string
    action?: {
      route: string
      text: string
    }
  }
}

export const Error: FC<ErrorProps> = ({ status, data }) => {
  return (
    <div className='relative w-full overflow-hidden'>
      <Container className='overflow-x-hidden'>
        <main className='flex flex-grow flex-col items-start justify-center px-4 sm:p-8'>
          {status && (
            <p className='font-display text-9xl font-medium text-primary'>
              {status}
            </p>
          )}
          <div className='sm:mt-12'>
            <div>
              <h1 className='font-display text-4xl font-medium text-medium'>
                {data.title || 'An error occurred'}
              </h1>
              <p className='mt-2 font-sans text-weak'>
                {/*TODO put in support email here.*/}
                {data.body ||
                  'Please try again, and contact support if the problem persists.'}
              </p>
            </div>
            {data.action && (
              <div className='mt-10'>
                <Link to={data.action.route || '/home'}>
                  <span className='text-primary'>
                    {data.action.text || 'Go back home'}
                  </span>
                </Link>
              </div>
            )}
          </div>
        </main>
        <LeavesDecor />
      </Container>
    </div>
  )
}
