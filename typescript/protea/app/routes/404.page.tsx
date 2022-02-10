import { NextPage } from 'next'
import React from 'react'
import { Container, LeavesDecor, Router, Routes } from '../components'

const ErrorPage: NextPage = () => {
  return (
    <div className='relative w-full overflow-hidden'>
      <Container className='overflow-x-hidden'>
        <main className='flex flex-grow flex-col items-start justify-center px-4 sm:p-8'>
          <p className='font-display text-9xl font-medium text-primary'>404</p>
          <div className='sm:mt-12'>
            <div>
              <h1 className='font-display text-4xl font-medium text-medium'>
                Page not found
              </h1>
              <p className='mt-2 font-sans text-weak'>
                Please check the URL in the address bar and try again.
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

export default ErrorPage
