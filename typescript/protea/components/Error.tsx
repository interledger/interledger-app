import React, { FC } from 'react'
import { Container } from './Container'
import { Router, Routes } from './Routes'
import { LeavesDecor } from './Decor'

type ErrorProps = {
  statusCode?: number
  reason?: string
}

export const Error: FC<ErrorProps> = ({ reason, statusCode }) => {
  return (
    <div className='relative overflow-hidden w-full'>
      <Container className='overflow-x-hidden'>
        <main className='flex flex-grow flex-col px-4 sm:p-8 justify-center items-start'>
          {statusCode && (
            <p className='text-9xl font-medium font-display text-primary'>
              {statusCode}
            </p>
          )}
          <div className='sm:mt-12'>
            <div>
              <h1 className='text-4xl font-medium font-display text-medium'>
                An error occurred
              </h1>
              <p className='mt-2 text-weak font-sans'>
                {/*TODO put in support email here.*/}
                {reason ||
                  'Please try again, and contact support if the problem persists.'}
              </p>
            </div>
            <div className='mt-10'>
              <Router href={Routes.home}>
                <span className='text-primary'>Go back home</span>
              </Router>
            </div>
          </div>
        </main>
        <LeavesDecor />
      </Container>
    </div>
  )
}
