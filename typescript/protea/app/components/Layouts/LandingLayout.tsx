import { Outlet } from '@remix-run/react'

export function LandingLayout() {
  return (
    <div className='relative w-full overflow-hidden'>
      <Outlet />
    </div>
  )
}
