import { Router } from '~/components'

export function MarketingPage() {
  return (
    <div className='flex grow flex-col items-center justify-center p-4 text-center'>
      <h1 className='text-5xl'>De-complicated</h1>
      <p className='mt-5 max-w-3xl text-3xl text-strong'>
        <Router to='/about' className='hover:text-rose-500'>
          We are Fynbos
        </Router>
        , a global{' '}
        <Router to='/legal' className='hover:text-rose-500'>
          fintech company
        </Router>{' '}
        based in{' '}
        <Router to='/contact' className='hover:text-rose-500'>
          Cape Town
        </Router>
        . We solve complex problems by building{' '}
        <Router to='/wealth' className='hover:text-rose-500'>
          simple products
        </Router>{' '}
        that are a joy to use.
      </p>
      <svg
        width='200'
        height='50'
        className='mt-16'
        viewBox='0 0 200 50'
        fill='none'
        xmlns='http://www.w3.org/2000/svg'
      >
        <path
          d='M0 0C13.8071 0 24.9999 11.1928 24.9999 24.9999L24.9999 50C11.1928 50 0 38.8072 0 25.0001L0 0Z'
          fill='#9333EA'
        />
        <path
          d='M25 24.9999C25 11.1928 36.1928 0 49.9999 0V25.0001C49.9999 38.8072 38.8071 50 25 50L25 24.9999Z'
          fill='#C084FC'
        />
        <path
          d='M75 0C88.8071 2.597e-06 100 11.1929 100 25L75 25L75 0Z'
          fill='#F97316'
        />
        <path
          d='M100 25C113.807 25 125 36.1929 125 50L100 50L100 25Z'
          fill='#FB923C'
        />
        <path
          d='M100 50C86.1929 50 75 38.8071 75 25L100 25L100 50Z'
          fill='#F97316'
        />
        <path
          d='M100 -8.05209e-07C113.807 -3.60504e-07 125 11.1929 125 25L100 25L100 -8.05209e-07Z'
          fill='#FB923C'
        />
        <circle
          cx='116.5'
          cy='14.5'
          r='2'
          transform='rotate(-90 116.5 14.5)'
          fill='#7C2D12'
        />
        <path
          d='M200 25C200 38.8071 188.807 49.9999 175 49.9999C175 36.1928 186.193 25 200 25Z'
          fill='#4ADE80'
        />
        <path
          d='M175 25C175 11.1929 186.193 0.000115909 200 0.000117719C200 13.8072 188.807 25 175 25Z'
          fill='#4ADE80'
        />
        <path
          d='M150 -1.09278e-06C163.807 -4.89254e-07 175 11.1928 175 24.9999C161.193 24.9999 150 13.8071 150 -1.09278e-06Z'
          fill='#16A34A'
        />
        <path
          d='M175 50C161.193 50 150 38.8072 150 25.0001C163.807 25.0001 175 36.193 175 50Z'
          fill='#16A34A'
        />
      </svg>
    </div>
  )
}
