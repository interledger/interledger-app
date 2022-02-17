import React, { FC } from 'react'
import { Container } from './Container'
import { LeavesDecor } from './Decor'
import { Link } from 'remix'
import { route } from 'routes-gen'

type ErrorProps = {
  statusCode?: number
  reason?: string
}

export const Error: FC<ErrorProps> = ({ reason, statusCode }) => {
  return (
    <div className='relative w-full overflow-hidden'>
      <Container className='overflow-x-hidden'>
        <main className='flex flex-grow flex-col items-start justify-center px-4 sm:p-8'>
          {statusCode && (
            <p className='font-display text-9xl font-medium text-primary'>
              {statusCode}
            </p>
          )}
          <div className='sm:mt-12'>
            <div>
              <h1 className='font-display text-4xl font-medium text-medium'>
                An error occurred
              </h1>
              <p className='mt-2 font-sans text-weak'>
                {/*TODO put in support email here.*/}
                {reason ||
                  'Please try again, and contact support if the problem persists.'}
              </p>
            </div>
            <div className='mt-10'>
              <Link to={route('/home')}>
                <span className='text-primary'>Go back home</span>
              </Link>
            </div>
          </div>
        </main>
        <LeavesDecor />
      </Container>
    </div>
  )
}
